package services

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"
	"strings"
)

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
	GetThreadsByTag(tagName string, params models.PaginationParams) (*PaginatedThreadsResponseDTO, error)
}

// DTOs pour les threads
type CreateThreadDTO struct {
	Title       string   `json:"title" validate:"required,min=5,max=200"`
	Description string   `json:"description" validate:"required,min=10"`
	Tags        []string `json:"tags" validate:"required,min=1,max=10"`
	Visibility  string   `json:"visibility" validate:"oneof=public privé"`
}

type UpdateThreadDTO struct {
	Title       string `json:"title" validate:"required,min=5,max=200"`
	Description string `json:"description" validate:"required,min=10"`
	State       string `json:"state" validate:"oneof=ouvert fermé archivé"`
	Visibility  string `json:"visibility" validate:"oneof=public privé"`
}

type ThreadResponseDTO struct {
	ID           uint             `json:"id"`
	Title        string           `json:"title"`
	Description  string           `json:"description"`
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
	threadRepo repositories.ThreadRepository
	tagRepo    repositories.TagRepository
	db         *sql.DB
}

// NewThreadService crée une nouvelle instance du service
func NewThreadService(threadRepo repositories.ThreadRepository, tagRepo repositories.TagRepository, db *sql.DB) ThreadService {
	return &threadService{
		threadRepo: threadRepo,
		tagRepo:    tagRepo,
		db:         db,
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

	// Mettre à jour les champs
	thread.Title = strings.TrimSpace(dto.Title)
	thread.Description = strings.TrimSpace(dto.Description)
	thread.State = dto.State
	thread.Visibility = dto.Visibility

	return s.threadRepo.Update(thread)
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
