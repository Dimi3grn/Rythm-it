package handlers

import (
	"encoding/json"
	"net/http"
	"rythmitbackend/internal/controllers"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/services"

	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// MessageHandler gère les requêtes liées aux messages directs
type MessageHandler struct {
	messageService services.MessageService
}

// NewMessageHandler crée une nouvelle instance du handler
func NewMessageHandler(messageService services.MessageService) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
	}
}

// GetConversations récupère toutes les conversations de l'utilisateur
func (h *MessageHandler) GetConversations(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	conversations, err := h.messageService.GetUserConversations(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération conversations", http.StatusInternalServerError)
		return
	}

	if conversations == nil {
		conversations = []*models.ConversationWithDetails{}
	}

	// Mettre à jour le statut en ligne réel depuis le MessageHub
	messageHub := GetMessageHub()
	for _, conv := range conversations {
		if conv.OtherUser != nil {
			conv.IsOnline = messageHub.IsUserOnline(conv.OtherUser.ID)
		}
	}

	sendAPISuccess(w, "Conversations récupérées avec succès", map[string]interface{}{
		"conversations": conversations,
	})
}

// GetOrCreateConversation récupère ou crée une conversation
func (h *MessageHandler) GetOrCreateConversation(w http.ResponseWriter, r *http.Request) {
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

	conversation, err := h.messageService.GetOrCreateConversation(userID, uint(otherUserID))
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Conversation récupérée", map[string]interface{}{
		"conversation": conversation,
	})
}

// GetConversationMessages récupère les messages d'une conversation
func (h *MessageHandler) GetConversationMessages(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	conversationIDStr := vars["conversationId"]
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID conversation invalide", http.StatusBadRequest)
		return
	}

	// Paramètres de pagination
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 50
	if limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	offset := 0
	if offsetStr != "" {
		if parsedOffset, err := strconv.Atoi(offsetStr); err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	messages, err := h.messageService.GetConversationMessages(uint(conversationID), userID, limit, offset)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusForbidden)
		return
	}

	if messages == nil {
		messages = []*models.DirectMessage{}
	}

	sendAPISuccess(w, "Messages récupérés", map[string]interface{}{
		"messages": messages,
	})
}

// SendMessage envoie un message (via REST API, pas WebSocket)
func (h *MessageHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	var req struct {
		ReceiverID uint   `json:"receiver_id"`
		Content    string `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendAPIError(w, "Données invalides", http.StatusBadRequest)
		return
	}

	if req.ReceiverID == 0 {
		sendAPIError(w, "ID destinataire requis", http.StatusBadRequest)
		return
	}

	if req.Content == "" {
		sendAPIError(w, "Le message ne peut pas être vide", http.StatusBadRequest)
		return
	}

	message, err := h.messageService.SendMessage(userID, req.ReceiverID, req.Content)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Notifier via WebSocket
	wsMsg := &models.WebSocketMessage{
		Type:      "message",
		Message:   message,
		Timestamp: time.Now(),
	}
	messageHub := GetMessageHub()
	messageHub.broadcast <- wsMsg

	sendAPISuccess(w, "Message envoyé", map[string]interface{}{
		"message": message,
	})
}

// MarkAsRead marque une conversation comme lue
func (h *MessageHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	conversationIDStr := vars["conversationId"]
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID conversation invalide", http.StatusBadRequest)
		return
	}

	err = h.messageService.MarkConversationAsRead(uint(conversationID), userID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusForbidden)
		return
	}

	// Notifier via WebSocket
	wsMsg := &models.WebSocketMessage{
		Type: "read",
		Data: map[string]interface{}{
			"conversation_id": conversationID,
			"user_id":         userID,
		},
		Timestamp: time.Now(),
	}
	messageHub := GetMessageHub()
	messageHub.broadcast <- wsMsg

	sendAPISuccess(w, "Conversation marquée comme lue", nil)
}

// GetUnreadCount récupère le nombre de messages non lus
func (h *MessageHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	count, err := h.messageService.GetUnreadCount(userID)
	if err != nil {
		sendAPIError(w, "Erreur récupération compteur", http.StatusInternalServerError)
		return
	}

	sendAPISuccess(w, "Compteur récupéré", map[string]interface{}{
		"unread_count": count,
	})
}

// DeleteConversation supprime une conversation
func (h *MessageHandler) DeleteConversation(w http.ResponseWriter, r *http.Request) {
	userID, exists := controllers.GetUserIDFromContext(r)
	if !exists {
		sendAPIError(w, "Utilisateur non authentifié", http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(r)
	conversationIDStr := vars["conversationId"]
	conversationID, err := strconv.ParseUint(conversationIDStr, 10, 32)
	if err != nil {
		sendAPIError(w, "ID conversation invalide", http.StatusBadRequest)
		return
	}

	err = h.messageService.DeleteConversation(uint(conversationID), userID)
	if err != nil {
		sendAPIError(w, err.Error(), http.StatusForbidden)
		return
	}

	sendAPISuccess(w, "Conversation supprimée", nil)
}
