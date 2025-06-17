// Fichier: backend/internal/handlers/notification_handler.go
package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

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
