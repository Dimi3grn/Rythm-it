package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"

	"github.com/gorilla/mux"
)

// MessageAPIHandler gère les API des messages
func ThreadMessagesHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'ID du thread depuis l'URL
	vars := mux.Vars(r)
	threadIDStr := vars["id"]
	threadID, err := strconv.ParseUint(threadIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID thread invalide", http.StatusBadRequest)
		return
	}

	switch r.Method {
	case "GET":
		getThreadMessages(w, r, uint(threadID))
	case "POST":
		createThreadMessage(w, r, uint(threadID))
	default:
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// getThreadMessages récupère les messages d'un thread
func getThreadMessages(w http.ResponseWriter, r *http.Request, threadID uint) {
	// Récupérer l'utilisateur connecté (optionnel)
	var userID *uint
	if userIDVal, ok := r.Context().Value("user_id").(uint); ok && userIDVal != 0 {
		userID = &userIDVal
	}

	// Paramètres de pagination
	page := 1
	if pageStr := r.URL.Query().Get("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPage := 20
	if perPageStr := r.URL.Query().Get("per_page"); perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 100 {
			perPage = pp
		}
	}

	sortBy := r.URL.Query().Get("sort")
	if sortBy == "" {
		sortBy = "date"
	}

	// Créer le service
	db := database.DB
	messageRepo := repositories.NewMessageRepository(db)
	threadRepo := repositories.NewThreadRepository(db)
	messageService := services.NewMessageService(messageRepo, threadRepo, db)

	// Récupérer les messages
	params := models.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "created_at",
		Order:   "ASC",
	}

	response, err := messageService.GetThreadMessages(threadID, userID, params, sortBy)
	if err != nil {
		log.Printf("❌ Erreur récupération messages thread %d: %v", threadID, err)
		http.Error(w, "Erreur récupération messages", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
	})
}

// createThreadMessage crée un nouveau message dans un thread
func createThreadMessage(w http.ResponseWriter, r *http.Request, threadID uint) {
	// Vérifier l'authentification
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	// Parser le body JSON
	var dto services.CreateMessageDTO
	if err := json.NewDecoder(r.Body).Decode(&dto); err != nil {
		http.Error(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	// Créer le service
	db := database.DB
	messageRepo := repositories.NewMessageRepository(db)
	threadRepo := repositories.NewThreadRepository(db)
	messageService := services.NewMessageService(messageRepo, threadRepo, db)

	// Créer le message
	response, err := messageService.PostMessage(dto, threadID, userID)
	if err != nil {
		log.Printf("❌ Erreur création message: %v", err)

		// Gestion des erreurs spécifiques
		if err.Error() == "thread not found" {
			http.Error(w, "Thread non trouvé", http.StatusNotFound)
			return
		}
		if err.Error() == "unauthorized" {
			http.Error(w, "Non autorisé", http.StatusForbidden)
			return
		}

		http.Error(w, "Erreur création message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "Message créé avec succès",
	})

	log.Printf("✅ Message créé dans thread %d par user %d", threadID, userID)
}

// MessageVoteHandler gère les votes sur les messages
func MessageVoteHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du message depuis l'URL
	vars := mux.Vars(r)
	messageIDStr := vars["id"]
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID message invalide", http.StatusBadRequest)
		return
	}

	// Parser le body JSON
	var voteRequest struct {
		Vote string `json:"vote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&voteRequest); err != nil {
		http.Error(w, "Format JSON invalide", http.StatusBadRequest)
		return
	}

	// Créer le service
	db := database.DB
	messageRepo := repositories.NewMessageRepository(db)
	threadRepo := repositories.NewThreadRepository(db)
	messageService := services.NewMessageService(messageRepo, threadRepo, db)

	// Voter sur le message
	response, err := messageService.VoteMessage(uint(messageID), userID, voteRequest.Vote)
	if err != nil {
		log.Printf("❌ Erreur vote message %d: %v", messageID, err)

		// Gestion des erreurs spécifiques
		if err.Error() == "message not found" {
			http.Error(w, "Message non trouvé", http.StatusNotFound)
			return
		}
		if err.Error() == "vous ne pouvez pas voter pour votre propre message" {
			http.Error(w, "Vous ne pouvez pas voter pour votre propre message", http.StatusForbidden)
			return
		}

		http.Error(w, "Erreur lors du vote", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    response,
		"message": "Vote enregistré avec succès",
	})

	log.Printf("✅ Vote %s enregistré sur message %d par user %d", voteRequest.Vote, messageID, userID)
}

// MessageLikeHandler gère les likes simples sur les messages
func MessageLikeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok || userID == 0 {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du message depuis l'URL
	vars := mux.Vars(r)
	messageIDStr := vars["id"]
	messageID, err := strconv.ParseUint(messageIDStr, 10, 32)
	if err != nil {
		http.Error(w, "ID message invalide", http.StatusBadRequest)
		return
	}

	// Créer les services
	db := database.DB
	messageRepo := repositories.NewMessageRepository(db)

	// Vérifier que le message existe
	_, err = messageRepo.FindByID(uint(messageID))
	if err != nil {
		log.Printf("❌ Message %d non trouvé: %v", messageID, err)
		http.Error(w, "Message non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier si l'utilisateur a déjà liké ce message
	var isCurrentlyLiked bool
	checkQuery := "SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND message_id = ?"
	var count int
	err = db.QueryRow(checkQuery, userID, messageID).Scan(&count)
	if err != nil {
		log.Printf("❌ Erreur vérification like: %v", err)
		http.Error(w, "Erreur interne", http.StatusInternalServerError)
		return
	}
	isCurrentlyLiked = count > 0

	// Toggle du like
	var newLikedState bool
	if isCurrentlyLiked {
		// Supprimer le like
		deleteQuery := "DELETE FROM comment_likes WHERE user_id = ? AND message_id = ?"
		_, err = db.Exec(deleteQuery, userID, messageID)
		if err != nil {
			log.Printf("❌ Erreur suppression like: %v", err)
			http.Error(w, "Erreur suppression like", http.StatusInternalServerError)
			return
		}
		newLikedState = false
		log.Printf("👎 Like supprimé sur message %d par user %d", messageID, userID)
	} else {
		// Ajouter le like
		insertQuery := "INSERT INTO comment_likes (user_id, message_id) VALUES (?, ?)"
		_, err = db.Exec(insertQuery, userID, messageID)
		if err != nil {
			log.Printf("❌ Erreur ajout like: %v", err)
			http.Error(w, "Erreur ajout like", http.StatusInternalServerError)
			return
		}
		newLikedState = true
		log.Printf("👍 Like ajouté sur message %d par user %d", messageID, userID)
	}

	// Compter le total de likes pour ce message
	countQuery := "SELECT COUNT(*) FROM comment_likes WHERE message_id = ?"
	var totalLikes int
	err = db.QueryRow(countQuery, messageID).Scan(&totalLikes)
	if err != nil {
		log.Printf("❌ Erreur comptage likes: %v", err)
		totalLikes = 0
	}

	// Retourner la réponse
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message_id":  messageID,
			"is_liked":    newLikedState,
			"likes_count": totalLikes,
		},
		"message": "Like mis à jour avec succès",
	})
}
