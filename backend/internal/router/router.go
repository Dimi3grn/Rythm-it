// Fichier: backend/internal/router/router.go
package router

import (
	"net/http"
	"text/template"

	"rythmitbackend/configs"
	"rythmitbackend/internal/handlers"
	"rythmitbackend/internal/middleware"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Router instance globale
var Router *mux.Router

// Templates pour les pages HTML
var templates *template.Template

// Init initialise le router avec toutes les routes
func Init(cfg *configs.Config) *mux.Router {
	Router = mux.NewRouter()

	// Initialiser les templates
	if err := handlers.InitTemplates(); err != nil {
		panic("Erreur chargement templates: " + err.Error())
	}

	// Middleware global
	Router.Use(middleware.Logger)
	Router.Use(middleware.Recovery)

	// Servir les fichiers statiques (CSS, JS, images)
	Router.PathPrefix("/styles/").Handler(http.StripPrefix("/styles/",
		http.FileServer(http.Dir("../frontend/styles/"))))

	// Servir les images uploadées
	Router.PathPrefix("/uploads/").Handler(http.StripPrefix("/uploads/",
		http.FileServer(http.Dir("uploads/"))))

	// Routes pour les pages HTML
	setupPageRoutes()

	// Routes API publiques
	setupAPIRoutes()

	// Configuration CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://localhost:8085"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	})

	// Appliquer CORS comme middleware
	Router.Use(c.Handler)

	return Router
}

// setupPageRoutes configure les routes pour les pages HTML
func setupPageRoutes() {
	// Page d'accueil - index.html
	Router.HandleFunc("/", handlers.IndexHandler).Methods("GET")

	// Actions sur les posts
	Router.HandleFunc("/post", handlers.PostHandler).Methods("POST")
	Router.HandleFunc("/new-post", handlers.NewPostHandler).Methods("POST")

	// Autres pages
	Router.HandleFunc("/discover", handlers.DiscoverHandler).Methods("GET")
	Router.HandleFunc("/friends", handlers.FriendsHandler).Methods("GET")
	Router.HandleFunc("/messages", handlers.MessagesHandler).Methods("GET")
	Router.HandleFunc("/profile", handlers.ProfileHandler).Methods("GET", "POST")
	Router.HandleFunc("/settings", handlers.SettingsHandler).Methods("GET")
	Router.HandleFunc("/hub", handlers.HubHandler).Methods("GET")

	// Page thread individuel
	Router.HandleFunc("/thread/{id:[0-9]+}", handlers.ThreadHandler).Methods("GET", "POST")
	Router.HandleFunc("/thread/{id:[0-9]+}/delete", handlers.DeleteThreadHandler).Methods("POST")
	Router.HandleFunc("/thread/{id:[0-9]+}/edit", handlers.EditThreadHandler).Methods("GET", "POST")

	// Pages d'authentification
	Router.HandleFunc("/signin", handlers.SigninHandler).Methods("GET", "POST")
	Router.HandleFunc("/login", handlers.SigninHandler).Methods("GET", "POST") // Alias pour /signin
	Router.HandleFunc("/signup", handlers.SignupHandler).Methods("GET", "POST")
	Router.HandleFunc("/logout", handlers.LogoutHandler).Methods("GET", "POST")

	// Upload d'images
	Router.HandleFunc("/upload/image", handlers.UploadImageHandler).Methods("POST")

	// Actions de profil simples (sans JavaScript)
	Router.HandleFunc("/profile/action", handlers.SimpleProfileUpdateHandler).Methods("POST")

	// WebSocket pour notifications temps réel
	Router.HandleFunc("/ws", handlers.WebSocketHandler).Methods("GET")
}

// setupAPIRoutes configure les routes API
func setupAPIRoutes() {
	// Routes API avec préfixe /api
	api := Router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONMiddleware)

	// Routes publiques
	public := api.PathPrefix("/public").Subrouter()

	// Tags disponibles (pour autocomplete)
	public.HandleFunc("/tags", handlers.TagsAPIHandler).Methods("GET")

	// Threads publics avec pagination
	public.HandleFunc("/threads", handlers.ThreadsAPIHandler).Methods("GET")

	// Recherche publique
	public.HandleFunc("/search", handlers.SearchAPIHandler).Methods("GET")

	// Recherche de threads spécifique
	public.HandleFunc("/threads/search", handlers.ThreadSearchAPIHandler).Methods("GET")

	// Routes avec authentification optionnelle (sans préfixe v1)
	mixed := api.NewRoute().Subrouter()
	mixed.Use(middleware.OptionalAuthMiddleware)

	// Likes sur threads et messages
	mixed.HandleFunc("/threads/{id:[0-9]+}/like", handlers.ToggleLikeHandler).Methods("POST")
	mixed.HandleFunc("/messages/{id:[0-9]+}/like", handlers.MessageLikeHandler).Methods("POST")

	// Messages dans les threads
	mixed.HandleFunc("/threads/{id:[0-9]+}/messages", handlers.ThreadMessagesHandler).Methods("GET", "POST")
	mixed.HandleFunc("/messages/{id:[0-9]+}/vote", handlers.MessageVoteHandler).Methods("POST")

	// Profil utilisateur
	mixed.HandleFunc("/profile", handlers.ProfileAPIHandler).Methods("GET")

	// Notifications
	mixed.HandleFunc("/notifications", handlers.NotificationAPIHandler).Methods("GET", "POST")
	mixed.HandleFunc("/activity", handlers.ActivityAPIHandler).Methods("POST")

	// Validation et traitement de formulaires
	mixed.HandleFunc("/validate", handlers.ValidationAPIHandler).Methods("POST")
	mixed.HandleFunc("/form-processing", handlers.FormProcessingAPIHandler).Methods("POST")
	mixed.HandleFunc("/preprocess", handlers.PreprocessDataAPIHandler).Methods("POST")

	// Routes d'amitiés (authentification requise)
	setupFriendshipRoutes(mixed)

	// Routes avec préfixe v1 (pour compatibilité frontend)
	v1 := api.PathPrefix("/v1").Subrouter()
	v1.Use(middleware.OptionalAuthMiddleware)

	// Mêmes routes que mixed mais avec préfixe v1
	v1.HandleFunc("/threads/{id:[0-9]+}/like", handlers.ToggleLikeHandler).Methods("POST")
	v1.HandleFunc("/messages/{id:[0-9]+}/like", handlers.MessageLikeHandler).Methods("POST")
	v1.HandleFunc("/threads/{id:[0-9]+}/messages", handlers.ThreadMessagesHandler).Methods("GET", "POST")
	v1.HandleFunc("/messages/{id:[0-9]+}/vote", handlers.MessageVoteHandler).Methods("POST")
	v1.HandleFunc("/profile", handlers.ProfileAPIHandler).Methods("GET")
	v1.HandleFunc("/notifications", handlers.NotificationAPIHandler).Methods("GET", "POST")
	v1.HandleFunc("/activity", handlers.ActivityAPIHandler).Methods("POST")
	v1.HandleFunc("/validate", handlers.ValidationAPIHandler).Methods("POST")
	v1.HandleFunc("/form-processing", handlers.FormProcessingAPIHandler).Methods("POST")
	v1.HandleFunc("/preprocess", handlers.PreprocessDataAPIHandler).Methods("POST")

	// Routes d'amitiés pour v1 aussi
	setupFriendshipRoutes(v1)
}

