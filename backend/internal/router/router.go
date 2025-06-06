package router

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"path"
	"path/filepath"
	"strings"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/middleware"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/web"
	"rythmitbackend/pkg/database"

	// "rythmitbackend/internal/web"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Router instance globale
var Router *mux.Router

// Remplacer la fonction debugProfileHandler dans router.go par :

// Remplacer la fonction debugProfileHandler dans router.go par :

func profileHandler(w http.ResponseWriter, r *http.Request) {
	// Pour l'instant, retourner un profil fictif
	// Plus tard, vous extrairez l'user_id du token JWT et ferez une vraie requête DB

	profileData := map[string]interface{}{
		"id":               1,
		"username":         "admin",
		"email":            "admin@rythmit.com",
		"is_admin":         true,
		"message_count":    0,
		"thread_count":     0,
		"favorite_genres":  []string{"rap", "hip-hop"},
		"favorite_artists": []string{"kendrick lamar", "drake"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Utiliser le helper response
	// utils.Success(w, "Profil récupéré", profileData)

	// Ou réponse simple pour l'instant :
	response := map[string]interface{}{
		"success":   true,
		"message":   "Profil récupéré avec succès",
		"data":      profileData,
		"timestamp": time.Now().Unix(),
	}

	json.NewEncoder(w).Encode(response)
}

// Init initialise le router avec toutes les routes
func Init(cfg *configs.Config) *mux.Router {
	Router = mux.NewRouter()

	// Middleware global
	Router.Use(middleware.Logger)
	Router.Use(middleware.Recovery)

	// Configuration CORS pour permettre la communication frontend-backend
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8085", "http://127.0.0.1:8085"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Requested-With"},
		ExposedHeaders:   []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           86400,
	})
	Router.Use(c.Handler)

	// Initialiser les repositories et contrôleurs
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	threadController := controllers.NewThreadController(threadRepo)

	// --- Routes API ---
	api := Router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONMiddleware)

	// Health routes
	healthController := &controllers.HealthController{}
	registerRoutes(api, healthController)

	// Routes publiques API
	publicAPI := api.PathPrefix("/public").Subrouter()
	publicAPI.HandleFunc("/threads", threadController.GetPublicThreads).Methods("GET")
	setupPublicRoutes(publicAPI)

	// Routes protégées API
	protectedAPI := api.PathPrefix("/v1").Subrouter()
	protectedAPI.Use(middleware.AuthMiddleware)
	setupProtectedRoutes(protectedAPI)

	// --- Routes Web (fichiers statiques) ---
	// Servir les fichiers statiques depuis le répertoire 'frontend'
	frontendPath := "../frontend"
	homePage := "index.html"

	// Servir le répertoire frontend complet sous la racine (/) pour capturer index.html
	Router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Si la requête commence par /api, la laisser passer
		if strings.HasPrefix(r.URL.Path, "/api") {
			Router.ServeHTTP(w, r)
			return
		}

		// Sinon, servir les fichiers statiques
		requestPath := r.URL.Path
		if requestPath == "/" {
			requestPath = "/" + homePage
		}

		// Nettoyer le chemin pour éviter les traversées de répertoire
		cleanedPath := path.Clean(requestPath)

		// Construire le chemin complet vers le fichier dans le répertoire frontend
		fullPath := filepath.Join(frontendPath, cleanedPath)

		// Servir le fichier
		http.ServeFile(w, r, fullPath)
	})).Methods("GET")

	return Router
}

// registerRoutes enregistre les routes d'un controller
func registerRoutes(router *mux.Router, controller controllers.BaseController) {
	for _, route := range controller.Routes() {
		router.HandleFunc(route.Pattern, route.HandlerFunc).Methods(route.Method).Name(route.Name)
	}
}

// setupPublicRoutes configure les routes publiques
func setupPublicRoutes(router *mux.Router) {
	// Auth routes (inscription/connexion)
	router.HandleFunc("/register", handleNotImplemented).Methods("POST")
	router.HandleFunc("/login", handleNotImplemented).Methods("POST")

	// Threads publics
	// La route /threads est maintenant gérée par le ThreadController.GetPublicThreads
	// router.HandleFunc("/threads", handleNotImplemented).Methods("GET") // COMMENTÉ OU SUPPRIMÉ
	router.HandleFunc("/threads/{id:[0-9]+}", handleNotImplemented).Methods("GET")

	// Battles publiques
	router.HandleFunc("/battles/active", handleNotImplemented).Methods("GET")
	router.HandleFunc("/battles/{id:[0-9]+}", handleNotImplemented).Methods("GET")
}

