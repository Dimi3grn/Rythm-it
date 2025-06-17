package services

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"
	"strings"
	"time"
)

// ThreadDTO represents a thread for the frontend
type ThreadDTO struct {
	ID           uint      `json:"id"`
	Title        string    `json:"title"`
	Content      string    `json:"content"`
	UserID       uint      `json:"user_id"`
	Username     string    `json:"username"`
	MessageCount int       `json:"message_count"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	Tags         []string  `json:"tags"`
}

// ThreadService interface pour la logique métier des threads
type ThreadService interface {
	CreateThread(dto CreateThreadDTO, userID uint) (*ThreadResponseDTO, error)
	GetThread(id uint, userID *uint) (*ThreadResponseDTO, error)
	GetPublicThreads(params models.PaginationParams, filters ThreadFilters) (*PaginatedThreadsResponseDTO, error)
	GetUserThreads(userID uint, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error)
	UpdateThread(id uint, dto UpdateThreadDTO, userID uint, isAdmin bool) error
	DeleteThread(id uint, userID uint, isAdmin bool) error
	ChangeThreadState(id uint, state string, userID uint, isAdmin bool) error
	SearchThreads(query string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error)
	SearchThreadsWithTags(query string, tags []string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error)
	GetThreadsByTag(tagName string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error)
	GetAllThreads() ([]ThreadDTO, error)
}

// DTOs pour les threads
type CreateThreadDTO struct {
	Title       string   `json:"title" validate:"required,min=5,max=200"`
	Description string   `json:"description" validate:"required,min=10"`
	ImageURL    *string  `json:"image_url" validate:"omitempty"`
	Tags        []string `json:"tags" validate:"required,min=1,max=10"`
	Visibility  string   `json:"visibility" validate:"oneof=public privé"`
}

type UpdateThreadDTO struct {
	Title       string   `json:"title" validate:"required,min=5,max=200"`
	Description string   `json:"description" validate:"required,min=10"`
	ImageURL    *string  `json:"image_url" validate:"omitempty"`
	Tags        []string `json:"tags" validate:"omitempty,max=10"`
	State       string   `json:"state" validate:"oneof=ouvert fermé archivé"`
	Visibility  string   `json:"visibility" validate:"oneof=public privé"`
}

type ThreadResponseDTO struct {
	ID           uint             `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
	ImageURL     *string          `json:"image_url"`
	State        string           `json:"state"`
	Visibility   string           `json:"visibility"`
	CreatedAt    string           `json:"created_at"`
	UpdatedAt    string           `json:"updated_at"`
	Author       UserSummaryDTO   `json:"author"`
	Tags         []TagResponseDTO `json:"tags"`
	MessageCount int              `json:"message_count"`
	FireCount    int              `json:"fire_count"`
	SkipCount    int              `json:"skip_count"`
	UserVote     *string          `json:"user_vote,omitempty"` // pour les threads avec votes
}

type TagResponseDTO struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

type UserSummaryDTO struct {
	ID         uint    `json:"id"`
	Username   string  `json:"username"`
	ProfilePic *string `json:"profile_pic"`
}

type PaginatedThreadsResponseDTO struct {
	Threads    []ThreadResponseDTO `json:"threads"`
	Pagination PaginationInfo      `json:"pagination"`
}

type PaginationInfo struct {
	Page       int   `json:"page"`
	PerPage    int   `json:"per_page"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"total_pages"`
}

type ThreadFilters struct {
	TagName string `json:"tag_name"`
	State   string `json:"state"`
	UserID  *uint  `json:"user_id"`
}

// threadService implémentation
type threadService struct {
	threadRepo  repositories.ThreadRepository
	tagRepo     repositories.TagRepository
	messageRepo repositories.MessageRepository
	db          *sql.DB
}

// NewThreadService crée une nouvelle instance du service
func NewThreadService(threadRepo repositories.ThreadRepository, tagRepo repositories.TagRepository, messageRepo repositories.MessageRepository, db *sql.DB) ThreadService {
	return &threadService{
		threadRepo:  threadRepo,
		tagRepo:     tagRepo,
		messageRepo: messageRepo,
		db:          db,
	}
}

