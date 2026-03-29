// Fichier: backend/internal/handlers/page_handlers.go
package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"rythmitbackend/configs"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/mux"
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
	Profile     *ProfileData // Données du profil personnalisé
	// Données pour la page thread
	Thread   *Thread   `json:"thread,omitempty"`
	Comments []Comment `json:"comments,omitempty"`
	// Données pour l'authentification
	IsSignupMode   bool
	ErrorMessage   string
	SuccessMessage string
	AuthTitle      string
	AuthSubtitle   string
	AuthButtonText string
	AuthFooterText template.HTML
	// Données pour le profil d'un autre utilisateur
	IsOwnProfile     bool
	CurrentUser      *User   // Utilisateur connecté (différent de User si on visite un autre profil)
	FriendshipStatus *string // Statut d'amitié avec l'utilisateur affiché
}

// ProfileData structure pour les données de profil personnalisé
type ProfileData struct {
	DisplayName *string `json:"display_name"`
	AvatarImage *string `json:"avatar_image"`
	BannerImage *string `json:"banner_image"`
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
	ImageURL     *string     `json:"image_url,omitempty"`
	Author       string      `json:"author"`
	AuthorAvatar string      `json:"author_avatar"`
	TimeAgo      string      `json:"time_ago"`
	Genre        string      `json:"genre"`
	Tags         []string    `json:"tags"`
	Likes        int         `json:"likes"`
	IsLiked      bool        `json:"is_liked"`
	Comments     int         `json:"comments"`
	Shares       int         `json:"shares"`
	Visibility   string      `json:"visibility"`
	State        string      `json:"state"`
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

// Comment structure pour les commentaires de threads
type Comment struct {
	ID           uint      `json:"id"`
	Content      string    `json:"content"`
	ImageURL     *string   `json:"image_url,omitempty"`
	Author       string    `json:"author"`
	AuthorAvatar string    `json:"author_avatar"`
	TimeAgo      string    `json:"time_ago"`
	Likes        int       `json:"likes"`    // Nombre de likes
	IsLiked      bool      `json:"is_liked"` // Utilisateur a liké
	IsOP         bool      `json:"is_op"`    // Original Poster
	Replies      []Comment `json:"replies,omitempty"`
}

// Trend structure pour les tendances
type Trend struct {
	Name        string `json:"name"`
	Category    string `json:"category"`
	Discussions int    `json:"discussions"`
	CoverURL    string `json:"cover_url"`
}

// Tag structure pour les tags disponibles
type Tag struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// PostHandler gère les actions sur les posts (like, comment, etc.)
func PostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Parser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de parsing du formulaire", http.StatusBadRequest)
		return
	}

	action := r.FormValue("action")
	threadID := r.FormValue("thread_id")

	log.Printf("🎯 PostHandler - Action: %s, Thread: %s, User: %s", action, threadID, user.Username)

	switch action {
	case "like":
		handleLike(w, r, threadID, user)
	case "comment":
		handleComment(w, r, threadID, user)
	case "share":
		handleShare(w, r, threadID, user)
	case "create_post":
		handleCreatePost(w, r, user)
	default:
		http.Error(w, "Action non supportée", http.StatusBadRequest)
		return
	}

	// Rediriger vers la page d'accueil après l'action
	http.Redirect(w, r, "/?success=action_completed", http.StatusSeeOther)
}

// NewPostHandler gère la création de nouveaux posts
func NewPostHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Parser le formulaire
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Erreur de parsing du formulaire", http.StatusBadRequest)
		return
	}

	content := r.FormValue("content")
	title := r.FormValue("title")        // Nouveau champ pour le titre
	tags := r.FormValue("tags")          // Nouveau champ pour les tags
	imageURL := r.FormValue("image_url") // Nouveau champ pour l'image

	// DEBUG: Afficher les valeurs reçues
	log.Printf("🔍 DEBUG - Valeurs reçues du formulaire:")
	log.Printf("  - content: %s", content)
	log.Printf("  - title: %s", title)
	log.Printf("  - tags: %s", tags)
	log.Printf("  - image_url: %s", imageURL)

	// Validation des champs obligatoires
	if strings.TrimSpace(content) == "" {
		http.Redirect(w, r, "/?error=empty_content", http.StatusSeeOther)
		return
	}

	// Générer un titre automatiquement si pas fourni
	if strings.TrimSpace(title) == "" {
		title = generateTitleFromContent(content)
	}

	// Traiter les tags
	var tagList []string
	if strings.TrimSpace(tags) != "" {
		// Séparer les tags par virgule et nettoyer
		tagList = strings.Split(tags, ",")
		for i, tag := range tagList {
			tagList[i] = strings.TrimSpace(tag)
		}
	} else {
		// Tags par défaut si aucun fourni
		tagList = []string{"général", "discussion"}
	}

	log.Printf("📝 Nouveau thread de %s: titre='%s', contenu='%s', tags=%v", user.Username, title, content, tagList)

	// Créer le service pour sauvegarder en base de données
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	// Créer le DTO pour le nouveau thread
	createThreadDTO := services.CreateThreadDTO{
		Title:       title,
		Description: content,
		Tags:        tagList,
		Visibility:  "public", // Par défaut public
	}

	// Ajouter l'image si elle est fournie
	if strings.TrimSpace(imageURL) != "" {
		createThreadDTO.ImageURL = &imageURL
	}

	// Créer le thread
	createdThread, err := threadService.CreateThread(createThreadDTO, user.ID)
	if err != nil {
		log.Printf("❌ Erreur création thread: %v", err)
		http.Redirect(w, r, "/?error=thread_creation_failed", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Thread créé avec succès: ID=%d, titre='%s'", createdThread.ID, createdThread.Title)

	// Rediriger avec succès
	http.Redirect(w, r, "/?success=thread_created", http.StatusSeeOther)
}

// generateTitleFromContent génère automatiquement un titre à partir du contenu
func generateTitleFromContent(content string) string {
	words := strings.Fields(content)
	if len(words) == 0 {
		return "Nouvelle discussion"
	}

	// Prendre les premiers mots (max 8 mots ou 50 caractères)
	var title strings.Builder
	totalLength := 0
	for i, word := range words {
		if i >= 8 || totalLength+len(word) > 50 {
			break
		}
		if i > 0 {
			title.WriteString(" ")
			totalLength++
		}
		title.WriteString(word)
		totalLength += len(word)
	}

	result := title.String()
	if len(words) > 8 || totalLength >= 47 {
		result += "..."
	}

	return result
}

// Handlers d'actions spécifiques
func handleLike(w http.ResponseWriter, r *http.Request, threadID string, user *User) {
	log.Printf("❤️ Like - Thread: %s, User: %s", threadID, user.Username)
	// TODO: Implémenter le like en base de données
}

func handleComment(w http.ResponseWriter, r *http.Request, threadID string, user *User) {
	comment := r.FormValue("comment")
	log.Printf("💬 Comment - Thread: %s, User: %s, Comment: %s", threadID, user.Username, comment)
	// TODO: Sauvegarder le commentaire en base de données
}

func handleShare(w http.ResponseWriter, r *http.Request, threadID string, user *User) {
	log.Printf("🔄 Share - Thread: %s, User: %s", threadID, user.Username)
	// TODO: Implémenter le partage
}

func handleCreatePost(w http.ResponseWriter, r *http.Request, user *User) {
	content := r.FormValue("content")
	log.Printf("✍️ Create Post - User: %s, Content: %s", user.Username, content)
	// TODO: Créer le post en base de données
}

// IndexHandler gère la page d'accueil
func IndexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🏠 IndexHandler appelé pour: %s", r.URL.Path)

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)

	// Récupérer les messages d'erreur/succès depuis les query parameters
	errorParam := r.URL.Query().Get("error")
	successParam := r.URL.Query().Get("success")

	var errorMessage, successMessage string

	// Traduction des messages d'erreur
	switch errorParam {
	case "empty_content":
		errorMessage = "Le contenu ne peut pas être vide"
	case "action_failed":
		errorMessage = "L'action n'a pas pu être réalisée"
	case "thread_creation_failed":
		errorMessage = "Erreur lors de la création du thread. Veuillez réessayer."
	}

	// Traduction des messages de succès
	switch successParam {
	case "post_created":
		successMessage = "Votre post a été publié avec succès !"
	case "thread_created":
		successMessage = "Votre discussion a été créée avec succès !"
	case "thread_deleted":
		successMessage = "Le thread a été supprimé avec succès !"
	case "action_completed":
		successMessage = "Action réalisée avec succès !"
	}

	// Créer le service pour récupérer les threads de la DB
	threadsFromDB, err := getThreadsFromDatabase()
	var threads []Thread

	if err != nil {
		log.Printf("❌ Erreur récupération threads DB: %v", err)
		threads = []Thread{} // Liste vide en cas d'erreur
	} else {
		// Convertir les threads de la DB au format attendu par le template
		threads = convertDBThreadsToPageThreads(threadsFromDB, user)
		log.Printf("✅ %d threads récupérés de la DB", len(threads))
	}

	data := PageData{
		Title:          "Accueil - Rythm'it",
		CurrentPage:    "index",
		IsLoggedIn:     isLoggedIn,
		User:           user,
		Threads:        threads, // Utiliser les threads de la DB
		ErrorMessage:   errorMessage,
		SuccessMessage: successMessage,
		Trends: []Trend{
			{Name: "Synthwave Summer", Category: "Electronic", Discussions: 1200},
			{Name: "Indie Folk Revival", Category: "Folk", Discussions: 890},
			{Name: "Techno Underground", Category: "Electronic", Discussions: 654},
		},
	}

	// Ajouter des données supplémentaires si l'utilisateur est connecté
	if isLoggedIn {
		data.Friends = []Friend{
			{ID: 1, Username: "MixMaster", Avatar: "MX", Status: "online", Activity: "Écoute: Techno Vibes"},
			{ID: 2, Username: "SoundBliss", Avatar: "SB", Status: "online", Activity: "Écoute: Jazz Evening"},
			{ID: 3, Username: "VibeWave", Avatar: "VW", Status: "away", Activity: "Absent - il y a 2h"},
		}
		data.Messages = []Message{
			{ID: 1, From: "MixMaster", Content: "Salut ! Tu as écouté le dernier album de...", TimeAgo: "il y a 5min", Unread: true},
			{ID: 2, From: "SoundBliss", Content: "Cette playlist jazz est incroyable ! 🎷", TimeAgo: "il y a 1h", Unread: true},
			{ID: 3, From: "RhythmHunter", Content: "On fait une session d'écoute ce soir ?", TimeAgo: "il y a 3h", Unread: false},
		}
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

	// Récupérer l'utilisateur connecté depuis le cookie
	user, isLoggedIn := getUserFromCookie(r)

	data := PageData{
		Title:       "Mes Amis - Rythm'it",
		CurrentPage: "friends",
		IsLoggedIn:  isLoggedIn,
		User:        user,
		Friends:     []Friend{}, // Liste vide - chargée dynamiquement via JavaScript
	}

	renderTemplate(w, "friends.html", data)
}