// setupProtectedRoutes configure les routes protégées
func setupProtectedRoutes(router *mux.Router) {
	// User routes - MODIFICATION ICI
	router.HandleFunc("/profile", profileHandler).Methods("GET") // <-- CHANGÉ
	router.HandleFunc("/profile", handleNotImplemented).Methods("PUT")

	// Thread management
	router.HandleFunc("/threads", handleNotImplemented).Methods("POST")
	router.HandleFunc("/threads/{id:[0-9]+}", handleNotImplemented).Methods("PUT", "DELETE")

	// Messages
	router.HandleFunc("/threads/{id:[0-9]+}/messages", handleNotImplemented).Methods("GET", "POST")
	router.HandleFunc("/messages/{id:[0-9]+}/fire", handleNotImplemented).Methods("POST")
	router.HandleFunc("/messages/{id:[0-9]+}/skip", handleNotImplemented).Methods("POST")

	// Battles
	router.HandleFunc("/battles", handleNotImplemented).Methods("POST")
	router.HandleFunc("/battles/{id:[0-9]+}/vote", handleNotImplemented).Methods("POST")

	// Admin routes
	admin := router.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.AdminMiddleware)

	admin.HandleFunc("/dashboard", handleNotImplemented).Methods("GET")
	admin.HandleFunc("/users/{id:[0-9]+}/ban", handleNotImplemented).Methods("POST")
	admin.HandleFunc("/threads/{id:[0-9]+}/state", handleNotImplemented).Methods("PUT")
}

// homeHandler page d'accueil
func homeHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{
		"message": "Bienvenue sur Rythmit API - Hot Reload fonctionne!",
		"version": "0.1.0",
		"endpoints": {
			"health": "/health",
			"api": "/api",
			"docs": "Coming soon"
		}
	}`))
}

// handleNotImplemented pour les routes pas encore implémentées
func handleNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "Cette fonctionnalité n'est pas encore implémentée", "status": 501}`))
}

// SetupRoutes configure toutes les routes de l'application
func SetupRoutes(router *mux.Router, db *sql.DB, tm *web.TemplateManager) {

	// Désactiver la route de la page d'accueil qui utilisait les templates backend
	// router.HandleFunc("/", web.HomeHandler(db, tm)).Methods("GET")

	// Servir les fichiers statiques depuis le répertoire 'frontend'
	// Cela permettra de servir index.html, les styles, les scripts, etc.
	frontendPath := "../frontend"
	homePage := "index.html"

	// Servir le répertoire frontend complet sous la racine (/) pour capturer index.html
	router.PathPrefix("/").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Construire le chemin du fichier demandé
		// Si la requête est pour la racine (/), servir index.html
		// Sinon, servir le fichier demandé dans le répertoire frontend
		requestPath := r.URL.Path
		if requestPath == "/" {
			requestPath = "/" + homePage
		}

		// Nettoyer le chemin pour éviter les traversées de répertoire
		cleanedPath := path.Clean(requestPath)

		// Construire le chemin complet vers le fichier dans le répertoire frontend
		fullPath := filepath.Join(frontendPath, cleanedPath)

		// Servir le fichier
		http.ServeFile(w, r, fullPath)

	})).Methods("GET")

	// Ancienne configuration pour servir les assets sous /frontend_assets/, désactivée.
	// fileServer := http.FileServer(http.Dir("../frontend"))
	// router.PathPrefix("/frontend_assets/").Handler(http.StripPrefix("/frontend_assets/", fileServer))

	// Initialiser les repositories et contrôleurs
	threadRepo := repositories.NewThreadRepository(db)
	// battleRepo := repositories.NewBattleRepository(db) // Initialiser le BattleRepository
	threadController := controllers.NewThreadController(threadRepo)

	// Initialiser le gestionnaire de templates web
	tmpl := web.NewTemplateManager()
	if err := tmpl.LoadTemplates(); err != nil {
		// Utiliser log.Fatal ou panic selon la gravité
		panic(fmt.Sprintf("Erreur chargement templates: %v", err))
	}

	// --- Routes Web (servies par Go avec templates) ---
	// Route racine gérée par le handler de page d'accueil web
	// homeHandlerWeb := web.NewHomeHandler(threadRepo, battleRepo, tmpl) // Passer le battleRepo
	// router.Handle("/", homeHandlerWeb).Methods("GET")

	// TODO: Ajouter d'autres routes web pour les pages HTML (login, register, etc.)
	// Exemple pour la page de login :
	// loginHandlerWeb := web.NewTemplateHandler(tmpl, "login.html", func(r *http.Request) (web.PageData, error) {
	//     return web.BasePageData("Connexion"), nil
	// })
	// Router.Handle("/login", loginHandlerWeb).Methods("GET")

	// --- Routes API (restent inchangées) ---
	api := router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONMiddleware)

	// Health routes (toujours accessibles)
	healthController := &controllers.HealthController{}
	registerRoutes(api, healthController)

	// Routes publiques API
	publicAPI := api.PathPrefix("/public").Subrouter()
	// La route /api/public/threads n'est plus nécessaire si le frontend ne l'appelle pas directement
	// mais on peut la laisser pour une API publique si souhaité
	publicAPI.HandleFunc("/threads", threadController.GetPublicThreads).Methods("GET")
	setupPublicRoutes(publicAPI) // Cette fonction contient encore des routes 'not implemented' API publiques

	// Routes protégées API (auth requise)
	protectedAPI := api.PathPrefix("/v1").Subrouter()
	protectedAPI.Use(middleware.AuthMiddleware)
	setupProtectedRoutes(protectedAPI)

	// Health check direct à la racine (peut être laissé ou supprimé si /health suffit)
	router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"rythmit-api","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `}`))
	}).Methods("GET")

}
