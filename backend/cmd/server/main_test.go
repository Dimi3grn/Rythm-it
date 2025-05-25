package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"rythmitbackend/configs"
	"rythmitbackend/internal/router"
)

func TestServerStart(t *testing.T) {
	// Charger la configuration
	cfg := configs.Load()

	// Initialiser le router
	handler := router.Init(cfg)

	if handler == nil {
		t.Fatal("Le router n'a pas pu être initialisé")
	}
}

func TestHomeEndpoint(t *testing.T) {
	// Créer une requête de test
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Créer un ResponseRecorder pour capturer la réponse
	rr := httptest.NewRecorder()

	// Initialiser le router
	cfg := configs.Load()
	handler := router.Init(cfg)

	// Exécuter la requête
	handler.ServeHTTP(rr, req)

	// Vérifier le status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler retourne un mauvais status code: got %v want %v",
			status, http.StatusOK)
	}

	// Vérifier que la réponse contient "Rythmit"
	expected := "Rythmit"
	if !contains(rr.Body.String(), expected) {
		t.Errorf("Handler retourne un body inattendu: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestHealthEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		want     int
	}{
		{"Health API", "/api/health", http.StatusOK},
		{"Ready API", "/api/ready", http.StatusOK},
		{"Health Root", "/health", http.StatusOK},
	}

	cfg := configs.Load()
	handler := router.Init(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", tt.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != tt.want {
				t.Errorf("%s: mauvais status code: got %v want %v",
					tt.name, status, tt.want)
			}
		})
	}
}

func TestNotImplementedEndpoints(t *testing.T) {
	endpoints := []struct {
		method   string
		endpoint string
	}{
		{"GET", "/api/public/threads"},
		{"POST", "/api/public/register"},
		{"POST", "/api/public/login"},
	}

	cfg := configs.Load()
	handler := router.Init(cfg)

	for _, ep := range endpoints {
		t.Run(ep.endpoint, func(t *testing.T) {
			req, err := http.NewRequest(ep.method, ep.endpoint, nil)
			if err != nil {
				t.Fatal(err)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if status := rr.Code; status != http.StatusNotImplemented {
				t.Errorf("%s %s devrait retourner 501: got %v", ep.method, ep.endpoint, status)
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr || len(s) > len(substr) && contains(s[1:], substr)
}
