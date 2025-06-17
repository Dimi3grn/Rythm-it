package handlers

import (
	"encoding/json"
	"net/http"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/services"
	"strconv"

	"github.com/gorilla/mux"
)

// FriendshipHandler gère les requêtes liées aux amitiés
type FriendshipHandler struct {
	friendshipService services.FriendshipService
}

// NewFriendshipHandler crée une nouvelle instance du handler
func NewFriendshipHandler(friendshipService services.FriendshipService) *FriendshipHandler {
	return &FriendshipHandler{
		friendshipService: friendshipService,
	}
}

// SendFriendRequestRequest représente une demande d'envoi d'amitié
type SendFriendRequestRequest struct {
	AddresseeID uint `json:"addressee_id"`
}

// FriendRequestActionRequest représente une action sur une demande d'amitié
type FriendRequestActionRequest struct {
	RequesterID uint `json:"requester_id"`
}

// SearchUsersRequest représente une demande de recherche d'utilisateurs
type SearchUsersRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit,omitempty"`
}

// SendFriendRequest envoie une demande d'amitié
func (h *FriendshipHandler) SendFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	var req SendFriendRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAPIError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.AddresseeID == 0 {
		sendAPIError(w, "ID utilisateur destinataire requis", http.StatusBadRequest)
		return
	}

	err := h.friendshipService.SendFriendRequest(userID, req.AddresseeID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Demande d'amitié envoyée avec succès", nil)
}

// AcceptFriendRequest accepte une demande d'amitié
func (h *FriendshipHandler) AcceptFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	var req FriendRequestActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAPIError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.RequesterID == 0 {
		sendAPIError(w, "ID utilisateur demandeur requis", http.StatusBadRequest)
		return
	}

	err := h.friendshipService.AcceptFriendRequest(req.RequesterID, userID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Demande d'amitié acceptée avec succès", nil)
}

// RejectFriendRequest rejette une demande d'amitié
func (h *FriendshipHandler) RejectFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	var req FriendRequestActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAPIError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.RequesterID == 0 {
		sendAPIError(w, "ID utilisateur demandeur requis", http.StatusBadRequest)
		return
	}

	err := h.friendshipService.RejectFriendRequest(req.RequesterID, userID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Demande d'amitié rejetée avec succès", nil)
}

// CancelFriendRequest annule une demande d'amitié envoyée
func (h *FriendshipHandler) CancelFriendRequest(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	var req FriendRequestActionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAPIError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.RequesterID == 0 {
		sendAPIError(w, "ID utilisateur destinataire requis", http.StatusBadRequest)
		return
	}

	err := h.friendshipService.CancelFriendRequest(userID, req.RequesterID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Demande d'amitié annulée avec succès", nil)
}

// RemoveFriend supprime un ami
func (h *FriendshipHandler) RemoveFriend(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	friendIDStr := vars["friendId"]
	friendID, err := strconv.ParseUint(friendIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID ami invalide", http.StatusBadRequest)
		return
	}

	err = h.friendshipService.RemoveFriend(userID, uint(friendID))
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Ami supprimé avec succès", nil)
}

// BlockUser bloque un utilisateur
func (h *FriendshipHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	blockedIDStr := vars["userId"]
	blockedID, err := strconv.ParseUint(blockedIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	err = h.friendshipService.BlockUser(userID, uint(blockedID))
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Utilisateur bloqué avec succès", nil)
}

// UnblockUser débloque un utilisateur
func (h *FriendshipHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	blockedIDStr := vars["userId"]
	blockedID, err := strconv.ParseUint(blockedIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	err = h.friendshipService.UnblockUser(userID, uint(blockedID))
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Utilisateur débloqué avec succès", nil)
}

// GetFriends récupère la liste des amis
func (h *FriendshipHandler) GetFriends(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	friends, err := h.friendshipService.GetFriends(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération amis", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Amis récupérés avec succès", map[string]interface{}{
		"friends": friends,
	})
}

// GetFriendRequests récupère les demandes d'amitié reçues
func (h *FriendshipHandler) GetFriendRequests(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	requests, err := h.friendshipService.GetFriendRequests(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération demandes d'amitié", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Demandes d'amitié récupérées avec succès", map[string]interface{}{
		"requests": requests,
	})
}

// GetSentRequests récupère les demandes d'amitié envoyées
func (h *FriendshipHandler) GetSentRequests(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	requests, err := h.friendshipService.GetSentRequests(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération demandes envoyées", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Demandes envoyées récupérées avec succès", map[string]interface{}{
		"requests": requests,
	})
}

// GetMutualFriends récupère les amis mutuels avec un autre utilisateur
func (h *FriendshipHandler) GetMutualFriends(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	otherUserIDStr := vars["userId"]
	otherUserID, err := strconv.ParseUint(otherUserIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	friends, err := h.friendshipService.GetMutualFriends(userID, uint(otherUserID))
	if err != nil {
		sendAPIError(w, "Erreur récupération amis mutuels", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Amis mutuels récupérés avec succès", map[string]interface{}{
		"mutual_friends": friends,
	})
}

// SearchUsers recherche des utilisateurs
func (h *FriendshipHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	query := r.URL.Query().Get("q")
	if query == "" {
		sendAPIError(w, "Paramètre de recherche requis", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 20 // Valeur par défaut
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	users, err := h.friendshipService.SearchUsers(query, userID, limit)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Utilisateurs trouvés", map[string]interface{}{
		"users": users,
	})
}

// GetSuggestedFriends récupère des suggestions d'amis
func (h *FriendshipHandler) GetSuggestedFriends(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	limit := 10 // Valeur par défaut
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	suggestions, err := h.friendshipService.GetSuggestedFriends(userID, limit)
	if err != nil {
		sendAPIError(w, "Erreur récupération suggestions", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Suggestions récupérées avec succès", map[string]interface{}{
		"suggestions": suggestions,
	})
}

// GetFriendshipStats récupère les statistiques d'amitié
func (h *FriendshipHandler) GetFriendshipStats(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	stats, err := h.friendshipService.GetFriendshipStats(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération statistiques", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Statistiques récupérées avec succès", map[string]interface{}{
		"stats": stats,
	})
}

// GetFriendshipStatus récupère le statut d'amitié avec un autre utilisateur
func (h *FriendshipHandler) GetFriendshipStatus(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	otherUserIDStr := vars["userId"]
	otherUserID, err := strconv.ParseUint(otherUserIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID utilisateur invalide", http.StatusBadRequest)
		return
	}

	status, err := h.friendshipService.GetFriendshipStatus(userID, uint(otherUserID))
	if err != nil {
		sendAPIError(w, "Erreur récupération statut", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Statut récupéré avec succès", map[string]interface{}{
		"status": status,
	})
}
