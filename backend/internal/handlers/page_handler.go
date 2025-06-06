// Fichier: backend/internal/handlers/page_handlers.go
package handlers

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"
	"strings"
	"time"
)

// Variables globales pour les templates
var templates *template.Template

// InitTemplates initialise les templates (à appeler au démarrage)
func InitTemplates() error {
	var err error

	// Chemin vers les templates
	templatePath := "../frontend/*.html"
	log.Printf("🔍 Chargement des templates depuis: %s", templatePath)

	// Vérifier si les fichiers existent
	matches, err := filepath.Glob(templatePath)
	if err != nil {
		return err
	}

	log.Printf("📂 Templates trouvés: %v", matches)

	// Charger les templates
	templates, err = template.ParseGlob(templatePath)
	if err != nil {
		return err
	}

	// Lister les templates chargés
	if templates != nil {
		for _, tmpl := range templates.Templates() {
			log.Printf("✅ Template chargé: %s", tmpl.Name())
		}
	}

	return nil
}

// PageData structure de données pour les templates
type PageData struct {
	Title       string
	User        *User
	IsLoggedIn  bool
	Friends     []Friend
	Threads     []Thread
	Messages    []Message
	Trends      []Trend
	CurrentPage string
}

// User structure pour l'utilisateur connecté
type User struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	IsAdmin  bool   `json:"is_admin"`
	Avatar   string `json:"avatar"`
}

// Friend structure pour les amis
type Friend struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Avatar   string `json:"avatar"`
	Status   string `json:"status"` // online, away, offline
	Activity string `json:"activity"`
}

// Thread structure pour les discussions
type Thread struct {
	ID           uint        `json:"id"`
	Title        string      `json:"title"`
	Content      string      `json:"content"`
	Author       string      `json:"author"`
	AuthorAvatar string      `json:"author_avatar"`
	TimeAgo      string      `json:"time_ago"`
	Genre        string      `json:"genre"`
	Likes        int         `json:"likes"`
	Comments     int         `json:"comments"`
	Shares       int         `json:"shares"`
	MusicTrack   *MusicTrack `json:"music_track,omitempty"`
}

// MusicTrack structure pour les pistes musicales
type MusicTrack struct {
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Duration string `json:"duration"`
	CoverURL string `json:"cover_url"`
}

// Message structure pour les messages
type Message struct {
	ID      uint   `json:"id"`
	From    string `json:"from"`
	Content string `json:"content"`
	TimeAgo string `json:"time_ago"`
	Unread  bool   `json:"unread"`
}

// Trend structure pour les tendances
type Trend struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Discussions int    `json:"discussions"`
	CoverURL    string `json:"cover_url"`
}

// IndexHandler gère la page d'accueil
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🏠 IndexHandler appelé pour: %s", r.URL.Path)

	// Créer le service pour récupérer les threads de la DB
	threadsFromDB, err := getThreadsFromDatabase()
	var threads []Thread

	if err != nil {
		log.Printf("❌ Erreur récupération threads DB: %v", err)
		threads = []Thread{} // Liste vide en cas d'erreur
	} else {
		// Convertir les threads de la DB au format attendu par le template
		threads = convertDBThreadsToPageThreads(threadsFromDB)
		log.Printf("✅ %d threads récupérés de la DB", len(threads))
	}

	data := PageData{
		Title:       "Accueil - Rythm'it",
		CurrentPage: "index",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
		Friends: []Friend{
			{ID: 1, Username: "MixMaster", Avatar: "MX", Status: "online", Activity: "Écoute: Techno Vibes"},
			{ID: 2, Username: "SoundBliss", Avatar: "SB", Status: "online", Activity: "Écoute: Jazz Evening"},
			{ID: 3, Username: "VibeWave", Avatar: "VW", Status: "away", Activity: "Absent - il y a 2h"},
		},
		Threads: threads, // Utiliser les threads de la DB
		Messages: []Message{
			{ID: 1, From: "MixMaster", Content: "Salut ! Tu as écouté le dernier album de...", TimeAgo: "il y a 5min", Unread: true},
			{ID: 2, From: "SoundBliss", Content: "Cette playlist jazz est incroyable ! 🎷", TimeAgo: "il y a 1h", Unread: true},
			{ID: 3, From: "RhythmHunter", Content: "On fait une session d'écoute ce soir ?", TimeAgo: "il y a 3h", Unread: false},
		},
		Trends: []Trend{
			{Name: "Synthwave Summer", Category: "Electronic", Discussions: 1200},
			{Name: "Indie Folk Revival", Category: "Folk", Discussions: 890},
			{Name: "Techno Underground", Category: "Electronic", Discussions: 654},
		},
	}

	log.Printf("📊 Données préparées: %d amis, %d threads, %d messages", len(data.Friends), len(data.Threads), len(data.Messages))

	// Vérifier que les templates sont chargés
	if templates == nil {
		log.Printf("❌ Templates non chargés!")
		http.Error(w, "Templates non chargés", http.StatusInternalServerError)
		return
	}

	// Lister les templates disponibles
	log.Printf("📋 Templates disponibles:")
	for _, tmpl := range templates.Templates() {
		log.Printf("   - %s", tmpl.Name())
	}

	// Essayer de rendre le template
	log.Printf("🎨 Rendu du template index...")
	err = templates.ExecuteTemplate(w, "index", data)
	if err != nil {
		log.Printf("❌ Erreur rendu template index: %v", err)
		http.Error(w, fmt.Sprintf("Erreur template: %v", err), http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Template rendu avec succès")
}

// DiscoverHandler gère la page de découverte
func DiscoverHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔍 DiscoverHandler appelé")
	data := PageData{
		Title:       "Découvrir - Rythm'it",
		CurrentPage: "discover",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
	}

	renderTemplate(w, "discover.html", data)
}

