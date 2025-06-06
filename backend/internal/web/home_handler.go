package web

import (
	"net/http"

	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
)

// HomeData structure pour les données de la page d'accueil
type HomeData struct {
	PageData
	Threads []*models.Thread
	Battles []*models.Battle // Pourra être utilisé plus tard pour afficher les battles
}

// HomeHandler gère la page d'accueil
type HomeHandler struct {
	threadRepo repositories.ThreadRepository
	battleRepo repositories.BattleRepository // Pourra être utilisé plus tard
	tmpl       *TemplateManager
}

// NewHomeHandler crée une nouvelle instance de HomeHandler
func NewHomeHandler(threadRepo repositories.ThreadRepository, battleRepo repositories.BattleRepository, tmpl *TemplateManager) *HomeHandler {
	return &HomeHandler{
		threadRepo: threadRepo,
		battleRepo: battleRepo,
		tmpl:       tmpl,
	}
}

// ServeHTTP implémente l'interface http.Handler pour HomeHandler
func (h *HomeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Récupérer les derniers threads publics
	// Utilisons des paramètres de pagination par défaut pour l'instant
	threadParams := models.PaginationParams{
		Page:    1,
		PerPage: 10,
		Sort:    "created_at",
		Order:   "desc",
	}

	threads, _, err := h.threadRepo.FindPublicThreads(threadParams)
	if err != nil {
		// En cas d'erreur, on peut quand même essayer de rendre la page avec un message d'erreur
		// ou retourner une erreur interne du serveur.
		// Pour l'instant, affichons une erreur interne et loguons.
		// log.Printf("erreur récupération threads pour page d\'accueil: %v", err) // Loguer l'erreur
		http.Error(w, "Erreur lors de la récupération des threads", http.StatusInternalServerError)
		return
	}

	// Récupérer les battles actives
	// Supposons qu'on veut afficher les 3 battles les plus récentes ou les plus actives
	// La méthode FindActive pourrait prendre un paramètre pour le nombre
	battles, err := h.battleRepo.FindActive(5) // Utilisons 5 comme exemple pour le nombre de battles actives
	if err != nil {
		// En cas d'erreur, on peut quand même essayer de rendre la page avec un message d'erreur
		// log.Printf("erreur récupération battles pour page d\'accueil: %v", err) // Loguer l'erreur
		http.Error(w, "Erreur lors de la récupération des battles", http.StatusInternalServerError)
		return
	}

	// Préparer les données pour le template
	data := HomeData{
		PageData: BasePageData("Accueil"),
		Threads:  threads,
		Battles:  battles, // Inclure les battles récupérées ici
	}

	// Rendre le template home.html avec les données
	if err := h.tmpl.RenderTemplate(w, "home.html", data); err != nil {
		// Le RenderTemplate gère déjà l'erreur, mais on peut loguer ici si nécessaire
		// log.Printf("erreur rendu template page d\'accueil: %v", err) // Loguer l'erreur
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
	}
}

// TODO: Ajouter d'autres handlers si nécessaire pour les pages web
