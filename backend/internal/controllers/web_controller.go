package controllers

import (
	"html/template"
	"log"
	"net/http"

	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"
)

type WebController struct {
	Templates     *template.Template
	threadService services.ThreadService
}

type IndexData struct {
	Title   string
	Threads []services.ThreadDTO
	Error   string
}

// NewWebController creates a new web controller
func NewWebController(templates *template.Template) *WebController {
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	return &WebController{
		Templates:     templates,
		threadService: threadService,
	}
}

func (wc *WebController) IndexHandler(w http.ResponseWriter, r *http.Request) {
	data := IndexData{
		Title: "Rythm'it - Forum Musical",
	}

	// Récupérer les threads via le service
	threads, err := wc.threadService.GetAllThreads()
	if err != nil {
		log.Printf("Erreur lors de la récupération des threads: %v", err)
		data.Error = "Impossible de charger les discussions"
		data.Threads = []services.ThreadDTO{}
	} else {
		data.Threads = threads
	}

	// Afficher le template
	err = wc.Templates.ExecuteTemplate(w, "index.html", data)
	if err != nil {
		log.Printf("Erreur template: %v", err)
		http.Error(w, "Erreur interne du serveur", http.StatusInternalServerError)
		return
	}
}