// FriendsHandler gère la page des amis
func FriendsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("👥 FriendsHandler appelé")
	data := PageData{
		Title:       "Mes Amis - Rythm'it",
		CurrentPage: "friends",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
		Friends: []Friend{
			{ID: 1, Username: "MixMaster", Avatar: "MX", Status: "online", Activity: "Écoute: Techno Vibes Mix"},
			{ID: 2, Username: "SoundBliss", Avatar: "SB", Status: "online", Activity: "Écoute: Jazz Evening Collection"},
			{ID: 3, Username: "VibeWave", Avatar: "VW", Status: "away", Activity: "Absent - il y a 2h"},
			{ID: 4, Username: "RhythmHunter", Avatar: "RH", Status: "online", Activity: "En ligne"},
			{ID: 5, Username: "EchoBeat", Avatar: "EB", Status: "online", Activity: "Écoute: Synthwave Dreams"},
			{ID: 6, Username: "DeepSounds", Avatar: "DS", Status: "offline", Activity: "Hors ligne - il y a 1j"},
		},
	}

	renderTemplate(w, "friends.html", data)
}

// MessagesHandler gère la page des messages
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("💬 MessagesHandler appelé")
	data := PageData{
		Title:       "Messages - Rythm'it",
		CurrentPage: "messages",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
	}

	renderTemplate(w, "messages.html", data)
}

// ProfileHandler gère la page de profil
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("👤 ProfileHandler appelé")
	data := PageData{
		Title:       "Mon Profil - Rythm'it",
		CurrentPage: "profile",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
	}

	renderTemplate(w, "profile.html", data)
}

// SettingsHandler gère la page des paramètres
func SettingsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("⚙️ SettingsHandler appelé")
	data := PageData{
		Title:       "Paramètres - Rythm'it",
		CurrentPage: "settings",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
	}

	renderTemplate(w, "settings.html", data)
}

// HubHandler gère la page hub
func HubHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🎯 HubHandler appelé")
	data := PageData{
		Title:       "Hub - Rythm'it",
		CurrentPage: "hub",
		IsLoggedIn:  true,
		User: &User{
			ID:       1,
			Username: "MonProfil",
			Avatar:   "MO",
		},
	}

	renderTemplate(w, "hub.html", data)
}

// SigninHandler gère la page de connexion
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔑 SigninHandler appelé")
	data := PageData{
		Title:       "Connexion - Rythm'it",
		CurrentPage: "signin",
		IsLoggedIn:  false,
	}

	renderTemplate(w, "signin.html", data)
}

// SignupHandler gère la page d'inscription
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📝 SignupHandler appelé")
	data := PageData{
		Title:       "Inscription - Rythm'it",
		CurrentPage: "signup",
		IsLoggedIn:  false,
	}

	renderTemplate(w, "signup.html", data)
}

// ProfileAPIHandler gère l'API du profil (pour le JavaScript)
func ProfileAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔌 ProfileAPIHandler appelé")
	// Pour l'instant, retourner un profil fictif
	profileData := map[string]interface{}{
		"id":               1,
		"username":         "MonProfil",
		"email":            "mon.email@rythmit.com",
		"is_admin":         false,
		"message_count":    0,
		"thread_count":     0,
		"favorite_genres":  []string{"electronic", "ambient"},
		"favorite_artists": []string{"deadmau5", "brian eno"},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"message": "Profil récupéré avec succès",
		"data":    profileData,
	}

	json.NewEncoder(w).Encode(response)
}

