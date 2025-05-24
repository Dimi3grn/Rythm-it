package utils

import (
	"encoding/json"
	"net/http"
	"time"
)

// Response structure standard pour toutes les réponses API
type Response struct {
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Error     string      `json:"error,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PaginatedResponse structure pour les réponses paginées
type PaginatedResponse struct {
	Success    bool        `json:"success"`
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
	Timestamp  int64       `json:"timestamp"`
}

// Pagination informations de pagination
type Pagination struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

// ValidationError structure pour les erreurs de validation
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// SendJSON envoie une réponse JSON avec le status code spécifié
func SendJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Erreur encodage JSON", http.StatusInternalServerError)
	}
}

// Success envoie une réponse de succès
func Success(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	SendJSON(w, http.StatusOK, response)
}

// Created envoie une réponse de création réussie (201)
func Created(w http.ResponseWriter, message string, data interface{}) {
	response := Response{
		Success:   true,
		Message:   message,
		Data:      data,
		Timestamp: time.Now().Unix(),
	}
	SendJSON(w, http.StatusCreated, response)
}

// Error envoie une réponse d'erreur
func Error(w http.ResponseWriter, statusCode int, message string) {
	response := Response{
		Success:   false,
		Error:     message,
		Timestamp: time.Now().Unix(),
	}
	SendJSON(w, statusCode, response)
}

// BadRequest envoie une erreur 400
func BadRequest(w http.ResponseWriter, message string) {
	Error(w, http.StatusBadRequest, message)
}

// Unauthorized envoie une erreur 401
func Unauthorized(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Authentification requise"
	}
	Error(w, http.StatusUnauthorized, message)
}

// Forbidden envoie une erreur 403
func Forbidden(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Accès interdit"
	}
	Error(w, http.StatusForbidden, message)
}

// NotFound envoie une erreur 404
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Ressource non trouvée"
	}
	Error(w, http.StatusNotFound, message)
}

// InternalServerError envoie une erreur 500
func InternalServerError(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Erreur interne du serveur"
	}
	Error(w, http.StatusInternalServerError, message)
}

// ValidationErrors envoie une erreur de validation avec détails
func ValidationErrors(w http.ResponseWriter, errors []ValidationError) {
	response := Response{
		Success:   false,
		Error:     "Erreur de validation",
		Data:      errors,
		Timestamp: time.Now().Unix(),
	}
	SendJSON(w, http.StatusBadRequest, response)
}

// Paginated envoie une réponse paginée
func Paginated(w http.ResponseWriter, data interface{}, page, perPage int, total int64) {
	totalPages := int(total) / perPage
	if int(total)%perPage > 0 {
		totalPages++
	}

	response := PaginatedResponse{
		Success: true,
		Data:    data,
		Pagination: Pagination{
			Page:       page,
			PerPage:    perPage,
			Total:      total,
			TotalPages: totalPages,
		},
		Timestamp: time.Now().Unix(),
	}
	SendJSON(w, http.StatusOK, response)
}
