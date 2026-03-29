// Fichier: backend/internal/handlers/notification_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"rythmitbackend/internal/models"

	"github.com/gorilla/websocket"
)

// NotificationManager gère les notifications temps réel
type NotificationManager struct {
	clients    map[uint]*websocket.Conn // UserID -> Connection
	clientsMux sync.RWMutex
	broadcast  chan NotificationMessage
	register   chan ClientConnection
	unregister chan ClientConnection
}

// ClientConnection représente une connexion client
type ClientConnection struct {
	UserID uint
	Conn   *websocket.Conn
}

// NotificationMessage représente un message de notification
type NotificationMessage struct {
	Type      string      `json:"type"`
	Title     string      `json:"title"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data,omitempty"`
	UserID    uint        `json:"user_id"`
	Timestamp time.Time   `json:"timestamp"`
}

// ActivityData structure pour les données d'activité
type ActivityData struct {
	Type       string  `json:"type"`
	UserName   string  `json:"user_name"`
	ThreadID   *uint   `json:"thread_id,omitempty"`
	ThreadName *string `json:"thread_name,omitempty"`
	TrackName  *string `json:"track_name,omitempty"`
	Artist     *string `json:"artist,omitempty"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			// En production, vérifier l'origine
			return true
		},
	}

	notificationManager *NotificationManager
	once                sync.Once
)

// GetNotificationManager retourne l'instance singleton du gestionnaire de notifications
func GetNotificationManager() *NotificationManager {
	once.Do(func() {
		notificationManager = &NotificationManager{
			clients:    make(map[uint]*websocket.Conn),
			broadcast:  make(chan NotificationMessage, 256),
			register:   make(chan ClientConnection),
			unregister: make(chan ClientConnection),
		}
		go notificationManager.run()
	})
	return notificationManager
}

// WebSocketHandler gère les connexions WebSocket pour les notifications
func WebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	// Upgrader la connexion vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Erreur upgrade WebSocket: %v", err)
		return
	}

	// Enregistrer le client
	client := ClientConnection{
		UserID: user.ID,
		Conn:   conn,
	}

	notificationManager := GetNotificationManager()
	notificationManager.register <- client

	// Gérer la connexion en arrière-plan
	go notificationManager.handleClient(client)

	log.Printf("✅ Connexion WebSocket établie pour l'utilisateur %s (ID: %d)", user.Username, user.ID)
}

// run gère le hub de notifications en arrière-plan
func (nm *NotificationManager) run() {
	for {
		select {
		case client := <-nm.register:
			nm.clientsMux.Lock()
			nm.clients[client.UserID] = client.Conn
			nm.clientsMux.Unlock()
			log.Printf("🔌 Client WebSocket enregistré: UserID %d", client.UserID)

		case client := <-nm.unregister:
			nm.clientsMux.Lock()
			if _, ok := nm.clients[client.UserID]; ok {
				delete(nm.clients, client.UserID)
				client.Conn.Close()
			}
			nm.clientsMux.Unlock()
			log.Printf("🔌 Client WebSocket déconnecté: UserID %d", client.UserID)

		case message := <-nm.broadcast:
			nm.clientsMux.RLock()
			if conn, ok := nm.clients[message.UserID]; ok {
				if err := conn.WriteJSON(message); err != nil {
					log.Printf("❌ Erreur envoi notification WebSocket: %v", err)
					conn.Close()
					delete(nm.clients, message.UserID)
				}
			}
			nm.clientsMux.RUnlock()
		}
	}
}

// handleClient gère une connexion client WebSocket
func (nm *NotificationManager) handleClient(client ClientConnection) {
	defer func() {
		nm.unregister <- client
	}()

	// Ping/Pong pour maintenir la connexion
	client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	client.Conn.SetPongHandler(func(string) error {
		client.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// Lire les messages du client (pour maintenir la connexion)
	for {
		_, _, err := client.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ Erreur WebSocket: %v", err)
			}
			break
		}
	}
}

// SendNotification envoie une notification à un utilisateur spécifique
func (nm *NotificationManager) SendNotification(userID uint, notType, title, message string, data interface{}) {
	notification := NotificationMessage{
		Type:      notType,
		Title:     title,
		Message:   message,
		Data:      data,
		UserID:    userID,
		Timestamp: time.Now(),
	}

	select {
	case nm.broadcast <- notification:
	default:
		log.Printf("⚠️ Canal de notification plein, message ignoré pour l'utilisateur %d", userID)
	}
}

// BroadcastActivity diffuse une activité utilisateur
func (nm *NotificationManager) BroadcastActivity(userID uint, userName string, activityType string, data interface{}) {
	// Récupérer la liste des amis pour diffuser l'activité
	friends := nm.getFriendsOfUser(userID)

	activityData := ActivityData{
		Type:     activityType,
		UserName: userName,
	}

	// Enrichir les données selon le type d'activité
	switch activityType {
	case "listening":
		if musicData, ok := data.(map[string]string); ok {
			if track, exists := musicData["track"]; exists {
				activityData.TrackName = &track
			}
			if artist, exists := musicData["artist"]; exists {
				activityData.Artist = &artist
			}
		}
	case "thread_created", "thread_liked":
		if threadData, ok := data.(map[string]interface{}); ok {
			if threadID, ok := threadData["thread_id"].(uint); ok {
				activityData.ThreadID = &threadID
			}
			if threadName, ok := threadData["thread_name"].(string); ok {
				activityData.ThreadName = &threadName
			}
		}
	}

	// Envoyer la notification à tous les amis
	for _, friendID := range friends {
		nm.SendNotification(friendID, "activity",
			fmt.Sprintf("Activité de %s", userName),
			nm.formatActivityMessage(activityType, userName, activityData),
			activityData)
	}
}

// getFriendsOfUser récupère les amis d'un utilisateur (simulation pour l'instant)
func (nm *NotificationManager) getFriendsOfUser(userID uint) []uint {
	// TODO: Implémenter la logique réelle pour récupérer les amis
	// Pour l'instant, retourner une liste vide
	return []uint{}
}

// formatActivityMessage formate le message d'activité
func (nm *NotificationManager) formatActivityMessage(activityType, userName string, data ActivityData) string {
	switch activityType {
	case "listening":
		if data.TrackName != nil && data.Artist != nil {
			return fmt.Sprintf("%s écoute \"%s\" par %s", userName, *data.TrackName, *data.Artist)
		}
		return fmt.Sprintf("%s écoute de la musique", userName)
	case "thread_created":
		if data.ThreadName != nil {
			return fmt.Sprintf("%s a créé un nouveau thread: %s", userName, *data.ThreadName)
		}
		return fmt.Sprintf("%s a créé un nouveau thread", userName)
	case "thread_liked":
		if data.ThreadName != nil {
			return fmt.Sprintf("%s a liké le thread: %s", userName, *data.ThreadName)
		}
		return fmt.Sprintf("%s a liké un thread", userName)
	default:
		return fmt.Sprintf("%s a une nouvelle activité", userName)
	}
}

// NotificationAPIHandler gère les API de notifications
func NotificationAPIHandler(w http.ResponseWriter, r *http.Request) {
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		sendAPIError(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case "GET":
		getNotifications(w, r, user)
	case "POST":
		createNotification(w, r, user)
	default:
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// getNotifications récupère les notifications de l'utilisateur
func getNotifications(w http.ResponseWriter, r *http.Request, user *User) {
	// TODO: Implémenter la récupération des notifications depuis la base de données
	// Pour l'instant, retourner des notifications fictives
	notifications := []map[string]interface{}{
		{
			"id":        1,
			"type":      "like",
			"title":     "Nouveau like",
			"message":   "Votre thread a été liké",
			"read":      false,
			"timestamp": time.Now().Add(-1 * time.Hour),
		},
		{
			"id":        2,
			"type":      "comment",
			"title":     "Nouveau commentaire",
			"message":   "Nouveau commentaire sur votre thread",
			"read":      false,
			"timestamp": time.Now().Add(-2 * time.Hour),
		},
	}

	sendAPISuccess(w, "Notifications récupérées", map[string]interface{}{
		"notifications": notifications,
		"unread_count":  2,
	})
}

// createNotification crée une nouvelle notification
func createNotification(w http.ResponseWriter, r *http.Request, user *User) {
	var requestData struct {
		Type    string      `json:"type"`
		Title   string      `json:"title"`
		Message string      `json:"message"`
		Data    interface{} `json:"data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendAPIError(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	// Envoyer la notification via WebSocket
	nm := GetNotificationManager()
	nm.SendNotification(user.ID, requestData.Type, requestData.Title, requestData.Message, requestData.Data)

	sendAPISuccess(w, "Notification envoyée", map[string]interface{}{
		"sent_at": time.Now(),
	})
}