// renderTemplate fonction helper pour rendre les templates
func renderTemplate(w http.ResponseWriter, templateName string, data PageData) {
	log.Printf("🎨 Rendu template: %s", templateName)

	if templates == nil {
		log.Printf("❌ Templates non chargés!")
		http.Error(w, "Templates non disponibles", http.StatusInternalServerError)
		return
	}

	err := templates.ExecuteTemplate(w, templateName, data)
	if err != nil {
		log.Printf("❌ Erreur rendu template %s: %v", templateName, err)

		// Essayer sans l'extension
		baseName := templateName[:len(templateName)-5] // enlever .html
		err2 := templates.ExecuteTemplate(w, baseName, data)
		if err2 != nil {
			log.Printf("❌ Erreur rendu template %s: %v", baseName, err2)
			http.Error(w, fmt.Sprintf("Erreur template: %v", err), http.StatusInternalServerError)
			return
		}
	}

	log.Printf("✅ Template %s rendu avec succès", templateName)
}

// getThreadsFromDatabase récupère les threads depuis la base de données
func getThreadsFromDatabase() ([]services.ThreadDTO, error) {
	// Créer les dépendances
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	// Récupérer les threads
	return threadService.GetAllThreads()
}

// convertDBThreadsToPageThreads convertit les threads de la DB au format attendu par le template
func convertDBThreadsToPageThreads(dbThreads []services.ThreadDTO) []Thread {
	var pageThreads []Thread

	for _, dbThread := range dbThreads {
		// Debug: afficher les données du thread
		log.Printf("🔍 Thread ID=%d, Title='%s', Content='%s', User='%s'",
			dbThread.ID, dbThread.Title, dbThread.Content, dbThread.Username)

		// Générer les initiales pour l'avatar
		initials := generateInitials(dbThread.Username)

		// Formater le temps
		timeAgo := formatTimeAgo(dbThread.CreatedAt)

		// Déterminer le genre principal (premier tag ou "Général")
		genre := "Général"
		if len(dbThread.Tags) > 0 {
			genre = strings.Title(dbThread.Tags[0])
		}

		// Créer le thread pour le template
		pageThread := Thread{
			ID:           dbThread.ID,
			Title:        dbThread.Title,
			Content:      dbThread.Content,
			Author:       dbThread.Username,
			AuthorAvatar: initials,
			TimeAgo:      timeAgo,
			Genre:        genre,
			Likes:        0, // TODO: implémenter les likes
			Comments:     dbThread.MessageCount,
			Shares:       0,   // TODO: implémenter les partages
			MusicTrack:   nil, // Pas de piste musicale pour l'instant
		}

		log.Printf("✅ Thread converti: Title='%s', Content='%s'", pageThread.Title, pageThread.Content)
		pageThreads = append(pageThreads, pageThread)
	}

	return pageThreads
}

// generateInitials génère les initiales à partir d'un nom d'utilisateur
func generateInitials(username string) string {
	if len(username) == 0 {
		return "?"
	}
	// Prendre les deux premières lettres en majuscules
	runes := []rune(username)
	if len(runes) >= 2 {
		return strings.ToUpper(string(runes[0])) + strings.ToUpper(string(runes[1]))
	}
	return strings.ToUpper(string(runes[0]))
}

// formatTimeAgo formate une date en "il y a X [unité]"
func formatTimeAgo(t time.Time) string {
	diff := time.Since(t)

	if diff.Hours() < 24 && time.Now().Day() == t.Day() {
		if diff.Minutes() < 1 {
			return "À l'instant"
		} else if diff.Minutes() < 60 {
			return fmt.Sprintf("il y a %.0f min", diff.Minutes())
		} else {
			return fmt.Sprintf("il y a %.0f h", diff.Hours())
		}
	} else if diff.Hours() < 24*7 && time.Now().Year() == t.Year() {
		return fmt.Sprintf("il y a %.0f j", diff.Hours()/24)
	} else if diff.Hours() < 24*30 {
		return fmt.Sprintf("il y a %.0f sem", diff.Hours()/24/7)
	} else if diff.Hours() < 24*365 {
		return fmt.Sprintf("il y a %.0f mois", diff.Hours()/24/30)
	}
	return fmt.Sprintf("il y a %.0f ans", diff.Hours()/24/365)
}
