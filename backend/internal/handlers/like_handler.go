package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"

	"github.com/gorilla/mux"
)

// LikeResponse structure pour la réponse des likes
type LikeResponse struct {
	Success    bool   `json:"success"`
	Liked      bool   `json:"liked"`
	LikesCount int    `json:"likes_count"`
	Message    string `json:"message,omitempty"`
}

// ToggleLikeHandler gère le like/unlike d'un thread
func ToggleLikeHandler(w http.ResponseWriter, r *http.Request) {
	// Récupérer l'utilisateur depuis le contexte (injecté par le middleware)
	userID, ok := r.Context().Value("user_id").(uint)
	if !ok || userID == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(LikeResponse{
			Success: false,
			Message: "Authentification requise",
		})
		return
	}

	// Récupérer les infos utilisateur pour les logs
	var username string
	if user, ok := r.Context().Value("user").(*services.UserResponseDTO); ok {
		username = user.Username
	} else {
		username = "unknown"
	}

	// Récupérer l'ID du thread depuis l'URL
	vars := mux.Vars(r)
	threadIDStr := vars["id"]
	threadID, err := strconv.ParseUint(threadIDStr, 10, 32)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(LikeResponse{
			Success: false,
			Message: "ID thread invalide",
		})
		return
	}

	log.Printf("🎯 ToggleLikeHandler - Thread: %d, User: %s (ID: %d)", threadID, username, userID)

	// Créer le repository des likes
	db := database.DB
	likeRepo := repositories.NewLikeRepository(db)

	// Vérifier si l'utilisateur a déjà liké ce thread
	currentlyLiked, err := likeRepo.IsThreadLikedByUser(userID, uint(threadID))
	if err != nil {
		log.Printf("❌ Erreur vérification like: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(LikeResponse{
			Success: false,
			Message: "Erreur interne du serveur",
		})
		return
	}

	// Toggle du like
	var newLikedState bool
	if currentlyLiked {
		// Unlike
		err = likeRepo.UnlikeThread(userID, uint(threadID))
		if err != nil {
			log.Printf("❌ Erreur unlike thread: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LikeResponse{
				Success: false,
				Message: "Erreur lors du retrait du like",
			})
			return
		}
		newLikedState = false
		log.Printf("👎 Thread %d unliké par %s", threadID, username)
	} else {
		// Like
		err = likeRepo.LikeThread(userID, uint(threadID))
		if err != nil {
			log.Printf("❌ Erreur like thread: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(LikeResponse{
				Success: false,
				Message: "Erreur lors de l'ajout du like",
			})
			return
		}
		newLikedState = true
		log.Printf("👍 Thread %d liké par %s", threadID, username)
	}

	// Récupérer le nouveau nombre de likes
	likesCount, err := likeRepo.GetThreadLikesCount(uint(threadID))
	if err != nil {
		log.Printf("❌ Erreur récupération compteur likes: %v", err)
		// Continuer même si on ne peut pas récupérer le compteur
		likesCount = 0
	}

	// Réponse succès
	response := LikeResponse{
		Success:    true,
		Liked:      newLikedState,
		LikesCount: likesCount,
		Message:    "Like mis à jour avec succès",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ Like togglé avec succès - Thread: %d, Liked: %t, Count: %d", threadID, newLikedState, likesCount)
}