// ActivityAPIHandler gère les API d'activité utilisateur
func ActivityAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		sendAPIError(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendAPIError(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	// Diffuser l'activité
	nm := GetNotificationManager()
	nm.BroadcastActivity(user.ID, user.Username, requestData.Type, requestData.Data)

	sendAPISuccess(w, "Activité diffusée", map[string]interface{}{
		"activity_type": requestData.Type,
		"broadcast_at":  time.Now(),
	})
}

// ==========================================
// WEBSOCKET POUR MESSAGES DIRECTS
// ==========================================

// MessageClient représente un client WebSocket pour les messages
type MessageClient struct {
	ID     uint
	Conn   *websocket.Conn
	Send   chan []byte
	Hub    *MessageHub
	UserID uint
}

// MessageHub maintient l'ensemble des clients actifs pour les messages
type MessageHub struct {
	clients    map[uint]*MessageClient
	broadcast  chan *models.WebSocketMessage
	register   chan *MessageClient
	unregister chan *MessageClient
	mu         sync.RWMutex
}

var (
	messageHub     *MessageHub
	messageHubOnce sync.Once
)

// GetMessageHub retourne l'instance du hub de messages
func GetMessageHub() *MessageHub {
	messageHubOnce.Do(func() {
		messageHub = &MessageHub{
			broadcast:  make(chan *models.WebSocketMessage),
			register:   make(chan *MessageClient),
			unregister: make(chan *MessageClient),
			clients:    make(map[uint]*MessageClient),
		}
		go messageHub.Run()
	})
	return messageHub
}

// Run démarre la boucle principale du hub de messages
func (h *MessageHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("✅ Client messages connecté: User ID %d (Total: %d)", client.UserID, len(h.clients))

			// Diffuser que l'utilisateur est en ligne
			h.broadcastUserStatus(client.UserID, true)

		case client := <-h.unregister:
			userID := client.UserID
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; ok {
				delete(h.clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("❌ Client messages déconnecté: User ID %d (Total: %d)", userID, len(h.clients))

			// Diffuser que l'utilisateur est hors ligne
			h.broadcastUserStatus(userID, false)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// broadcastMessage diffuse un message aux destinataires appropriés
func (h *MessageHub) broadcastMessage(message *models.WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("❌ Erreur marshalling message: %v", err)
		return
	}

	// Déterminer les destinataires selon le type de message
	switch message.Type {
	case "message":
		if message.Message != nil {
			// Envoyer au destinataire
			if client, ok := h.clients[message.Message.ReceiverID]; ok {
				select {
				case client.Send <- messageJSON:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}

			// Envoyer à l'expéditeur (confirmation)
			if client, ok := h.clients[message.Message.SenderID]; ok {
				select {
				case client.Send <- messageJSON:
				default:
					close(client.Send)
					delete(h.clients, client.UserID)
				}
			}
		}
	case "typing", "read", "status", "user_online", "user_offline":
		// Diffuser à tous les clients de la conversation
		// Pour simplifier, on diffuse à tous les clients connectés
		for _, client := range h.clients {
			select {
			case client.Send <- messageJSON:
			default:
				close(client.Send)
				delete(h.clients, client.UserID)
			}
		}
	}
}

// broadcastUserStatus diffuse le statut en ligne d'un utilisateur
func (h *MessageHub) broadcastUserStatus(userID uint, isOnline bool) {
	msgType := "user_offline"
	if isOnline {
		msgType = "user_online"
	}

	statusMsg := &models.WebSocketMessage{
		Type: msgType,
		Data: map[string]interface{}{
			"user_id":   userID,
			"is_online": isOnline,
		},
		Timestamp: time.Now(),
	}

	h.broadcastMessage(statusMsg)
	log.Printf("🟢 Statut diffusé: User %d - %s", userID, msgType)
}

// GetOnlineUsers retourne la liste des utilisateurs connectés
func (h *MessageHub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()

	userIDs := make([]uint, 0, len(h.clients))
	for userID := range h.clients {
		userIDs = append(userIDs, userID)
	}
	return userIDs
}

// IsUserOnline vérifie si un utilisateur est connecté
func (h *MessageHub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, ok := h.clients[userID]
	return ok
}

// SendToUser envoie un message à un utilisateur spécifique
func (h *MessageHub) SendToUser(userID uint, message *models.WebSocketMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if client, ok := h.clients[userID]; ok {
		messageJSON, err := json.Marshal(message)
		if err != nil {
			log.Printf("❌ Erreur marshalling message: %v", err)
			return
		}

		select {
		case client.Send <- messageJSON:
		default:
			close(client.Send)
			delete(h.clients, client.UserID)
		}
	}
}

// readPump pompe les messages du websocket vers le hub
func (c *MessageClient) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ Erreur WebSocket: %v", err)
			}
			break
		}

		// Parser le message
		var wsMsg models.WebSocketMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("❌ Erreur parsing message WebSocket: %v", err)
			continue
		}

		log.Printf("📥 Message WebSocket reçu: Type=%s", wsMsg.Type)

		// Traiter selon le type
		switch wsMsg.Type {
		case "message":
			if wsMsg.Message != nil {
				wsMsg.Message.SenderID = c.UserID
				wsMsg.Timestamp = time.Now()
				c.Hub.broadcast <- &wsMsg
			}
		case "typing":
			if wsMsg.Data == nil {
				wsMsg.Data = map[string]interface{}{}
			}
			wsMsg.Data["user_id"] = c.UserID
			wsMsg.Timestamp = time.Now()
			c.Hub.broadcast <- &wsMsg
		}
	}
}

// writePump pompe les messages du hub vers le websocket
func (c *MessageClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Ajouter les messages en attente
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// MessagesWebSocketHandler gère les connexions WebSocket pour les messages
func MessagesWebSocketHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		log.Printf("❌ WebSocket Messages: Utilisateur non authentifié")
		http.Error(w, "Non authentifié", http.StatusUnauthorized)
		return
	}

	log.Printf("🔌 WebSocket Messages: Tentative de connexion User ID %d", user.ID)

	// Upgrade vers WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("❌ Erreur upgrade WebSocket: %v", err)
		return
	}

	hub := GetMessageHub()
	client := &MessageClient{
		Conn:   conn,
		Send:   make(chan []byte, 256),
		Hub:    hub,
		UserID: user.ID,
	}

	client.Hub.register <- client

	// Démarrer les goroutines
	go client.writePump()
	go client.readPump()

	log.Printf("✅ WebSocket Messages: Client %d connecté avec succès", user.ID)
}
