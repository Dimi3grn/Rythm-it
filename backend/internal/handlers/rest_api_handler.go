// Fichier: backend/internal/handlers/rest_api_handler.go
package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/services"
	"rythmitbackend/pkg/database"
)

// ValidationAPIHandler valide les données côté serveur au lieu du JavaScript
func ValidationAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendAPIError(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	validationService := services.NewValidationService()
	var result services.ValidationResult

	switch requestData.Type {
	case "thread":
		data := services.ThreadValidationData{
			Content: getString(requestData.Data, "content"),
			Tags:    getStringArray(requestData.Data, "tags"),
			Genre:   getString(requestData.Data, "genre"),
		}
		if imageURL := getStringPtr(requestData.Data, "image_url"); imageURL != nil {
			data.ImageURL = imageURL
		}
		result = validationService.ValidateThread(data)

	case "comment":
		data := services.CommentValidationData{
			Content: getString(requestData.Data, "content"),
		}
		if imageURL := getStringPtr(requestData.Data, "image_url"); imageURL != nil {
			data.ImageURL = imageURL
		}
		result = validationService.ValidateComment(data)

	case "user":
		data := services.UserValidationData{
			Username: getString(requestData.Data, "username"),
			Email:    getString(requestData.Data, "email"),
			Password: getString(requestData.Data, "password"),
		}
		result = validationService.ValidateUser(data)

	default:
		sendAPIError(w, "Type de validation non supporté", http.StatusBadRequest)
		return
	}

	sendAPISuccess(w, "Validation effectuée", map[string]interface{}{
		"is_valid": result.IsValid,
		"errors":   result.Errors,
	})
}