// MessagesHandler gère la page des messages
func MessagesHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("💬 MessagesHandler appelé")

	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	data := PageData{
		Title:       "Messages - Rythm'it",
		CurrentPage: "messages",
		IsLoggedIn:  true,
		User:        user,
	}

	renderTemplate(w, "messages.html", data)
}

// ThreadHandler gère la page d'un thread individuel
func ThreadHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🧵 ThreadHandler appelé pour: %s", r.URL.Path)

	// Extraire l'ID du thread depuis l'URL
	path := strings.TrimPrefix(r.URL.Path, "/thread/")
	if path == "" {
		http.Error(w, "ID de thread manquant", http.StatusBadRequest)
		return
	}

	threadID, err := strconv.ParseUint(path, 10, 32)
	if err != nil {
		http.Error(w, "ID de thread invalide", http.StatusBadRequest)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)

	// Si c'est une requête POST, traiter l'ajout de commentaire
	if r.Method == "POST" {
		if !isLoggedIn {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		handleAddComment(w, r, uint(threadID), user)
		return
	}

	// Créer les services
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	likeRepo := repositories.NewLikeRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	// Récupérer le thread
	var userIDPtr *uint
	if user != nil {
		userIDPtr = &user.ID
	}

	threadDetails, err := threadService.GetThread(uint(threadID), userIDPtr)
	if err != nil {
		log.Printf("❌ Thread %d non trouvé: %v", threadID, err)
		http.Error(w, "Thread non trouvé", http.StatusNotFound)
		return
	}

	// Convertir le thread
	thread := convertDBThreadToPageThread(*threadDetails, user, likeRepo)

	// Récupérer les commentaires
	params := models.PaginationParams{
		Page:    1,
		PerPage: 50,
		Sort:    "created_at",
		Order:   "ASC",
	}

	var userID *uint
	if user != nil {
		userID = &user.ID
	}

	messages, _, err := messageRepo.GetMessagesWithVotes(uint(threadID), userID, params, "newest")
	if err != nil {
		log.Printf("❌ Erreur récupération commentaires: %v", err)
		// Continuer avec des commentaires vides plutôt que d'échouer
		messages = []*models.Message{}
	}

	// Convertir les messages en commentaires
	comments := convertMessagesToComments(messages, threadDetails.Author.Username, userIDPtr)

	// Récupérer les messages d'erreur/succès
	errorParam := r.URL.Query().Get("error")
	successParam := r.URL.Query().Get("success")

	var errorMessage, successMessage string
	switch errorParam {
	case "comment_failed":
		errorMessage = "Erreur lors de l'ajout du commentaire"
	case "empty_comment":
		errorMessage = "Le commentaire ne peut pas être vide"
	}

	switch successParam {
	case "comment_added":
		successMessage = "Commentaire ajouté avec succès !"
	}

	data := PageData{
		Title:          thread.Title + " - Rythm'it",
		CurrentPage:    "thread",
		IsLoggedIn:     isLoggedIn,
		User:           user,
		Thread:         &thread,
		Comments:       comments,
		ErrorMessage:   errorMessage,
		SuccessMessage: successMessage,
	}

	log.Printf("✅ Thread %d chargé: %s avec %d commentaires", threadID, thread.Title, len(comments))
	renderTemplate(w, "thread.html", data)
}