// CreateThread crée un nouveau thread avec ses tags
func (s *threadService) CreateThread(dto CreateThreadDTO, userID uint) (*ThreadResponseDTO, error) {
	// Validation du DTO
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return nil, fmt.Errorf("erreur validation: %v", validationErrors)
	}

	// Valeurs par défaut
	if dto.Visibility == "" {
		dto.Visibility = models.VisibilityPublic
	}

	// Créer le thread
	thread := &models.Thread{
		Title:       strings.TrimSpace(dto.Title),
		Description: strings.TrimSpace(dto.Description),
		ImageURL:    dto.ImageURL,
		State:       models.ThreadStateOpen,
		Visibility:  dto.Visibility,
		UserID:      userID,
	}

	// Transaction pour créer le thread et ses tags
	err := s.threadRepo.Transaction(func(tx *sql.Tx) error {
		// Créer le thread
		if err := s.threadRepo.Create(thread); err != nil {
			return fmt.Errorf("erreur création thread: %w", err)
		}

		// Traiter les tags
		var tagIDs []uint
		for _, tagName := range dto.Tags {
			tagName = strings.TrimSpace(tagName)
			if tagName == "" {
				continue
			}

			// Déterminer le type de tag basé sur le nom (heuristique simple)
			tagType := determineTagType(tagName)

			// FindOrCreate pour chaque tag
			tag, err := s.tagRepo.FindOrCreate(tagName, tagType)
			if err != nil {
				return fmt.Errorf("erreur gestion tag '%s': %w", tagName, err)
			}

			tagIDs = append(tagIDs, tag.ID)
		}

		// Attacher les tags au thread
		if len(tagIDs) > 0 {
			if err := s.threadRepo.AttachTags(thread.ID, tagIDs); err != nil {
				return fmt.Errorf("erreur attachement tags: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Récupérer le thread complet pour la réponse
	return s.GetThread(thread.ID, &userID)
}

// GetThread récupère un thread par son ID
func (s *threadService) GetThread(id uint, userID *uint) (*ThreadResponseDTO, error) {
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrThreadNotFound
	}

	// Vérifier les permissions pour les threads privés
	if thread.Visibility == models.VisibilityPrivate {
		if userID == nil || (*userID != thread.UserID) {
			return nil, utils.ErrUnauthorized
		}
	}

	// Vérifier si le thread est archivé (sauf pour le propriétaire et admin)
	if thread.State == models.ThreadStateArchived {
		if userID == nil || (*userID != thread.UserID) {
			// TODO: Vérifier si user est admin
			return nil, utils.ErrThreadArchived
		}
	}

	return s.threadToDTO(thread), nil
}

// ValidatePagination valide et normalise les paramètres de pagination
func ValidatePagination(params *models.PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}
	if params.PerPage < 1 {
		params.PerPage = 10
	} else if params.PerPage > 100 {
		params.PerPage = 100
	}
	if params.Sort == "" {
		params.Sort = "created_at"
	}
	if params.Order != "ASC" && params.Order != "DESC" {
		params.Order = "DESC"
	}
}

// GetPublicThreads récupère les threads publics avec pagination et filtres
func (s *threadService) GetPublicThreads(params models.PaginationParams, filters ThreadFilters) (*PaginatedThreadsResponseDTO, error) {
	ValidatePagination(&params)

	var threads []*models.Thread
	var total int64
	var err error

	// Appliquer les filtres
	if filters.TagName != "" {
		// Recherche par tag
		tag, tagErr := s.tagRepo.FindByName(filters.TagName)
		if tagErr != nil {
			return &PaginatedThreadsResponseDTO{
				Threads:    []ThreadResponseDTO{},
				Pagination: s.buildPaginationInfo(params, 0),
			}, nil
		}
		threads, total, err = s.threadRepo.FindByTag(tag.ID, params)
	} else {
		// Liste normale
		threads, total, err = s.threadRepo.FindPublicThreads(params)
	}

	if err != nil {
		return nil, fmt.Errorf("erreur récupération threads: %w", err)
	}

	// Convertir en DTOs
	var threadDTOs []ThreadResponseDTO
	for _, thread := range threads {
		threadDTOs = append(threadDTOs, *s.threadToDTO(thread))
	}

	return &PaginatedThreadsResponseDTO{
		Threads:    threadDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// GetUserThreads récupère les threads d'un utilisateur
func (s *threadService) GetUserThreads(userID uint, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error) {
	ValidatePagination(&params)

	threads, err := s.threadRepo.FindByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération threads utilisateur: %w", err)
	}

	// Convertir en DTOs
	var threadDTOs []ThreadResponseDTO
	for _, thread := range threads {
		threadDTOs = append(threadDTOs, *s.threadToDTO(thread))
	}

	// Pagination manuelle pour les threads utilisateur
	total := int64(len(threadDTOs))
	start := (params.Page - 1) * params.PerPage
	end := start + params.PerPage

	if start > len(threadDTOs) {
		threadDTOs = []ThreadResponseDTO{}
	} else {
		if end > len(threadDTOs) {
			end = len(threadDTOs)
		}
		threadDTOs = threadDTOs[start:end]
	}

	return &PaginatedThreadsResponseDTO{
		Threads:    threadDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// UpdateThread met à jour un thread
func (s *threadService) UpdateThread(id uint, dto UpdateThreadDTO, userID uint, isAdmin bool) error {
	// Validation
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return fmt.Errorf("erreur validation: %v", validationErrors)
	}

	// Récupérer le thread
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return utils.ErrThreadNotFound
	}

	// Vérifier les permissions
	if !isAdmin && thread.UserID != userID {
		return utils.ErrUnauthorized
	}

	// Transaction pour mettre à jour le thread et ses tags
	return s.threadRepo.Transaction(func(tx *sql.Tx) error {
		// Mettre à jour les champs du thread
		thread.Title = strings.TrimSpace(dto.Title)
		thread.Description = strings.TrimSpace(dto.Description)
		thread.ImageURL = dto.ImageURL
		thread.State = dto.State
		thread.Visibility = dto.Visibility

		// Mettre à jour le thread
		if err := s.threadRepo.Update(thread); err != nil {
			return fmt.Errorf("erreur mise à jour thread: %w", err)
		}

		// Traiter les tags si fournis
		if len(dto.Tags) > 0 {
			var tagIDs []uint
			for _, tagName := range dto.Tags {
				tagName = strings.TrimSpace(tagName)
				if tagName == "" {
					continue
				}

				// Déterminer le type de tag
				tagType := determineTagType(tagName)

				// FindOrCreate pour chaque tag
				tag, err := s.tagRepo.FindOrCreate(tagName, tagType)
				if err != nil {
					return fmt.Errorf("erreur gestion tag '%s': %w", tagName, err)
				}

				tagIDs = append(tagIDs, tag.ID)
			}

			// Mettre à jour les tags du thread
			if len(tagIDs) > 0 {
				if err := s.threadRepo.AttachTags(thread.ID, tagIDs); err != nil {
					return fmt.Errorf("erreur mise à jour tags: %w", err)
				}
			}
		}

		return nil
	})
}

// DeleteThread supprime un thread
func (s *threadService) DeleteThread(id uint, userID uint, isAdmin bool) error {
	// Récupérer le thread
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return utils.ErrThreadNotFound
	}

	// Vérifier les permissions
	if !isAdmin && thread.UserID != userID {
		return utils.ErrUnauthorized
	}

	return s.threadRepo.Delete(id)
}

// ChangeThreadState change l'état d'un thread (admin ou propriétaire)
func (s *threadService) ChangeThreadState(id uint, state string, userID uint, isAdmin bool) error {
	// Valider l'état
	if state != models.ThreadStateOpen && state != models.ThreadStateClosed && state != models.ThreadStateArchived {
		return fmt.Errorf("état invalide: %s", state)
	}

	// Récupérer le thread
	thread, err := s.threadRepo.FindByID(id)
	if err != nil {
		return utils.ErrThreadNotFound
	}

	// Vérifier les permissions
	if !isAdmin && thread.UserID != userID {
		return utils.ErrUnauthorized
	}

	return s.threadRepo.UpdateState(id, state)
}

// SearchThreads recherche des threads par titre
func (s *threadService) SearchThreads(query string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error) {
	ValidatePagination(&params)

	query = strings.TrimSpace(query)
	if query == "" {
		return s.GetPublicThreads(params, ThreadFilters{})
	}

	threads, total, err := s.threadRepo.Search(query, params)
	if err != nil {
		return nil, fmt.Errorf("erreur recherche threads: %w", err)
	}

	// Convertir en DTOs
	var threadDTOs []ThreadResponseDTO
	for _, thread := range threads {
		threadDTOs = append(threadDTOs, *s.threadToDTO(thread))
	}

	return &PaginatedThreadsResponseDTO{
		Threads:    threadDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// SearchThreadsWithTags recherche des threads par titre et/ou tags
func (s *threadService) SearchThreadsWithTags(query string, tags []string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error) {
	ValidatePagination(&params)

	query = strings.TrimSpace(query)

	// Si pas de query ni de tags, retourner tous les threads publics
	if query == "" && len(tags) == 0 {
		return s.GetPublicThreads(params, ThreadFilters{})
	}

	// Si seulement des tags, utiliser la nouvelle méthode optimisée
	if query == "" && len(tags) > 0 {
		threads, total, err := s.threadRepo.FindByTags(tags, params)
		if err != nil {
			return nil, fmt.Errorf("erreur recherche par tags: %w", err)
		}

		var threadDTOs []ThreadResponseDTO
		for _, thread := range threads {
			threadDTOs = append(threadDTOs, *s.threadToDTO(thread))
		}

		return &PaginatedThreadsResponseDTO{
			Threads:    threadDTOs,
			Pagination: s.buildPaginationInfo(params, total),
		}, nil
	}

	// Si seulement une query, utiliser la recherche normale
	if query != "" && len(tags) == 0 {
		return s.SearchThreads(query, params)
	}

	// Recherche combinée query + tags - utiliser la nouvelle méthode optimisée
	threads, total, err := s.threadRepo.SearchWithTags(query, tags, params)
	if err != nil {
		return nil, fmt.Errorf("erreur recherche combinée: %w", err)
	}

	// Convertir en DTOs
	var threadDTOs []ThreadResponseDTO
	for _, thread := range threads {
		threadDTOs = append(threadDTOs, *s.threadToDTO(thread))
	}

	return &PaginatedThreadsResponseDTO{
		Threads:    threadDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// GetThreadsByTag récupère les threads d'un tag spécifique
func (s *threadService) GetThreadsByTag(tagName string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error) {
	filters := ThreadFilters{TagName: tagName}
	return s.GetPublicThreads(params, filters)
}

// Helper methods

// threadToDTO convertit un thread en DTO
func (s *threadService) threadToDTO(thread *models.Thread) *ThreadResponseDTO {
	dto := &ThreadResponseDTO{
		ID:          thread.ID,
		Title:       thread.Title,
		Description: thread.Description,
		ImageURL:    thread.ImageURL,
		State:       thread.State,
		Visibility:  thread.Visibility,
		CreatedAt:   thread.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   thread.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		Author: UserSummaryDTO{
			ID:         thread.Author.ID,
			Username:   thread.Author.Username,
			ProfilePic: thread.Author.ProfilePic,
		},
		Tags:         []TagResponseDTO{},
		MessageCount: 0, // TODO: Implémenter le comptage des messages
		FireCount:    thread.FireCount,
		SkipCount:    thread.SkipCount,
	}

	// Convertir les tags
	for _, tag := range thread.Tags {
		dto.Tags = append(dto.Tags, TagResponseDTO{
			ID:   tag.ID,
			Name: tag.Name,
			Type: tag.Type,
		})
	}

	return dto
}

// buildPaginationInfo construit les infos de pagination
func (s *threadService) buildPaginationInfo(params models.PaginationParams, total int64) PaginationInfo {
	totalPages := int(total) / params.PerPage
	if int(total)%params.PerPage > 0 {
		totalPages++
	}

	return PaginationInfo{
		Page:       params.Page,
		PerPage:    params.PerPage,
		Total:      total,
		TotalPages: totalPages,
	}
}

// determineTagType détermine le type d'un tag basé sur son nom (heuristique)
func determineTagType(tagName string) string {
	tagName = strings.ToLower(tagName)

	// Liste des genres connus
	genres := []string{
		"rap", "hip-hop", "pop", "rock", "jazz", "electronic", "indie",
		"metal", "classical", "r&b", "reggae", "country", "folk", "blues",
		"techno", "house", "dubstep", "ambient", "punk", "alternative",
	}

	// Vérifier si c'est un genre connu
	for _, genre := range genres {
		if tagName == genre {
			return "genre"
		}
	}

	// Par défaut, considérer comme un artiste
	// Une amélioration future pourrait utiliser une API externe pour déterminer cela
	return "artist"
}

// GetAllThreads retrieves all threads with their tags
func (s *threadService) GetAllThreads() ([]ThreadDTO, error) {
	// Use default pagination params
	params := models.PaginationParams{
		Page:    1,
		PerPage: 100, // Get a reasonable number of threads
		Sort:    "created_at",
		Order:   "DESC",
	}

	threads, total, err := s.threadRepo.FindAll(params)
	if err != nil {
		return nil, err
	}

	if total == 0 {
		return []ThreadDTO{}, nil
	}

	var dtos []ThreadDTO
	for _, thread := range threads {
		// Get tags from the thread's Tags field since they should be preloaded
		tagNames := make([]string, len(thread.Tags))
		for i, tag := range thread.Tags {
			tagNames[i] = tag.Name
		}

		// Get message count from the message repository
		messageCount64, err := s.messageRepo.CountByThreadID(thread.ID)
		if err != nil {
			// Log the error but continue with 0 count
			fmt.Printf("Error getting message count for thread %d: %v\n", thread.ID, err)
			messageCount64 = 0
		}

		// Safely convert int64 to int
		messageCount := 0
		if messageCount64 <= int64(^uint(0)>>1) { // Check if it fits in an int
			messageCount = int(messageCount64)
		}

		dto := ThreadDTO{
			ID:           thread.ID,
			Title:        thread.Title,
			Content:      thread.Description,
			UserID:       thread.UserID,
			Username:     thread.Author.Username,
			MessageCount: messageCount,
			CreatedAt:    thread.CreatedAt,
			UpdatedAt:    thread.UpdatedAt,
			Tags:         tagNames,
		}
		dtos = append(dtos, dto)
	}

	return dtos, nil
}
