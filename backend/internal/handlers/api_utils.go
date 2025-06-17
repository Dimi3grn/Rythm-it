// Fichier: backend/internal/handlers/api_utils.go
package handlers

import (
	"encoding/json"
	"net/http"
)

// APIResponse structure standard pour toutes les réponses API
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// sendAPISuccess envoie une réponse de succès standardisée
func sendAPISuccess(w http.ResponseWriter, message string, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	}

	json.NewEncoder(w).Encode(response)
}

// sendAPIError envoie une réponse d'erreur standardisée
func sendAPIError(w http.ResponseWriter, errorMsg string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	response := APIResponse{
		Success: false,
		Error:   errorMsg,
	}

	json.NewEncoder(w).Encode(response)
}