// ProfileHandler gère la page de profil (GET) et la mise à jour (POST)
func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("👤 ProfileHandler appelé - Method: %s", r.Method)

	// Récupérer l'utilisateur connecté depuis le cookie
	currentUser, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		log.Printf("❌ Utilisateur non connecté")
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	log.Printf("👤 Utilisateur connecté: %s (ID: %d)", currentUser.Username, currentUser.ID)

	// Vérifier si on affiche le profil d'un autre utilisateur
	targetUserIDStr := r.URL.Query().Get("user")
	var targetUser *models.User
	var isOwnProfile bool = true
	
	// Récupérer la connexion DB
	db := database.DB
	
	if targetUserIDStr != "" {
		// On affiche le profil d'un autre utilisateur
		targetUserID, err := strconv.ParseUint(targetUserIDStr, 10, 32)
		if err != nil {
			log.Printf("❌ ID utilisateur invalide: %v", err)
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
		
		if uint(targetUserID) == currentUser.ID {
			// Rediriger vers son propre profil
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
		
		// Récupérer l'utilisateur cible
		userRepoTemp := repositories.NewUserRepository(db)
		targetUser, err = userRepoTemp.FindByID(uint(targetUserID))
		if err != nil {
			log.Printf("❌ Utilisateur non trouvé: %v", err)
			http.Redirect(w, r, "/profile", http.StatusSeeOther)
			return
		}
		
		isOwnProfile = false
		log.Printf("👤 Affichage du profil de: %s (ID: %d)", targetUser.Username, targetUser.ID)
	} else {
		// On affiche son propre profil
		// Convertir currentUser (*User) vers *models.User
		userRepoTemp := repositories.NewUserRepository(db)
		var err error
		targetUser, err = userRepoTemp.FindByID(currentUser.ID)
		if err != nil {
			log.Printf("❌ Erreur récupération propre profil: %v", err)
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
	}

	log.Printf("👤 Profil cible: %s (ID: %d), Own: %t", targetUser.Username, targetUser.ID, isOwnProfile)

	// Debug: Vérifier si la table user_profiles existe
	var tableExists int
	err := db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'rythmit_db' AND table_name = 'user_profiles'").Scan(&tableExists)
	if err != nil {
		log.Printf("❌ Erreur vérification table user_profiles: %v", err)
	} else {
		log.Printf("🔍 Table user_profiles existe: %t", tableExists > 0)
	}

	// Debug: Vérifier les profils existants
	rows, err := db.Query("SELECT id, user_id, display_name, avatar_image, banner_image FROM user_profiles LIMIT 5")
	if err != nil {
		log.Printf("❌ Erreur lecture user_profiles: %v", err)
	} else {
		defer rows.Close()
		log.Printf("📋 Profils existants:")
		for rows.Next() {
			var id, userID uint
			var displayName, avatarImage, bannerImage sql.NullString
			err := rows.Scan(&id, &userID, &displayName, &avatarImage, &bannerImage)
			if err != nil {
				log.Printf("❌ Erreur scan profil: %v", err)
				continue
			}
			log.Printf("  ID: %d, UserID: %d, DisplayName: %v, Avatar: %v, Banner: %v", 
				id, userID, displayName.String, avatarImage.String, bannerImage.String)
		}
	}

	// Créer les services nécessaires
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	if r.Method == "POST" {
		// Seul le propriétaire du profil peut le modifier
		if !isOwnProfile {
			log.Printf("❌ Tentative de modification du profil d'un autre utilisateur")
			http.Redirect(w, r, fmt.Sprintf("/profile?user=%d", targetUser.ID), http.StatusSeeOther)
			return
		}
		
		user := currentUser
		log.Printf("📝 ===== DEBUT ProfileHandler POST =====")
		
		// Traiter la mise à jour du profil
		if err := r.ParseForm(); err != nil {
			log.Printf("❌ Erreur parsing formulaire: %v", err)
			http.Redirect(w, r, "/profile?error=form_error", http.StatusSeeOther)
			return
		}

		// Debug: afficher tous les paramètres du formulaire
		log.Printf("📝 Tous les paramètres du formulaire:")
		for key, values := range r.Form {
			log.Printf("  %s: %v", key, values)
		}

		// Récupérer les données du formulaire
		displayName := r.FormValue("display_name")
		avatarImage := r.FormValue("avatar_image")
		bannerImage := r.FormValue("banner_image")
		
		log.Printf("📝 Données extraites:")
		log.Printf("  - display_name: '%s'", displayName)
		log.Printf("  - avatar_image: '%s'", avatarImage)
		log.Printf("  - banner_image: '%s'", bannerImage)

		// Créer le DTO de mise à jour
		updateDTO := services.ProfileUpdateDTO{}

		// Toujours mettre à jour les champs (permet la suppression avec valeur vide)
		if displayName != "" {
			updateDTO.DisplayName = &displayName
		}

		// Pour les images, on accepte les valeurs vides pour la suppression
		if avatarImage == "" {
			updateDTO.AvatarImage = nil
		} else {
			updateDTO.AvatarImage = &avatarImage
		}

		if bannerImage == "" {
			updateDTO.BannerImage = nil
		} else {
			updateDTO.BannerImage = &bannerImage
		}

		log.Printf("🔄 DTO créé: DisplayName=%v, Avatar=%v, Banner=%v", 
			updateDTO.DisplayName, updateDTO.AvatarImage, updateDTO.BannerImage)

		// Tenter la mise à jour du profil existant
		log.Printf("🔄 Appel profileService.UpdateProfile...")
		_, err := profileService.UpdateProfile(user.ID, updateDTO)
		if err != nil {
			log.Printf("⚠️ Échec mise à jour, tentative de création: %v", err)
			// Si le profil n'existe pas, le créer
			createDTO := services.ProfileCreateDTO{
				DisplayName: updateDTO.DisplayName,
				AvatarImage: updateDTO.AvatarImage,
				BannerImage: updateDTO.BannerImage,
			}
			log.Printf("🔄 Appel profileService.CreateProfile...")
			_, err = profileService.CreateProfile(user.ID, createDTO)
			if err != nil {
				log.Printf("❌ Erreur création profil: %v", err)
				http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
				return
			}
			log.Printf("✅ Profil créé avec succès")
		} else {
			log.Printf("✅ Profil mis à jour avec succès")
		}

		log.Printf("✅ Profil mis à jour pour utilisateur %s", user.Username)
		log.Printf("📝 ===== FIN ProfileHandler POST =====")
		http.Redirect(w, r, "/profile?success=profile_updated", http.StatusSeeOther)
		return
	}

	// GET - Afficher la page de profil
	// Récupérer ou créer le profil utilisateur
	profileData, err := profileService.GetOrCreateProfile(targetUser.ID)
	if err != nil {
		log.Printf("❌ Erreur récupération profil: %v", err)
		// Continuer avec un profil vide plutôt que de faire échouer la page
		profileData = &services.ProfileResponseDTO{
			UserID:      targetUser.ID,
			DisplayName: nil,
			AvatarImage: nil,
			BannerImage: nil,
		}
	}

	log.Printf("📋 Profil récupéré: DisplayName=%v, Avatar=%v, Banner=%v", 
		profileData.DisplayName, profileData.AvatarImage, profileData.BannerImage)

	// Convertir le profil en format pour le template
	profile := &ProfileData{
		DisplayName: profileData.DisplayName,
		AvatarImage: profileData.AvatarImage,
		BannerImage: profileData.BannerImage,
	}

	// Récupérer le statut d'amitié si ce n'est pas son propre profil
	var friendshipStatus *string
	if !isOwnProfile {
		friendshipRepo := repositories.NewFriendshipRepository(db)
		status, err := friendshipRepo.GetFriendshipStatus(currentUser.ID, targetUser.ID)
		if err != nil {
			log.Printf("⚠️ Erreur récupération statut amitié: %v", err)
		}
		friendshipStatus = status
		log.Printf("👥 Statut d'amitié: %v", friendshipStatus)
	}

	// Récupérer les messages d'erreur/succès depuis les query parameters
	errorParam := r.URL.Query().Get("error")
	successParam := r.URL.Query().Get("success")

	var errorMessage, successMessage string

	// Traduction des messages d'erreur
	switch errorParam {
	case "form_error":
		errorMessage = "Erreur lors du traitement du formulaire"
	case "update_failed":
		errorMessage = "Erreur lors de la mise à jour du profil"
	case "invalid_action":
		errorMessage = "Action non reconnue"
	}

	// Traduction des messages de succès
	switch successParam {
	case "profile_updated":
		successMessage = "Profil mis à jour avec succès !"
	case "display_name_updated":
		successMessage = "Nom d'affichage mis à jour !"
	case "avatar_cleared":
		successMessage = "Avatar supprimé avec succès !"
	case "banner_cleared":
		successMessage = "Bannière supprimée avec succès !"
	}

	// Déterminer le titre de la page
	pageTitle := "Mon Profil - Rythm'it"
	if !isOwnProfile {
		pageTitle = fmt.Sprintf("Profil de %s - Rythm'it", targetUser.Username)
	}

	// Convertir targetUser en User pour le template
	displayUser := &User{
		ID:       targetUser.ID,
		Username: targetUser.Username,
		Email:    targetUser.Email,
		IsAdmin:  targetUser.IsAdmin,
		Avatar:   string(targetUser.Username[0:2]),
	}

	data := PageData{
		Title:            pageTitle,
		CurrentPage:      "profile",
		IsLoggedIn:       isLoggedIn,
		User:             displayUser,
		CurrentUser:      &User{ID: currentUser.ID, Username: currentUser.Username, Email: currentUser.Email, Avatar: string(currentUser.Username[0:2])},
		Profile:          profile,
		IsOwnProfile:     isOwnProfile,
		FriendshipStatus: friendshipStatus,
		ErrorMessage:     errorMessage,
		SuccessMessage:   successMessage,
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

// SigninHandler gère la page de connexion et la soumission du formulaire
func SigninHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔑 SigninHandler appelé - Method: %s", r.Method)

	if r.Method == "GET" {
		// Récupérer les messages d'erreur/succès depuis les query parameters
		errorParam := r.URL.Query().Get("error")
		successParam := r.URL.Query().Get("success")

		var errorMessage, successMessage string

		// Traduction des messages d'erreur
		switch errorParam {
		case "invalid_credentials":
			errorMessage = "Email ou mot de passe incorrect"
		case "missing_fields":
			errorMessage = "Veuillez remplir tous les champs"
		case "password_mismatch":
			errorMessage = "Les mots de passe ne correspondent pas"
		case "password_too_short":
			errorMessage = "Le mot de passe doit contenir au moins 6 caractères"
		case "registration_failed":
			errorMessage = "Erreur lors de l'inscription. L'email est peut-être déjà utilisé."
		}

		// Traduction des messages de succès
		switch successParam {
		case "registration_complete":
			successMessage = "Inscription réussie ! Vous pouvez maintenant vous connecter."
		}

		// Afficher la page de connexion
		data := PageData{
			Title:          "Connexion - Rythm'it",
			CurrentPage:    "signin",
			IsLoggedIn:     false,
			IsSignupMode:   false,
			ErrorMessage:   errorMessage,
			SuccessMessage: successMessage,
			AuthTitle:      "Bon retour !",
			AuthSubtitle:   "Connectez-vous pour retrouver votre univers musical",
			AuthButtonText: "Se connecter",
			AuthFooterText: template.HTML(`Vous n'avez pas de compte ? <a href="/signup">S'inscrire</a>`),
		}
		renderTemplate(w, "signin", data)
		return
	}

	if r.Method == "POST" {
		log.Printf("🔐 POST reçu sur /signin")

		// Traiter la soumission du formulaire de connexion
		if err := r.ParseForm(); err != nil {
			log.Printf("❌ Erreur lors du parsing du formulaire: %v", err)
			http.Error(w, "Erreur lors du traitement du formulaire", http.StatusBadRequest)
			return
		}

		email := r.FormValue("email")
		password := r.FormValue("password")
		rememberMe := r.FormValue("rememberMe") == "on"

		log.Printf("📧 Tentative de connexion pour: %s (Remember: %v)", email, rememberMe)
		log.Printf("🔍 Mot de passe reçu: [%d caractères]", len(password))

		// Validation des champs
		if email == "" || password == "" {
			log.Printf("❌ Champs manquants - Email: %s, Password: %s", email, password)
			http.Redirect(w, r, "/signin?error=missing_fields", http.StatusSeeOther)
			return
		}

		// Créer le DTO de connexion
		loginDTO := services.LoginDTO{
			Identifier: email,
			Password:   password,
		}

		log.Printf("📊 LoginDTO créé: %+v", loginDTO)

		// Créer le service d'authentification
		db := database.DB
		if db == nil {
			log.Printf("❌ Base de données non connectée")
			http.Redirect(w, r, "/signin?error=invalid_credentials", http.StatusSeeOther)
			return
		}

		userRepo := repositories.NewUserRepository(db)
		cfg := configs.Get()
		authService := services.NewAuthService(userRepo, cfg)

		log.Printf("🔧 Service d'authentification créé")

		token, user, err := authService.Login(loginDTO)
		if err != nil {
			log.Printf("❌ Échec de la connexion: %v", err)
			// Rediriger vers la page de connexion avec une erreur
			http.Redirect(w, r, "/signin?error=invalid_credentials", http.StatusSeeOther)
			return
		}

		log.Printf("🎉 Connexion réussie!")
		log.Printf("👤 Utilisateur: %s (ID: %d)", user.Username, user.ID)
		log.Printf("🔑 Token généré: %d caractères", len(token))

		// Définir le cookie d'authentification
		maxAge := 24 * 60 * 60 // 24 heures par défaut
		if rememberMe {
			maxAge = 7 * 24 * 60 * 60 // 7 jours si "se souvenir"
		}

		cookie := &http.Cookie{
			Name:     "auth_token",
			Value:    token,
			Path:     "/",
			MaxAge:   maxAge,
			HttpOnly: true,
			Secure:   false, // En développement
			SameSite: http.SameSiteLaxMode,
		}

		http.SetCookie(w, cookie)
		log.Printf("🍪 Cookie défini: %s", cookie.Name)

		log.Printf("✅ Connexion réussie pour: %s (ID: %d)", user.Username, user.ID)
		// Rediriger vers l'accueil
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
}

// SignupHandler gère la page d'inscription et la soumission du formulaire
func SignupHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("📝 SignupHandler appelé - Method: %s", r.Method)

	if r.Method == "GET" {
		// Récupérer les messages d'erreur depuis les query parameters
		errorParam := r.URL.Query().Get("error")

		var errorMessage string

		// Traduction des messages d'erreur
		switch errorParam {
		case "missing_fields":
			errorMessage = "Veuillez remplir tous les champs"
		case "password_mismatch":
			errorMessage = "Les mots de passe ne correspondent pas"
		case "password_too_short":
			errorMessage = "Le mot de passe doit contenir au moins 6 caractères"
		case "registration_failed":
			errorMessage = "Erreur lors de l'inscription. L'email est peut-être déjà utilisé."
		}

		// Afficher la page d'inscription
		data := PageData{
			Title:          "Inscription - Rythm'it",
			CurrentPage:    "signup",
			IsLoggedIn:     false,
			IsSignupMode:   true,
			ErrorMessage:   errorMessage,
			SuccessMessage: "",
			AuthTitle:      "Rejoignez Rythm'it !",
			AuthSubtitle:   "Créez votre compte pour découvrir votre univers musical",
			AuthButtonText: "S'inscrire",
			AuthFooterText: template.HTML(`Vous avez déjà un compte ? <a href="/signin">Se connecter</a>`),
		}
		renderTemplate(w, "signin", data)
		return
	}

	if r.Method == "POST" {
		// Traiter la soumission du formulaire d'inscription
		if err := r.ParseForm(); err != nil {
			log.Printf("❌ Erreur lors du parsing du formulaire: %v", err)
			http.Error(w, "Erreur lors du traitement du formulaire", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		email := r.FormValue("email")
		password := r.FormValue("password")
		confirmPassword := r.FormValue("confirmPassword")

		log.Printf("📝 Tentative d'inscription pour: %s (%s)", username, email)

		// Validation côté serveur
		if username == "" || email == "" || password == "" || confirmPassword == "" {
			log.Printf("❌ Champs manquants")
			http.Redirect(w, r, "/signup?error=missing_fields", http.StatusSeeOther)
			return
		}

		if password != confirmPassword {
			log.Printf("❌ Les mots de passe ne correspondent pas")
			http.Redirect(w, r, "/signup?error=password_mismatch", http.StatusSeeOther)
			return
		}

		if len(password) < 6 {
			log.Printf("❌ Mot de passe trop court")
			http.Redirect(w, r, "/signup?error=password_too_short", http.StatusSeeOther)
			return
		}

		// Créer le DTO d'inscription
		registerDTO := services.RegisterDTO{
			Username: username,
			Email:    email,
			Password: password,
		}

		// Créer le service d'authentification
		db := database.DB
		userRepo := repositories.NewUserRepository(db)
		cfg := configs.Get()
		authService := services.NewAuthService(userRepo, cfg)

		user, err := authService.Register(registerDTO)
		if err != nil {
			log.Printf("❌ Échec de l'inscription: %v", err)
			// Rediriger vers la page d'inscription avec une erreur
			http.Redirect(w, r, "/signup?error=registration_failed", http.StatusSeeOther)
			return
		}

		log.Printf("✅ Inscription réussie pour: %s (ID: %d)", user.Username, user.ID)
		// Rediriger vers la page de connexion avec un message de succès
		http.Redirect(w, r, "/signin?success=registration_complete", http.StatusSeeOther)
		return
	}

	http.Error(w, "Méthode non supportée", http.StatusMethodNotAllowed)
}

// LogoutHandler gère la déconnexion
func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🚪 LogoutHandler appelé")

	// Supprimer le cookie d'authentification
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1, // Supprime le cookie
		HttpOnly: true,
		Secure:   false, // En développement
		SameSite: http.SameSiteLaxMode,
	})

	// Rediriger vers la page de connexion
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
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

	// Récupérer les threads avec pagination (5 premiers threads)
	params := models.PaginationParams{
		Page:    1,
		PerPage: 5, // Afficher seulement 5 threads initialement
		Sort:    "created_at",
		Order:   "DESC",
	}

	// Utiliser la méthode avec pagination
	response, err := threadService.GetPublicThreads(params, services.ThreadFilters{})
	if err != nil {
		return nil, err
	}

	// Convertir les DTOs de réponse en ThreadDTO pour compatibilité
	var threads []services.ThreadDTO
	for _, threadResp := range response.Threads {
		// Parser la date de création depuis le format ISO
		createdAt, err := time.Parse("2006-01-02T15:04:05Z", threadResp.CreatedAt)
		if err != nil {
			log.Printf("❌ Erreur parsing date pour thread %d: %v", threadResp.ID, err)
			createdAt = time.Now() // Fallback
		}

		// Parser la date de mise à jour depuis le format ISO
		updatedAt, err := time.Parse("2006-01-02T15:04:05Z", threadResp.UpdatedAt)
		if err != nil {
			log.Printf("❌ Erreur parsing date update pour thread %d: %v", threadResp.ID, err)
			updatedAt = time.Now() // Fallback
		}

		thread := services.ThreadDTO{
			ID:           threadResp.ID,
			Title:        threadResp.Title,
			Content:      threadResp.Description,
			Username:     threadResp.Author.Username,
			UserID:       threadResp.Author.ID,
			MessageCount: threadResp.MessageCount, // Utiliser le vrai nombre de messages
			CreatedAt:    createdAt,
			UpdatedAt:    updatedAt,
			Tags:         make([]string, len(threadResp.Tags)),
		}

		// Convertir les tags
		for i, tag := range threadResp.Tags {
			thread.Tags[i] = tag.Name
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

// getAvailableTagsFromDatabase récupère les tags disponibles depuis la base de données
func getAvailableTagsFromDatabase() ([]Tag, error) {
	// Créer les dépendances
	db := database.DB
	tagRepo := repositories.NewTagRepository(db)

	// Récupérer tous les tags
	dbTags, err := tagRepo.FindAll()
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags: %w", err)
	}

	// Convertir au format pour le template
	var tags []Tag
	for _, dbTag := range dbTags {
		tag := Tag{
			ID:   dbTag.ID,
			Name: dbTag.Name,
			Type: dbTag.Type,
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// convertDBThreadsToPageThreads convertit les threads de la DB au format attendu par le template
func convertDBThreadsToPageThreads(dbThreads []services.ThreadDTO, currentUser *User) []Thread {
	var pageThreads []Thread

	for _, dbThread := range dbThreads {
		// Debug: afficher les données du thread
		log.Printf("🔍 Thread ID=%d, Title='%s', Content='%s', User='%s'",
			dbThread.ID, dbThread.Title, dbThread.Content, dbThread.Username)

		// Générer les initiales pour l'avatar
		initials := generateInitials(dbThread.Username)

		// Formater le temps
		timeAgo := formatTimeAgo(dbThread.CreatedAt)

		// Déterminer l'auteur à afficher (YOU si c'est l'utilisateur connecté)
		authorDisplay := dbThread.Username
		if currentUser != nil && currentUser.Username == dbThread.Username {
			authorDisplay = "YOU"
		}

		// Récupérer les données de likes
		db := database.DB
		likeRepo := repositories.NewLikeRepository(db)

		// Récupérer le nombre réel de likes depuis la base de données
		likesCount, err := likeRepo.GetThreadLikesCount(dbThread.ID)
		if err != nil {
			log.Printf("❌ Erreur récupération compteur likes pour thread %d: %v", dbThread.ID, err)
			likesCount = 0 // Fallback en cas d'erreur
		}

		// Vérifier si l'utilisateur connecté a liké ce thread
		isLiked := false
		if currentUser != nil {
			liked, err := likeRepo.IsThreadLikedByUser(currentUser.ID, dbThread.ID)
			if err == nil {
				isLiked = liked
			}
		}

		// Créer le thread pour le template
		pageThread := Thread{
			ID:           dbThread.ID,
			Title:        dbThread.Title,
			Content:      dbThread.Content,
			Author:       authorDisplay,
			AuthorAvatar: initials,
			TimeAgo:      timeAgo,
			Genre:        "Discussion", // Changé de genre spécifique à "Discussion"
			Tags:         dbThread.Tags,
			Likes:        likesCount,
			IsLiked:      isLiked,
			Comments:     dbThread.MessageCount,
			Shares:       0,        // TODO: implémenter les partages
			Visibility:   "public", // Valeur par défaut
			State:        "ouvert", // Valeur par défaut
			MusicTrack:   nil,      // Pas de piste musicale pour l'instant
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

// getUserFromCookie vérifie l'authentification depuis le cookie et retourne l'utilisateur
func getUserFromCookie(r *http.Request) (*User, bool) {
	// Récupérer le cookie d'authentification
	cookie, err := r.Cookie("auth_token")
	if err != nil {
		log.Printf("🔒 Aucun cookie d'authentification trouvé")
		return nil, false
	}

	// Créer le service d'authentification pour valider le token
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	cfg := configs.Get()
	authService := services.NewAuthService(userRepo, cfg)

	// Valider le token
	claims, err := authService.ParseToken(cookie.Value)
	if err != nil {
		log.Printf("🔒 Token invalide ou expiré: %v", err)
		return nil, false
	}

	// Créer l'utilisateur à partir des claims
	user := &User{
		ID:       claims.UserID,
		Username: claims.Username,
		Email:    claims.Email,
		IsAdmin:  claims.IsAdmin,
		Avatar:   generateInitials(claims.Username),
	}

	log.Printf("👤 Utilisateur connecté: %s (ID: %d)", claims.Username, claims.UserID)
	return user, true
}

// TagsAPIHandler retourne la liste des tags disponibles en JSON
func TagsAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🏷️ TagsAPIHandler appelé")

	// Récupérer les tags depuis la DB
	tags, err := getAvailableTagsFromDatabase()
	if err != nil {
		log.Printf("❌ Erreur récupération tags: %v", err)
		http.Error(w, "Erreur récupération tags", http.StatusInternalServerError)
		return
	}

	// Retourner en JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := map[string]interface{}{
		"success": true,
		"data":    tags,
	}

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("❌ Erreur encodage JSON: %v", err)
		http.Error(w, "Erreur encodage JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ %d tags retournés", len(tags))
}

// ThreadsAPIHandler retourne les threads avec pagination en JSON
func ThreadsAPIHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🧵 ThreadsAPIHandler appelé")

	// Récupérer les paramètres de pagination
	pageStr := r.URL.Query().Get("page")
	page := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	perPageStr := r.URL.Query().Get("per_page")
	perPage := 10
	if perPageStr != "" {
		if pp, err := strconv.Atoi(perPageStr); err == nil && pp > 0 && pp <= 50 {
			perPage = pp
		}
	}

	// Créer les dépendances
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	// Paramètres de pagination
	params := models.PaginationParams{
		Page:    page,
		PerPage: perPage,
		Sort:    "created_at",
		Order:   "DESC",
	}

	// Récupérer les threads
	response, err := threadService.GetPublicThreads(params, services.ThreadFilters{})
	if err != nil {
		log.Printf("❌ Erreur récupération threads: %v", err)
		http.Error(w, "Erreur récupération threads", http.StatusInternalServerError)
		return
	}

	// Récupérer l'utilisateur connecté pour les données de likes
	user, _ := getUserFromCookie(r)

	// Convertir en format Thread pour le frontend
	var threads []Thread
	for _, threadResp := range response.Threads {
		// Générer les initiales
		initials := generateInitials(threadResp.Author.Username)

		// Parser la date (format ISO)
		createdAt, err := time.Parse("2006-01-02T15:04:05Z", threadResp.CreatedAt)
		if err != nil {
			createdAt = time.Now() // Fallback
		}
		timeAgo := formatTimeAgo(createdAt)

		// Déterminer l'auteur à afficher
		authorDisplay := threadResp.Author.Username
		if user != nil && user.Username == threadResp.Author.Username {
			authorDisplay = "YOU"
		}

		// Récupérer les données de likes
		likeRepo := repositories.NewLikeRepository(db)
		likesCount, err := likeRepo.GetThreadLikesCount(threadResp.ID)
		if err != nil {
			likesCount = 0
		}

		isLiked := false
		if user != nil {
			liked, err := likeRepo.IsThreadLikedByUser(user.ID, threadResp.ID)
			if err == nil {
				isLiked = liked
			}
		}

		// Convertir les tags en slice de strings
		tags := make([]string, len(threadResp.Tags))
		for i, tag := range threadResp.Tags {
			tags[i] = tag.Name
		}

		// Créer le thread pour le frontend
		thread := Thread{
			ID:           threadResp.ID,
			Title:        threadResp.Title,
			Content:      threadResp.Description,
			ImageURL:     threadResp.ImageURL,
			Author:       authorDisplay,
			AuthorAvatar: initials,
			TimeAgo:      timeAgo,
			Genre:        "Discussion",
			Tags:         tags,
			Likes:        likesCount,
			IsLiked:      isLiked,
			Comments:     threadResp.MessageCount,
			Shares:       0,
			MusicTrack:   nil,
		}

		threads = append(threads, thread)
	}

	// Retourner en JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	apiResponse := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"threads":    threads,
			"pagination": response.Pagination,
		},
	}

	if err := json.NewEncoder(w).Encode(apiResponse); err != nil {
		log.Printf("❌ Erreur encodage JSON: %v", err)
		http.Error(w, "Erreur encodage JSON", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ %d threads retournés (page %d)", len(threads), page)
}

// SimpleProfileUpdateHandler gère les mises à jour de profil simples sans JavaScript
func SimpleProfileUpdateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔄 SimpleProfileUpdateHandler appelé - Method: %s", r.Method)
	
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Debug: afficher le Content-Type
	log.Printf("📝 Content-Type: %s", r.Header.Get("Content-Type"))

	// Parser le formulaire
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Erreur parsing formulaire: %v", err)
		http.Redirect(w, r, "/profile?error=form_error", http.StatusSeeOther)
		return
	}

	// Récupérer l'action
	action := r.FormValue("action")
	log.Printf("🔄 Action reçue: '%s'", action)
	
	// Debug: afficher tous les paramètres du formulaire
	log.Printf("📝 Tous les paramètres du formulaire:")
	for key, values := range r.Form {
		log.Printf("  %s: %v", key, values)
	}

	// Debug: afficher aussi les paramètres POST
	log.Printf("📝 Paramètres POST:")
	for key, values := range r.PostForm {
		log.Printf("  %s: %v", key, values)
	}

	switch action {
	case "update_display_name":
		handleDisplayNameUpdate(w, r, user)
	case "update_avatar":
		handleUpdateAvatar(w, r, user)
	case "update_banner":
		handleUpdateBanner(w, r, user)
	case "clear_avatar":
		handleClearAvatar(w, r, user)
	case "clear_banner":
		handleClearBanner(w, r, user)
	default:
		log.Printf("❌ Action non reconnue: '%s'", action)
		http.Redirect(w, r, "/profile?error=invalid_action", http.StatusSeeOther)
	}
}

func handleDisplayNameUpdate(w http.ResponseWriter, r *http.Request, user *User) {
	displayName := r.FormValue("display_name")

	// Créer les services
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	// Préparer le DTO
	updateDTO := services.ProfileUpdateDTO{}
	if displayName != "" {
		updateDTO.DisplayName = &displayName
	}

	// Mettre à jour le profil
	_, err := profileService.UpdateProfile(user.ID, updateDTO)
	if err != nil {
		// Si le profil n'existe pas, le créer
		createDTO := services.ProfileCreateDTO{
			DisplayName: updateDTO.DisplayName,
		}
		_, err = profileService.CreateProfile(user.ID, createDTO)
		if err != nil {
			log.Printf("❌ Erreur création profil: %v", err)
			http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
			return
		}
	}

	log.Printf("✅ Nom d'affichage mis à jour pour %s: %s", user.Username, displayName)
	http.Redirect(w, r, "/profile?success=display_name_updated", http.StatusSeeOther)
}

func handleUpdateAvatar(w http.ResponseWriter, r *http.Request, user *User) {
	avatarImage := r.FormValue("avatar_image")
	log.Printf("🔄 ===== DEBUT handleUpdateAvatar =====")
	log.Printf("👤 Utilisateur: %s (ID: %d)", user.Username, user.ID)
	log.Printf("🖼️ Avatar image URL reçue: '%s'", avatarImage)
	
	// Créer les services
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	// Préparer le DTO pour mettre à jour l'avatar
	updateDTO := services.ProfileUpdateDTO{}
	if avatarImage != "" {
		updateDTO.AvatarImage = &avatarImage
		log.Printf("📝 DTO créé avec Avatar image: %s", avatarImage)
	} else {
		log.Printf("⚠️ Avatar image vide, DTO sans avatar")
	}

	log.Printf("🔄 Appel profileService.UpdateProfile...")
	// Mettre à jour le profil
	_, err := profileService.UpdateProfile(user.ID, updateDTO)
	if err != nil {
		log.Printf("❌ ERREUR mise à jour avatar: %v", err)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Avatar mis à jour avec SUCCES pour %s", user.Username)
	log.Printf("🔄 ===== FIN handleUpdateAvatar =====")
	http.Redirect(w, r, "/profile?success=avatar_updated", http.StatusSeeOther)
}

func handleUpdateBanner(w http.ResponseWriter, r *http.Request, user *User) {
	bannerImage := r.FormValue("banner_image")
	log.Printf("🔄 ===== DEBUT handleUpdateBanner =====")
	log.Printf("👤 Utilisateur: %s (ID: %d)", user.Username, user.ID)
	log.Printf("🖼️ Banner image URL reçue: '%s'", bannerImage)
	
	// Créer les services
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	// Préparer le DTO pour mettre à jour la bannière
	updateDTO := services.ProfileUpdateDTO{}
	if bannerImage != "" {
		updateDTO.BannerImage = &bannerImage
		log.Printf("📝 DTO créé avec Banner image: %s", bannerImage)
	} else {
		log.Printf("⚠️ Banner image vide, DTO sans banner")
	}

	log.Printf("🔄 Appel profileService.UpdateProfile...")
	// Mettre à jour le profil
	_, err := profileService.UpdateProfile(user.ID, updateDTO)
	if err != nil {
		log.Printf("❌ ERREUR mise à jour bannière: %v", err)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Bannière mise à jour avec SUCCES pour %s", user.Username)
	log.Printf("🔄 ===== FIN handleUpdateBanner =====")
	http.Redirect(w, r, "/profile?success=banner_updated", http.StatusSeeOther)
}

func handleClearAvatar(w http.ResponseWriter, r *http.Request, user *User) {
	// Créer les services
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	// Préparer le DTO pour effacer l'avatar
	updateDTO := services.ProfileUpdateDTO{
		AvatarImage: nil,
	}

	// Mettre à jour le profil
	_, err := profileService.UpdateProfile(user.ID, updateDTO)
	if err != nil {
		log.Printf("❌ Erreur suppression avatar: %v", err)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Avatar supprimé pour %s", user.Username)
	http.Redirect(w, r, "/profile?success=avatar_cleared", http.StatusSeeOther)
}

func handleClearBanner(w http.ResponseWriter, r *http.Request, user *User) {
	// Créer les services
	db := database.DB
	userRepo := repositories.NewUserRepository(db)
	profileRepo := repositories.NewProfileRepository(db)
	profileService := services.NewProfileService(profileRepo, userRepo)

	// Préparer le DTO pour effacer la bannière
	updateDTO := services.ProfileUpdateDTO{
		BannerImage: nil,
	}

	// Mettre à jour le profil
	_, err := profileService.UpdateProfile(user.ID, updateDTO)
	if err != nil {
		log.Printf("❌ Erreur suppression bannière: %v", err)
		http.Redirect(w, r, "/profile?error=update_failed", http.StatusSeeOther)
		return
	}

	log.Printf("✅ Bannière supprimée pour %s", user.Username)
	http.Redirect(w, r, "/profile?success=banner_cleared", http.StatusSeeOther)
}

// handleAddComment gère l'ajout d'un commentaire à un thread
func handleAddComment(w http.ResponseWriter, r *http.Request, threadID uint, user *User) {
	if err := r.ParseForm(); err != nil {
		log.Printf("❌ Erreur parsing formulaire: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/thread/%d?error=comment_failed", threadID), http.StatusSeeOther)
		return
	}

	content := strings.TrimSpace(r.FormValue("comment"))
	commentImageURL := r.FormValue("comment_image_url") // Nouveau champ pour l'image

	if content == "" {
		http.Redirect(w, r, fmt.Sprintf("/thread/%d?error=empty_comment", threadID), http.StatusSeeOther)
		return
	}

	// Créer les services
	db := database.DB
	messageRepo := repositories.NewMessageRepository(db)

	// Créer le message/commentaire
	message := &models.Message{
		Content:  content,
		ThreadID: threadID,
		UserID:   user.ID,
	}

	// Ajouter l'image si elle est fournie
	if strings.TrimSpace(commentImageURL) != "" {
		message.ImageURL = &commentImageURL
	}

	err := messageRepo.Create(message)
	if err != nil {
		log.Printf("❌ Erreur création commentaire: %v", err)
		http.Redirect(w, r, fmt.Sprintf("/thread/%d?error=comment_failed", threadID), http.StatusSeeOther)
		return
	}

	log.Printf("✅ Commentaire ajouté par %s sur thread %d", user.Username, threadID)
	http.Redirect(w, r, fmt.Sprintf("/thread/%d?success=comment_added", threadID), http.StatusSeeOther)
}

// convertDBThreadToPageThread convertit un thread de la DB au format de la page
func convertDBThreadToPageThread(threadResp services.ThreadResponseDTO, user *User, likeRepo repositories.LikeRepository) Thread {
	// Générer les initiales
	initials := generateInitials(threadResp.Author.Username)

	// Parser la date
	createdAt, err := time.Parse("2006-01-02T15:04:05Z", threadResp.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}
	timeAgo := formatTimeAgo(createdAt)

	// Déterminer l'auteur à afficher
	authorDisplay := threadResp.Author.Username
	if user != nil && user.Username == threadResp.Author.Username {
		authorDisplay = "YOU"
	}

	// Récupérer les données de likes
	likesCount, err := likeRepo.GetThreadLikesCount(threadResp.ID)
	if err != nil {
		likesCount = 0
	}

	isLiked := false
	if user != nil {
		liked, err := likeRepo.IsThreadLikedByUser(user.ID, threadResp.ID)
		if err == nil {
			isLiked = liked
		}
	}

	// Convertir les tags
	tags := make([]string, len(threadResp.Tags))
	for i, tag := range threadResp.Tags {
		tags[i] = tag.Name
	}

	return Thread{
		ID:           threadResp.ID,
		Title:        threadResp.Title,
		Content:      threadResp.Description,
			ImageURL:     threadResp.ImageURL,
		Author:       authorDisplay,
		AuthorAvatar: initials,
		TimeAgo:      timeAgo,
			Genre:        "Discussion",
		Tags:         tags,
		Likes:        likesCount,
		IsLiked:      isLiked,
		Comments:     threadResp.MessageCount,
		Shares:       0,
		MusicTrack:   nil,
	}
}

// convertMessagesToComments convertit les messages de la DB en commentaires
func convertMessagesToComments(messages []*models.Message, threadAuthor string, userID *uint) []Comment {
	var comments []Comment
	db := database.DB

	for _, msg := range messages {
		// Générer les initiales
		initials := generateInitials(msg.Author.Username)

		// Formater le temps
		timeAgo := formatTimeAgo(msg.CreatedAt)

		// Vérifier si c'est l'auteur original du thread
		isOP := msg.Author.Username == threadAuthor

		// Récupérer le nombre de likes depuis la table comment_likes
		var likesCount int
		countQuery := "SELECT COUNT(*) FROM comment_likes WHERE message_id = ?"
		err := db.QueryRow(countQuery, msg.ID).Scan(&likesCount)
		if err != nil {
			log.Printf("❌ Erreur comptage likes pour message %d: %v", msg.ID, err)
			likesCount = 0
		}

		// Vérifier si l'utilisateur connecté a liké ce message
		var isLiked bool
		if userID != nil && *userID > 0 {
			var userLikeCount int
			userLikeQuery := "SELECT COUNT(*) FROM comment_likes WHERE user_id = ? AND message_id = ?"
			err = db.QueryRow(userLikeQuery, *userID, msg.ID).Scan(&userLikeCount)
			if err != nil {
				log.Printf("❌ Erreur vérification like utilisateur pour message %d: %v", msg.ID, err)
				isLiked = false
			} else {
				isLiked = userLikeCount > 0
			}
		}

		comment := Comment{
			ID:           msg.ID,
			Content:      msg.Content,
			ImageURL:     msg.ImageURL,
			Author:       msg.Author.Username,
			AuthorAvatar: initials,
			TimeAgo:      timeAgo,
			Likes:        likesCount,
			IsLiked:      isLiked,
			IsOP:         isOP,
			Replies:      []Comment{}, // TODO: Implémenter les réponses
		}

		comments = append(comments, comment)
	}

	return comments
}

// DeleteThreadHandler gère la suppression d'un thread
func DeleteThreadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	// Récupérer l'ID du thread depuis l'URL
	vars := mux.Vars(r)
	threadIDStr := vars["id"]
	threadID, err := strconv.ParseUint(threadIDStr, 10, 32)
	if err != nil {
		log.Printf("❌ ID thread invalide: %s", threadIDStr)
		http.Error(w, "ID thread invalide", http.StatusBadRequest)
		return
	}

	log.Printf("🗑️ Demande suppression thread %d par %s (ID: %d)", threadID, user.Username, user.ID)

	// Créer les services
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	// Récupérer le thread pour vérifier la propriété
	threadData, err := threadService.GetThread(uint(threadID), &user.ID)
	if err != nil {
		log.Printf("❌ Thread %d non trouvé: %v", threadID, err)
		http.Error(w, "Thread non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier que l'utilisateur est le propriétaire du thread OU un admin
	if threadData.Author.ID != user.ID && !user.IsAdmin {
		log.Printf("❌ Utilisateur %s (ID: %d) tente de supprimer thread %d qui appartient à %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	if user.IsAdmin && threadData.Author.ID != user.ID {
		log.Printf("🔧 Admin %s (ID: %d) supprime le thread %d de %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
	}

	// Supprimer le thread (l'utilisateur n'est pas admin, mais il est propriétaire)
	err = threadService.DeleteThread(uint(threadID), user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("❌ Erreur suppression thread %d: %v", threadID, err)
		http.Error(w, "Erreur lors de la suppression du thread", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Thread %d supprimé avec succès par %s", threadID, user.Username)

	// Retourner une réponse JSON pour AJAX
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Thread supprimé avec succès",
	})
}

// EditThreadHandler gère l'édition d'un thread
func EditThreadHandler(w http.ResponseWriter, r *http.Request) {
	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	// Récupérer l'ID du thread depuis l'URL
	vars := mux.Vars(r)
	threadIDStr := vars["id"]
	threadID, err := strconv.ParseUint(threadIDStr, 10, 32)
	if err != nil {
		log.Printf("❌ ID thread invalide: %s", threadIDStr)
		http.Error(w, "ID thread invalide", http.StatusBadRequest)
		return
	}

	// Créer les services
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	if r.Method == "GET" {
		// Afficher le formulaire d'édition
		handleShowEditForm(w, r, uint(threadID), user, threadService)
	} else if r.Method == "POST" {
		// Traiter la modification
		handleUpdateThread(w, r, uint(threadID), user, threadService)
	} else {
		http.Error(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// handleShowEditForm affiche le formulaire d'édition d'un thread
func handleShowEditForm(w http.ResponseWriter, r *http.Request, threadID uint, user *User, threadService services.ThreadService) {
	log.Printf("✏️ Affichage formulaire édition thread %d par %s (ID: %d)", threadID, user.Username, user.ID)

	// Récupérer le thread pour vérifier la propriété
	threadData, err := threadService.GetThread(threadID, &user.ID)
	if err != nil {
		log.Printf("❌ Thread %d non trouvé: %v", threadID, err)
		http.Error(w, "Thread non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier que l'utilisateur est le propriétaire du thread OU un admin
	if threadData.Author.ID != user.ID && !user.IsAdmin {
		log.Printf("❌ Utilisateur %s (ID: %d) tente de modifier thread %d qui appartient à %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	if user.IsAdmin && threadData.Author.ID != user.ID {
		log.Printf("🔧 Admin %s (ID: %d) accède à l'édition du thread %d de %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
	}

	// Récupérer les messages d'erreur/succès
	errorParam := r.URL.Query().Get("error")
	successParam := r.URL.Query().Get("success")

	var errorMessage, successMessage string
	switch errorParam {
	case "update_failed":
		errorMessage = "Erreur lors de la mise à jour du thread"
	case "validation_failed":
		errorMessage = "Données invalides. Vérifiez que le titre fait au moins 5 caractères et la description au moins 10 caractères."
	}

	switch successParam {
	case "updated":
		successMessage = "Thread mis à jour avec succès !"
	}

	// Préparer les données pour le template
	data := PageData{
		Title:          "Modifier le thread - " + threadData.Title,
		CurrentPage:    "edit-thread",
		IsLoggedIn:     true,
		User:           user,
		Thread:         convertThreadResponseToPageThread(*threadData, user),
		ErrorMessage:   errorMessage,
		SuccessMessage: successMessage,
	}

	log.Printf("✅ Formulaire d'édition préparé pour le thread %d", threadID)
	renderTemplate(w, "edit-thread.html", data)
}

// handleUpdateThread traite la mise à jour d'un thread
func handleUpdateThread(w http.ResponseWriter, r *http.Request, threadID uint, user *User, threadService services.ThreadService) {
	log.Printf("📝 Mise à jour thread %d par %s (ID: %d)", threadID, user.Username, user.ID)

	// Récupérer le thread pour vérifier la propriété
	threadData, err := threadService.GetThread(threadID, &user.ID)
	if err != nil {
		log.Printf("❌ Thread %d non trouvé: %v", threadID, err)
		http.Error(w, "Thread non trouvé", http.StatusNotFound)
		return
	}

	// Vérifier que l'utilisateur est le propriétaire du thread OU un admin
	if threadData.Author.ID != user.ID && !user.IsAdmin {
		log.Printf("❌ Utilisateur %s (ID: %d) tente de modifier thread %d qui appartient à %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
		http.Error(w, "Non autorisé", http.StatusForbidden)
		return
	}

	if user.IsAdmin && threadData.Author.ID != user.ID {
		log.Printf("🔧 Admin %s (ID: %d) modifie le thread %d de %s (ID: %d)",
			user.Username, user.ID, threadID, threadData.Author.Username, threadData.Author.ID)
	}

	// Parser le formulaire (essayer multipart d'abord, puis form normal)
	err = r.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		// Si ce n'est pas un formulaire multipart, essayer le parsing normal
		if err = r.ParseForm(); err != nil {
			log.Printf("❌ Erreur parsing formulaire: %v", err)
			http.Redirect(w, r, fmt.Sprintf("/thread/%d/edit?error=validation_failed", threadID), http.StatusSeeOther)
			return
		}
	}

	// Récupérer les données du formulaire
	title := strings.TrimSpace(r.FormValue("title"))
	description := strings.TrimSpace(r.FormValue("description"))
	imageURL := strings.TrimSpace(r.FormValue("image_url"))
	visibility := strings.TrimSpace(r.FormValue("visibility"))
	state := strings.TrimSpace(r.FormValue("state"))
	tagsStr := strings.TrimSpace(r.FormValue("tags"))

	log.Printf("📝 Données reçues: title='%s', description='%s', imageURL='%s', visibility='%s', state='%s', tags='%s'",
		title, description, imageURL, visibility, state, tagsStr)

	// Validation de base
	if len(title) < 5 || len(title) > 200 {
		log.Printf("❌ Titre invalide: '%s' (longueur: %d)", title, len(title))
		http.Redirect(w, r, fmt.Sprintf("/thread/%d/edit?error=validation_failed", threadID), http.StatusSeeOther)
		return
	}

	if len(description) < 10 {
		log.Printf("❌ Description invalide: '%s' (longueur: %d)", description, len(description))
		http.Redirect(w, r, fmt.Sprintf("/thread/%d/edit?error=validation_failed", threadID), http.StatusSeeOther)
		return
	}

	// Valeurs par défaut
	if visibility == "" {
		visibility = "public"
	}
	if state == "" {
		state = "ouvert"
	}

	// Traiter l'image
	var imageURLPtr *string
	if imageURL != "" {
		imageURLPtr = &imageURL
	}

	// Traiter les tags
	var tags []string
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		// Nettoyer les tags
		for i, tag := range tags {
			tags[i] = strings.TrimSpace(tag)
		}
		// Filtrer les tags vides
		var filteredTags []string
		for _, tag := range tags {
			if tag != "" {
				filteredTags = append(filteredTags, tag)
			}
		}
		tags = filteredTags
	}

	// Créer le DTO de mise à jour
	updateDTO := services.UpdateThreadDTO{
		Title:       title,
		Description: description,
		ImageURL:    imageURLPtr,
		Tags:        tags,
		State:       state,
		Visibility:  visibility,
	}

	// Mettre à jour le thread
	err = threadService.UpdateThread(threadID, updateDTO, user.ID, user.IsAdmin)
	if err != nil {
		log.Printf("❌ Erreur mise à jour thread %d: %v", threadID, err)
		http.Redirect(w, r, fmt.Sprintf("/thread/%d/edit?error=update_failed", threadID), http.StatusSeeOther)
		return
	}

	log.Printf("✅ Thread %d mis à jour avec succès par %s", threadID, user.Username)
	http.Redirect(w, r, fmt.Sprintf("/thread/%d?success=thread_updated", threadID), http.StatusSeeOther)
}

// convertThreadResponseToPageThread convertit une réponse thread en thread de page
func convertThreadResponseToPageThread(threadResp services.ThreadResponseDTO, user *User) *Thread {
	// Générer les initiales
	initials := generateInitials(threadResp.Author.Username)

	// Parser la date
	createdAt, err := time.Parse("2006-01-02T15:04:05Z", threadResp.CreatedAt)
	if err != nil {
		createdAt = time.Now()
	}
	timeAgo := formatTimeAgo(createdAt)

	// Déterminer l'auteur à afficher
	authorDisplay := threadResp.Author.Username
	if user != nil && user.Username == threadResp.Author.Username {
		authorDisplay = "YOU"
	}

	// Convertir les tags
	tags := make([]string, len(threadResp.Tags))
	for i, tag := range threadResp.Tags {
		tags[i] = tag.Name
	}

	return &Thread{
		ID:           threadResp.ID,
		Title:        threadResp.Title,
		Content:      threadResp.Description,
		ImageURL:     threadResp.ImageURL,
		Author:       authorDisplay,
		AuthorAvatar: initials,
		TimeAgo:      timeAgo,
		Genre:        "Discussion",
		Tags:         tags,
		Likes:        0, // Les likes seront récupérés si nécessaire
		IsLiked:      false,
		Comments:     threadResp.MessageCount,
		Shares:       0,
		Visibility:   threadResp.Visibility,
		State:        threadResp.State,
		MusicTrack:   nil,
	}
}
