package controllers

import (
	// "encoding/json"
	"net/http"

	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"
)

// ThreadController gère les requêtes liées aux threads
type ThreadController struct {
	ThreadRepo repositories.ThreadRepository
}

// NewThreadController crée une nouvelle instance de ThreadController
func NewThreadController(threadRepo repositories.ThreadRepository) *ThreadController {
	return &ThreadController{
		ThreadRepo: threadRepo,
	}
}

// GetPublicThreads gère la récupération des threads publics
func (c *ThreadController) GetPublicThreads(w http.ResponseWriter, r *http.Request) {
	// TODO: Récupérer les paramètres de pagination depuis la requête si nécessaire
	// Pour l'instant, utilisons des valeurs par défaut
	params := models.PaginationParams{
		Page:    1,
		PerPage: 10, // Nombre de threads à afficher par défaut
		Sort:    "created_at",
		Order:   "desc",
	}

	threads, total, err := c.ThreadRepo.FindPublicThreads(params)
	if err != nil {
		// Utiliser utils.InternalServerError ou utils.Error avec la signature correcte
		utils.InternalServerError(w, "Erreur lors de la récupération des threads")
		return
	}

	// Utiliser utils.Paginated pour la réponse paginée
	utils.Paginated(w, threads, params.Page, params.PerPage, total)
}

// TODO: Ajouter d'autres handlers pour les threads (création, mise à jour, suppression, etc.)
