package router

import (
	"fmt"
	"net/http"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/middleware"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"

	"github.com/gorilla/mux"
	"github.com/rs/cors"
)

// Router instance globale
var Router *mux.Router

// Init initialise le router avec toutes les routes
func Init(cfg *configs.Config) *mux.Router {
	Router = mux.NewRouter()

	// Middleware global
	Router.Use(middleware.Logger)
	Router.Use(middleware.Recovery)

	// Initialiser les services et contrôleurs
	userRepo := repositories.NewUserRepository(database.DB)
	authService := services.NewAuthService(userRepo, cfg)
	authController := controllers.NewAuthController(authService)

	// Routes API
	api := Router.PathPrefix("/api").Subrouter()
	api.Use(middleware.JSONMiddleware)

	// Health routes (toujours accessibles)
	healthController := &controllers.HealthController{}
	registerRoutes(api, healthController)

	// Routes publiques (pas besoin d'auth)
	public := api.PathPrefix("/public").Subrouter()
	setupPublicRoutes(public, authController)

	// Routes protégées (auth requise)
	protected := api.PathPrefix("/v1").Subrouter()
	protected.Use(middleware.AuthMiddleware)
	setupProtectedRoutes(protected, authController)

	// Route racine
	Router.HandleFunc("/", homeHandler).Methods("GET")

	// Health check direct à la racine (plus pratique)
	Router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"healthy","service":"rythmit-api","timestamp":` + fmt.Sprintf("%d", time.Now().Unix()) + `}`))
	}).Methods("GET")

	// Configuration CORS
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173"},
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

// registerRoutes enregistre les routes d'un controller
func registerRoutes(router *mux.Router, controller controllers.BaseController) {
	for _, route := range controller.Routes() {
		router.HandleFunc(route.Pattern, route.HandlerFunc).Methods(route.Method).Name(route.Name)
	}
}

// setupPublicRoutes configure les routes publiques avec AuthController
func setupPublicRoutes(router *mux.Router, authController *controllers.AuthController) {
	// Auth routes (inscription/connexion) - MAINTENANT FONCTIONNELLES
	router.HandleFunc("/register", authController.Register).Methods("POST")
	router.HandleFunc("/login", authController.Login).Methods("POST")
	router.HandleFunc("/refresh", authController.RefreshToken).Methods("POST")

	// Threads publics (consultation sans auth)
	router.HandleFunc("/threads", handleNotImplemented).Methods("GET")
	router.HandleFunc("/threads/{id:[0-9]+}", handleNotImplemented).Methods("GET")

	// Battles publiques
	router.HandleFunc("/battles/active", handleNotImplemented).Methods("GET")
	router.HandleFunc("/battles/{id:[0-9]+}", handleNotImplemented).Methods("GET")
}

// setupProtectedRoutes configure les routes protégées avec AuthController
func setupProtectedRoutes(router *mux.Router, authController *controllers.AuthController) {
	// User routes - MAINTENANT FONCTIONNELLES
	router.HandleFunc("/profile", authController.GetProfile).Methods("GET")
	router.HandleFunc("/profile", authController.UpdateProfile).Methods("PUT")

	// Thread management (à implémenter)
	router.HandleFunc("/threads", handleNotImplemented).Methods("POST")
	router.HandleFunc("/threads/{id:[0-9]+}", handleNotImplemented).Methods("PUT", "DELETE")

	// Messages (à implémenter)
	router.HandleFunc("/threads/{id:[0-9]+}/messages", handleNotImplemented).Methods("GET", "POST")
	router.HandleFunc("/messages/{id:[0-9]+}/fire", handleNotImplemented).Methods("POST")
	router.HandleFunc("/messages/{id:[0-9]+}/skip", handleNotImplemented).Methods("POST")

	// Battles (à implémenter)
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
		"message": "Bienvenue sur Rythmit API - Authentification fonctionnelle!",
		"version": "0.1.0",
		"status": "Phase 1 terminée",
		"endpoints": {
			"health": "/health",
			"auth": {
				"register": "POST /api/public/register",
				"login": "POST /api/public/login",
				"profile": "GET /api/v1/profile (auth required)",
				"refresh": "POST /api/public/refresh"
			},
			"docs": "Coming soon"
		}
	}`))
}

// handleNotImplemented pour les routes pas encore implémentées
func handleNotImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte(`{
		"error": "Cette fonctionnalité n'est pas encore implémentée", 
		"status": 501,
		"message": "En cours de développement pour la Phase 2"
	}`))
}