// FormProcessingAPIHandler traite les formulaires côté serveur
func FormProcessingAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	// Vérifier l'authentification
	user, isLoggedIn := getUserFromCookie(r)
	if !isLoggedIn {
		sendAPIError(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	var requestData struct {
		FormType string                 `json:"form_type"`
		Data     map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendAPIError(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	switch requestData.FormType {
	case "thread_create":
		handleThreadCreateForm(w, requestData.Data, user)
	case "comment_create":
		handleCommentCreateForm(w, requestData.Data, user)
	case "profile_update":
		handleProfileUpdateForm(w, requestData.Data, user)
	default:
		sendAPIError(w, "Type de formulaire non supporté", http.StatusBadRequest)
	}
}

// PreprocessDataAPIHandler prétraite les données côté serveur
func PreprocessDataAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	var requestData struct {
		Type string                 `json:"type"`
		Data map[string]interface{} `json:"data"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		sendAPIError(w, "Données JSON invalides", http.StatusBadRequest)
		return
	}

	validationService := services.NewValidationService()

	switch requestData.Type {
	case "sanitize_content":
		content := getString(requestData.Data, "content")
		sanitized := validationService.SanitizeInput(content)
		sendAPISuccess(w, "Contenu nettoyé", map[string]interface{}{
			"original":  content,
			"sanitized": sanitized,
		})

	case "generate_title":
		content := getString(requestData.Data, "content")
		title := generateTitleFromContent(content)
		sendAPISuccess(w, "Titre généré", map[string]interface{}{
			"title": title,
		})

	case "process_tags":
		tags := getStringArray(requestData.Data, "tags")
		processedTags := processTags(tags, validationService)
		sendAPISuccess(w, "Tags traités", map[string]interface{}{
			"tags": processedTags,
		})

	default:
		sendAPIError(w, "Type de prétraitement non supporté", http.StatusBadRequest)
	}
}

// Fonctions helpers pour traiter les formulaires

func handleThreadCreateForm(w http.ResponseWriter, data map[string]interface{}, user *User) {
	// Validation
	validationService := services.NewValidationService()
	threadData := services.ThreadValidationData{
		Content: getString(data, "content"),
		Tags:    getStringArray(data, "tags"),
		Genre:   getString(data, "genre"),
	}
	if imageURL := getStringPtr(data, "image_url"); imageURL != nil {
		threadData.ImageURL = imageURL
	}

	// Valider et nettoyer
	sanitizedData, validationResult := validationService.ValidateAndSanitize(threadData)
	if !validationResult.IsValid {
		sendAPIError(w, validationResult.Errors[0].Message, http.StatusBadRequest)
		return
	}

	cleanData := sanitizedData.(services.ThreadValidationData)

	// Simulation de création (à remplacer par la vraie logique)
	sendAPISuccess(w, "Thread créé avec succès", map[string]interface{}{
		"id":         time.Now().Unix(), // ID simulé
		"title":      generateTitleFromContent(cleanData.Content),
		"content":    cleanData.Content,
		"tags":       cleanData.Tags,
		"genre":      cleanData.Genre,
		"image_url":  cleanData.ImageURL,
		"author":     user.Username,
		"created_at": time.Now(),
	})
}

func handleCommentCreateForm(w http.ResponseWriter, data map[string]interface{}, user *User) {
	// Validation
	validationService := services.NewValidationService()
	commentData := services.CommentValidationData{
		Content: getString(data, "content"),
	}
	if imageURL := getStringPtr(data, "image_url"); imageURL != nil {
		commentData.ImageURL = imageURL
	}

	// Valider et nettoyer
	sanitizedData, validationResult := validationService.ValidateAndSanitize(commentData)
	if !validationResult.IsValid {
		sendAPIError(w, validationResult.Errors[0].Message, http.StatusBadRequest)
		return
	}

	cleanData := sanitizedData.(services.CommentValidationData)

	// Simulation de création (à remplacer par la vraie logique)
	sendAPISuccess(w, "Commentaire créé avec succès", map[string]interface{}{
		"id":         time.Now().Unix(), // ID simulé
		"content":    cleanData.Content,
		"image_url":  cleanData.ImageURL,
		"author":     user.Username,
		"created_at": time.Now(),
	})
}

func handleProfileUpdateForm(w http.ResponseWriter, data map[string]interface{}, user *User) {
	// Validation
	validationService := services.NewValidationService()
	userData := services.UserValidationData{
		Username: getString(data, "display_name"),
		Email:    user.Email, // Garder l'email existant
	}

	result := validationService.ValidateUser(userData)
	if !result.IsValid {
		sendAPIError(w, result.Errors[0].Message, http.StatusBadRequest)
		return
	}

	// Simulation de mise à jour (à remplacer par la vraie logique)
	sendAPISuccess(w, "Profil mis à jour avec succès", map[string]interface{}{
		"display_name": validationService.SanitizeInput(userData.Username),
		"updated_at":   time.Now(),
	})
}

// Fonctions utilitaires

func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func getStringPtr(data map[string]interface{}, key string) *string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok && str != "" {
			return &str
		}
	}
	return nil
}

func getStringArray(data map[string]interface{}, key string) []string {
	if val, ok := data[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

func processTags(tags []string, vs *services.ValidationService) []string {
	processed := make([]string, 0, len(tags))
	for _, tag := range tags {
		cleaned := vs.SanitizeInput(tag)
		if cleaned != "" {
			processed = append(processed, cleaned)
		}
	}
	return processed
}

// SearchAPIHandler remplace la logique de recherche JavaScript
func SearchAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	searchType := r.URL.Query().Get("type")
	limitStr := r.URL.Query().Get("limit")

	// Récupérer les tags depuis les paramètres tags[]
	tags := r.URL.Query()["tags[]"]

	limit := 10
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Validation de la recherche
	validationService := services.NewValidationService()
	query = validationService.SanitizeInput(query)

	// Si pas de query et pas de tags, erreur
	if len(query) < 2 && len(tags) == 0 {
		sendAPIError(w, "La recherche doit contenir au moins 2 caractères ou des tags", http.StatusBadRequest)
		return
	}

	// Nettoyer les tags
	var cleanTags []string
	for _, tag := range tags {
		cleanTag := validationService.SanitizeInput(tag)
		if cleanTag != "" {
			cleanTags = append(cleanTags, cleanTag)
		}
	}

	// Simulation de résultats de recherche (à remplacer par la vraie logique)
	var results []map[string]interface{}

	switch searchType {
	case "tags":
		results = simulateTagSearch(query, limit)
	case "users":
		results = simulateUserSearch(query, limit)
	case "threads":
		results = simulateThreadSearch(query, cleanTags, limit)
	default:
		results = simulateGlobalSearch(query, cleanTags, limit)
	}

	sendAPISuccess(w, "Recherche effectuée", map[string]interface{}{
		"query":   query,
		"type":    searchType,
		"tags":    cleanTags,
		"results": results,
		"count":   len(results),
	})
}

// Fonctions de simulation de recherche (à remplacer par la vraie logique)

func simulateTagSearch(query string, limit int) []map[string]interface{} {
	mockTags := []string{"electronic", "ambient", "house", "techno", "jazz", "rock", "pop", "hip-hop"}
	results := make([]map[string]interface{}, 0)

	for _, tag := range mockTags {
		if len(results) >= limit {
			break
		}
		if containsIgnoreCase(tag, query) {
			results = append(results, map[string]interface{}{
				"type":  "tag",
				"name":  tag,
				"count": 42, // Nombre fictif de threads avec ce tag
			})
		}
	}

	return results
}

func simulateUserSearch(query string, limit int) []map[string]interface{} {
	// Simulation - à remplacer par une vraie recherche en base
	return []map[string]interface{}{
		{
			"type":     "user",
			"id":       1,
			"username": "TestUser",
			"avatar":   "/uploads/avatars/default.png",
		},
	}
}

func simulateThreadSearch(query string, tags []string, limit int) []map[string]interface{} {
	// Utiliser le vrai service de recherche
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	params := models.PaginationParams{
		Page:    1,
		PerPage: limit,
		Sort:    "created_at",
		Order:   "DESC",
	}

	// Utiliser la nouvelle méthode qui supporte les tags
	result, err := threadService.SearchThreadsWithTags(query, tags, params)
	if err != nil {
		log.Printf("❌ Erreur recherche threads avec tags: %v", err)
		return []map[string]interface{}{}
	}

	results := make([]map[string]interface{}, 0, len(result.Threads))
	for _, thread := range result.Threads {
		// Convertir les tags en slice de strings
		threadTags := make([]string, len(thread.Tags))
		for i, tag := range thread.Tags {
			threadTags[i] = tag.Name
		}

		results = append(results, map[string]interface{}{
			"type":          "thread",
			"id":            thread.ID,
			"title":         thread.Title,
			"content":       thread.Description,
			"description":   thread.Description,
			"author":        thread.Author.Username,
			"username":      thread.Author.Username,
			"created_at":    thread.CreatedAt,
			"image_url":     thread.ImageURL,
			"tags":          threadTags,
			"likes":         0, // TODO: récupérer depuis le repository des likes
			"comments":      thread.MessageCount,
			"message_count": thread.MessageCount,
		})
	}

	return results
}

func simulateGlobalSearch(query string, tags []string, limit int) []map[string]interface{} {
	results := make([]map[string]interface{}, 0)

	// Combiner les résultats de différents types
	tagResults := simulateTagSearch(query, 3)
	users := simulateUserSearch(query, 3)
	threads := simulateThreadSearch(query, tags, 4)

	results = append(results, tagResults...)
	results = append(results, users...)
	results = append(results, threads...)

	if len(results) > limit {
		results = results[:limit]
	}

	return results
}

func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// ThreadSearchAPIHandler recherche spécifiquement dans les threads
func ThreadSearchAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		sendAPIError(w, "Méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}

	query := r.URL.Query().Get("q")
	tagsParam := r.URL.Query().Get("tags")
	limitStr := r.URL.Query().Get("limit")

	// Traiter les tags
	var tags []string
	if tagsParam != "" {
		tags = strings.Split(tagsParam, ",")
		// Nettoyer les tags
		cleanTags := make([]string, 0, len(tags))
		for _, tag := range tags {
			cleanTag := strings.TrimSpace(tag)
			if cleanTag != "" {
				cleanTags = append(cleanTags, cleanTag)
			}
		}
		tags = cleanTags
	}

	limit := 20
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 50 {
			limit = l
		}
	}

	// Validation de la recherche
	validationService := services.NewValidationService()
	query = validationService.SanitizeInput(query)

	// Si pas de query et pas de tags, erreur
	if (len(query) < 2 && len(tags) == 0) || (len(query) > 0 && len(query) < 2) {
		sendAPIError(w, "La recherche doit contenir au moins 2 caractères ou des tags", http.StatusBadRequest)
		return
	}

	log.Printf("🔍 Recherche threads: query='%s', tags=%v, limit=%d", query, tags, limit)

	// Utiliser le vrai service de recherche
	db := database.DB
	threadRepo := repositories.NewThreadRepository(db)
	tagRepo := repositories.NewTagRepository(db)
	messageRepo := repositories.NewMessageRepository(db)
	threadService := services.NewThreadService(threadRepo, tagRepo, messageRepo, db)

	params := models.PaginationParams{
		Page:    1,
		PerPage: limit,
		Sort:    "created_at",
		Order:   "DESC",
	}

	// Utiliser la méthode de recherche avec tags
	result, err := threadService.SearchThreadsWithTags(query, tags, params)
	if err != nil {
		log.Printf("❌ Erreur recherche threads: %v", err)
		sendAPIError(w, "Erreur lors de la recherche", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ %d threads trouvés", len(result.Threads))

	// Convertir les threads au format attendu par le frontend
	threads := make([]map[string]interface{}, 0, len(result.Threads))
	for _, thread := range result.Threads {
		// Convertir les tags en slice de strings
		threadTags := make([]string, len(thread.Tags))
		for i, tag := range thread.Tags {
			threadTags[i] = tag.Name
		}

		// Récupérer le nombre réel de likes
		likeRepo := repositories.NewLikeRepository(db)
		likesCount, err := likeRepo.GetThreadLikesCount(thread.ID)
		if err != nil {
			log.Printf("❌ Erreur récupération likes pour thread %d: %v", thread.ID, err)
			likesCount = 0
		}

		threads = append(threads, map[string]interface{}{
			"id":            thread.ID,
			"ID":            thread.ID, // Compatibilité
			"title":         thread.Title,
			"Title":         thread.Title, // Compatibilité
			"content":       thread.Description,
			"Content":       thread.Description, // Compatibilité
			"description":   thread.Description,
			"author":        thread.Author.Username,
			"Author":        thread.Author.Username, // Compatibilité
			"username":      thread.Author.Username,
			"created_at":    thread.CreatedAt,
			"CreatedAt":     thread.CreatedAt, // Compatibilité
			"image_url":     thread.ImageURL,
			"tags":          threadTags,
			"likes":         likesCount, // Nombre réel de likes
			"Likes":         likesCount, // Compatibilité
			"comments":      thread.MessageCount,
			"Comments":      thread.MessageCount, // Compatibilité
			"message_count": thread.MessageCount,
		})
	}

	// Réponse au format attendu par le frontend
	response := map[string]interface{}{
		"threads": threads,
		"count":   len(threads),
		"total":   result.Pagination.Total,
		"query":   query,
		"tags":    tags,
	}

	sendAPISuccess(w, "Recherche effectuée", response)
}
