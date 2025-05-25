package controllers

import (
	"fmt"
	"net/http"
	"rythmitbackend/pkg/database"
)

// BaseController interface que tous les controllers doivent implémenter
type BaseController interface {
	// Routes retourne les routes gérées par ce controller
	Routes() []Route
}

// Route structure définissant une route HTTP
type Route struct {
	Name        string
	Method      string
	Pattern     string
	HandlerFunc http.HandlerFunc
	Protected   bool // true si la route nécessite une authentification
}

// Controller structure de base avec méthodes communes
type Controller struct {
	// Peut contenir des dépendances communes comme logger, config, etc.
}

// HealthController pour les endpoints de santé
type HealthController struct {
	Controller
}

// Routes implémente l'interface BaseController
func (hc *HealthController) Routes() []Route {
	return []Route{
		{
			Name:        "Health",
			Method:      "GET",
			Pattern:     "/health",
			HandlerFunc: hc.HealthCheck,
			Protected:   false,
		},
		{
			Name:        "Ready",
			Method:      "GET",
			Pattern:     "/ready",
			HandlerFunc: hc.ReadinessCheck,
			Protected:   false,
		},
	}
}

// HealthCheck vérifie que l'API est en ligne
func (hc *HealthController) HealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy","service":"rythmit-api"}`))
}

// ReadinessCheck vérifie que l'API est prête (DB connectée, etc.)
func (hc *HealthController) ReadinessCheck(w http.ResponseWriter, r *http.Request) {
	dbStatus := "connected"
	if err := database.Health(); err != nil {
		dbStatus = "disconnected"
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(`{"status":"ready","database":"%s"}`, dbStatus)))
}
