// Fichier: backend/internal/router/router.go
package router

import (
	"net/http"
	"text/template"

	"rythmitbackend/configs"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/handlers"
	"rythmitbackend/internal/middleware"

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

	// Routes pour les pages HTML
	setupPageRoutes()

	// Routes API
	api := Router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONMiddleware)

	// Health routes (toujours accessibles)
	healthController := &controllers.HealthController{}
	registerRoutes(api, healthController)

	// Routes publiques (pas besoin d'auth)
	public := api.PathPrefix("/public").Subrouter()
	setupPublicRoutes(public)

	// Routes protégées (auth requise)
	protected := api.PathPrefix("/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	setupProtectedRoutes(protected)

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
	Router.HandleFunc("/", indexHandler).Methods("GET")

	// Autres pages
	Router.HandleFunc("/discover", discoverHandler).Methods("GET")
	Router.HandleFunc("/friends", friendsHandler).Methods("GET")
	Router.HandleFunc("/messages", messagesHandler).Methods("GET")
	Router.HandleFunc("/profile", profileHandler).Methods("GET")
	Router.HandleFunc("/settings", settingsHandler).Methods("GET")
	Router.HandleFunc("/hub", hubHandler).Methods("GET")

	// Pages d'authentification
	Router.HandleFunc("/signin", signinHandler).Methods("GET")
	Router.HandleFunc("/signup", signupHandler).Methods("GET")
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
	router.HandleFunc("/threads", handleNotImplemented).Methods("GET")
	router.HandleFunc("/threads/{id:[0-9]+}", handleNotImplemented).Methods("GET")

	// Battles publiques
	router.HandleFunc("/battles/active", handleNotImplemented).Methods("GET")
	router.HandleFunc("/battles/{id:[0-9]+}", handleNotImplemented).Methods("GET")
}

// setupProtectedRoutes configure les routes protégées
func setupProtectedRoutes(router *mux.Router) {
	// User routes
	router.HandleFunc("/profile", profileAPIHandler).Methods("GET")
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

// handleNotImplemented pour les routes pas encore implémentées
func handleNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{"error": "Cette fonctionnalité n'est pas encore implémentée", "status": 501}`))
}

// Page handlers - utilisant les handlers existants du package handlers
func indexHandler(w http.ResponseWriter, r *http.Request) {
	handlers.IndexHandler(w, r)
}

func discoverHandler(w http.ResponseWriter, r *http.Request) {
	handlers.DiscoverHandler(w, r)
}

func friendsHandler(w http.ResponseWriter, r *http.Request) {
	handlers.FriendsHandler(w, r)
}

func messagesHandler(w http.ResponseWriter, r *http.Request) {
	handlers.MessagesHandler(w, r)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
	handlers.ProfileHandler(w, r)
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	handlers.SettingsHandler(w, r)
}

func hubHandler(w http.ResponseWriter, r *http.Request) {
	handlers.HubHandler(w, r)
}

func signinHandler(w http.ResponseWriter, r *http.Request) {
	handlers.SigninHandler(w, r)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	handlers.SignupHandler(w, r)
}

// API handlers
func profileAPIHandler(w http.ResponseWriter, r *http.Request) {
	handlers.ProfileAPIHandler(w, r)
}