// setupFriendshipRoutes configure les routes pour l'API des amitiés
func setupFriendshipRoutes(router *mux.Router) {
	// Créer le handler d'amitiés
	db := database.DB
	friendshipService := services.NewFriendshipServiceWithDB(db)
	friendshipHandler := handlers.NewFriendshipHandler(friendshipService)

	// Routes pour les demandes d'amitié
	router.HandleFunc("/friends/request", friendshipHandler.SendFriendRequest).Methods("POST")
	router.HandleFunc("/friends/request/accept", friendshipHandler.AcceptFriendRequest).Methods("POST")
	router.HandleFunc("/friends/request/reject", friendshipHandler.RejectFriendRequest).Methods("POST")
	router.HandleFunc("/friends/request/cancel", friendshipHandler.CancelFriendRequest).Methods("POST")

	// Routes pour la gestion des amis
	router.HandleFunc("/friends", friendshipHandler.GetFriends).Methods("GET")
	router.HandleFunc("/friends/{friendId:[0-9]+}", friendshipHandler.RemoveFriend).Methods("DELETE")
	router.HandleFunc("/friends/requests", friendshipHandler.GetFriendRequests).Methods("GET")
	router.HandleFunc("/friends/requests/sent", friendshipHandler.GetSentRequests).Methods("GET")

	// Routes pour les amis mutuels
	router.HandleFunc("/friends/mutual/{userId:[0-9]+}", friendshipHandler.GetMutualFriends).Methods("GET")

	// Routes pour la recherche et suggestions
	router.HandleFunc("/users/search", friendshipHandler.SearchUsers).Methods("GET")
	router.HandleFunc("/friends/suggestions", friendshipHandler.GetSuggestedFriends).Methods("GET")

	// Routes pour les statistiques et statuts
	router.HandleFunc("/friends/stats", friendshipHandler.GetFriendshipStats).Methods("GET")
	router.HandleFunc("/friends/status/{userId:[0-9]+}", friendshipHandler.GetFriendshipStatus).Methods("GET")

	// Routes pour bloquer/débloquer
	router.HandleFunc("/users/{userId:[0-9]+}/block", friendshipHandler.BlockUser).Methods("POST")
	router.HandleFunc("/users/{userId:[0-9]+}/unblock", friendshipHandler.UnblockUser).Methods("POST")
}
