package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"rythmitbackend/configs"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/middleware"

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
