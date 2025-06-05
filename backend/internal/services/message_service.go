package services

import (
	"database/sql"
	"fmt"
	"regexp"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
	"rythmitbackend/internal/utils"
	"strings"
)

// MessageService interface pour la logique métier des messages
type MessageService interface {
	PostMessage(dto CreateMessageDTO, threadID uint, userID uint) (*MessageResponseDTO, error)
	GetThreadMessages(threadID uint, userID *uint, params models.PaginationParams, sortBy string) (*PaginatedMessagesResponseDTO, error)
	GetMessage(id uint, userID *uint) (*MessageResponseDTO, error)
	UpdateMessage(id uint, content string, userID uint, isAdmin bool) error
	DeleteMessage(id uint, userID uint, isAdmin bool) error
	VoteMessage(messageID uint, userID uint, vote string) (*VoteResponseDTO, error)
	GetUserMessages(userID uint, params models.PaginationParams) (*PaginatedMessagesResponseDTO, error)
}

// DTOs pour les messages
type CreateMessageDTO struct {
	Content string `json:"content" validate:"required,min=1,max=5000"`
}

type MessageResponseDTO struct {
	ID              uint              `json:"id"`
	Content         string            `json:"content"`
	CreatedAt       string            `json:"created_at"`
	UpdatedAt       string            `json:"updated_at"`
	Author          UserSummaryDTO    `json:"author"`
	PopularityScore int               `json:"popularity_score"`
	FireCount       int               `json:"fire_count"`
	SkipCount       int               `json:"skip_count"`
	UserVote        *string           `json:"user_vote,omitempty"`
	Embeds          *MessageEmbedsDTO `json:"embeds,omitempty"`
}

type MessageEmbedsDTO struct {
	YouTube *string `json:"youtube,omitempty"`
	Spotify *string `json:"spotify,omitempty"`
}

type PaginatedMessagesResponseDTO struct {
	Messages   []MessageResponseDTO `json:"messages"`
	Pagination PaginationInfo       `json:"pagination"`
}

type VoteResponseDTO struct {
	MessageID       uint   `json:"message_id"`
	UserVote        string `json:"user_vote"`
	PopularityScore int    `json:"popularity_score"`
	FireCount       int    `json:"fire_count"`
	SkipCount       int    `json:"skip_count"`
}

// messageService implémentation
type messageService struct {
	messageRepo repositories.MessageRepository
	threadRepo  repositories.ThreadRepository
	db          *sql.DB
}

// NewMessageService crée une nouvelle instance du service
func NewMessageService(messageRepo repositories.MessageRepository, threadRepo repositories.ThreadRepository, db *sql.DB) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		threadRepo:  threadRepo,
		db:          db,
	}
}

// PostMessage crée un nouveau message dans un thread
func (s *messageService) PostMessage(dto CreateMessageDTO, threadID uint, userID uint) (*MessageResponseDTO, error) {
	// Validation du DTO
	if validationErrors := utils.ValidateStruct(dto); len(validationErrors) > 0 {
		return nil, fmt.Errorf("erreur validation: %v", validationErrors)
	}

	// Vérifier que le thread existe et est accessible
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, utils.ErrThreadNotFound
	}

	// Vérifier l'état du thread
	if thread.State == models.ThreadStateClosed {
		return nil, utils.ErrThreadClosed
	}

	if thread.State == models.ThreadStateArchived {
		return nil, utils.ErrThreadArchived
	}

	// Vérifier les permissions pour les threads privés
	if thread.Visibility == models.VisibilityPrivate && thread.UserID != userID {
		return nil, utils.ErrUnauthorized
	}

	// Nettoyer et traiter le contenu
	content := strings.TrimSpace(dto.Content)

	// Détecter les embeds (YouTube, Spotify)
	embeds := s.extractEmbeds(content)

	// Créer le message
	message := &models.Message{
		Content:  content,
		ThreadID: threadID,
		UserID:   userID,
		Embeds:   embeds,
	}

	err = s.messageRepo.Create(message)
	if err != nil {
		return nil, fmt.Errorf("erreur création message: %w", err)
	}

	// TODO: Incrémenter le compteur de messages de l'utilisateur
	// TODO: Mettre à jour la date de dernière activité du thread

	// Récupérer le message complet pour la réponse
	return s.GetMessage(message.ID, &userID)
}

// GetThreadMessages récupère les messages d'un thread avec pagination
func (s *messageService) GetThreadMessages(threadID uint, userID *uint, params models.PaginationParams, sortBy string) (*PaginatedMessagesResponseDTO, error) {
	ValidatePagination(&params)

	// Vérifier que le thread existe
	thread, err := s.threadRepo.FindByID(threadID)
	if err != nil {
		return nil, utils.ErrThreadNotFound
	}

	// Vérifier les permissions pour les threads privés
	if thread.Visibility == models.VisibilityPrivate {
		if userID == nil || *userID != thread.UserID {
			return nil, utils.ErrUnauthorized
		}
	}

	// Vérifier si le thread est archivé
	if thread.State == models.ThreadStateArchived {
		if userID == nil || (*userID != thread.UserID) {
			// TODO: Vérifier si user est admin
			return nil, utils.ErrThreadArchived
		}
	}

	// Valider le type de tri
	if sortBy == "" {
		sortBy = "date"
	}

	// Récupérer les messages avec les votes de l'utilisateur
	messages, total, err := s.messageRepo.GetMessagesWithVotes(threadID, userID, params, sortBy)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages: %w", err)
	}

	// Convertir en DTOs
	var messageDTOs []MessageResponseDTO
	for _, message := range messages {
		dto := s.messageToDTO(message)
		messageDTOs = append(messageDTOs, *dto)
	}

	return &PaginatedMessagesResponseDTO{
		Messages:   messageDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// GetMessage récupère un message par son ID
func (s *messageService) GetMessage(id uint, userID *uint) (*MessageResponseDTO, error) {
	message, err := s.messageRepo.FindByID(id)
	if err != nil {
		return nil, utils.ErrMessageNotFound
	}

	// Récupérer le vote de l'utilisateur si connecté
	if userID != nil {
		vote, voteErr := s.messageRepo.GetUserVote(*userID, id)
		if voteErr == nil && vote != models.VoteNeutral {
			message.UserVote = &vote
		}
	}

	return s.messageToDTO(message), nil
}

// UpdateMessage met à jour le contenu d'un message
func (s *messageService) UpdateMessage(id uint, content string, userID uint, isAdmin bool) error {
	// Récupérer le message
	message, err := s.messageRepo.FindByID(id)
	if err != nil {
		return utils.ErrMessageNotFound
	}

	// Vérifier les permissions
	if !isAdmin && message.UserID != userID {
		return utils.ErrUnauthorized
	}

	// Valider le nouveau contenu
	content = strings.TrimSpace(content)
	if len(content) < 1 || len(content) > 5000 {
		return fmt.Errorf("le contenu doit faire entre 1 et 5000 caractères")
	}

	// Détecter les nouveaux embeds
	embeds := s.extractEmbeds(content)

	// Mettre à jour le message
	message.Content = content
	message.Embeds = embeds

	return s.messageRepo.Update(message)
}

// DeleteMessage supprime un message
func (s *messageService) DeleteMessage(id uint, userID uint, isAdmin bool) error {
	// Récupérer le message
	message, err := s.messageRepo.FindByID(id)
	if err != nil {
		return utils.ErrMessageNotFound
	}

	// Vérifier les permissions
	if !isAdmin && message.UserID != userID {
		return utils.ErrUnauthorized
	}

	return s.messageRepo.Delete(id)
}

// VoteMessage gère les votes Fire/Skip sur un message
func (s *messageService) VoteMessage(messageID uint, userID uint, vote string) (*VoteResponseDTO, error) {
	// Valider le vote
	if vote != models.VoteFire && vote != models.VoteSkip && vote != models.VoteNeutral {
		return nil, fmt.Errorf("vote invalide: %s (attendu: fire, skip, ou neutral)", vote)
	}

	// Vérifier que le message existe
	message, err := s.messageRepo.FindByID(messageID)
	if err != nil {
		return nil, utils.ErrMessageNotFound
	}

	// Empêcher l'auto-vote (utilisateur qui vote pour son propre message)
	if message.UserID == userID {
		return nil, fmt.Errorf("vous ne pouvez pas voter pour votre propre message")
	}

	// Récupérer le vote actuel
	currentVote, err := s.messageRepo.GetUserVote(userID, messageID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération vote actuel: %w", err)
	}

	// Si c'est le même vote, le retirer (toggle)
	if currentVote == vote && vote != models.VoteNeutral {
		vote = models.VoteNeutral
	}

	// Enregistrer le vote
	err = s.messageRepo.SetUserVote(userID, messageID, vote)
	if err != nil {
		return nil, fmt.Errorf("erreur enregistrement vote: %w", err)
	}

	// Calculer les nouveaux scores
	score, err := s.messageRepo.GetPopularityScore(messageID)
	if err != nil {
		score = 0
	}

	fireCount, skipCount, err := s.messageRepo.GetMessageVoteCounts(messageID)
	if err != nil {
		fireCount, skipCount = 0, 0
	}

	return &VoteResponseDTO{
		MessageID:       messageID,
		UserVote:        vote,
		PopularityScore: score,
		FireCount:       fireCount,
		SkipCount:       skipCount,
	}, nil
}

// GetUserMessages récupère les messages d'un utilisateur
func (s *messageService) GetUserMessages(userID uint, params models.PaginationParams) (*PaginatedMessagesResponseDTO, error) {
	ValidatePagination(&params)

	messages, total, err := s.messageRepo.FindByUserID(userID, params)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages utilisateur: %w", err)
	}

	// Convertir en DTOs
	var messageDTOs []MessageResponseDTO
	for _, message := range messages {
		dto := s.messageToDTO(message)
		messageDTOs = append(messageDTOs, *dto)
	}

	return &PaginatedMessagesResponseDTO{
		Messages:   messageDTOs,
		Pagination: s.buildPaginationInfo(params, total),
	}, nil
}

// Helper methods

// messageToDTO convertit un message en DTO
func (s *messageService) messageToDTO(message *models.Message) *MessageResponseDTO {
	// Calculer les compteurs de votes si pas déjà fait
	fireCount, skipCount, _ := s.messageRepo.GetMessageVoteCounts(message.ID)

	dto := &MessageResponseDTO{
		ID:              message.ID,
		Content:         message.Content,
		CreatedAt:       message.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:       message.UpdatedAt.Format("2006-01-02T15:04:05Z"),
		PopularityScore: message.PopularityScore,
		FireCount:       fireCount,
		SkipCount:       skipCount,
		Author: UserSummaryDTO{
			ID:         message.Author.ID,
			Username:   message.Author.Username,
			ProfilePic: message.Author.ProfilePic,
		},
	}

	// Ajouter le vote utilisateur s'il existe
	if message.UserVote != nil {
		dto.UserVote = message.UserVote
	}

	// Ajouter les embeds s'ils existent
	if message.Embeds != nil {
		dto.Embeds = &MessageEmbedsDTO{
			YouTube: message.Embeds.YouTube,
			Spotify: message.Embeds.Spotify,
		}
	}

	return dto
}

// buildPaginationInfo construit les infos de pagination
func (s *messageService) buildPaginationInfo(params models.PaginationParams, total int64) PaginationInfo {
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

// extractEmbeds détecte et extrait les liens YouTube et Spotify du contenu
func (s *messageService) extractEmbeds(content string) *models.MessageEmbeds {
	embeds := &models.MessageEmbeds{}

	// Regex pour YouTube
	youtubeRegex := regexp.MustCompile(`(?i)(?:youtube\.com\/watch\?v=|youtu\.be\/|youtube\.com\/embed\/)([a-zA-Z0-9_-]{11})`)
	if matches := youtubeRegex.FindStringSubmatch(content); len(matches) > 1 {
		videoID := matches[1]
		embedURL := fmt.Sprintf("https://www.youtube.com/embed/%s", videoID)
		embeds.YouTube = &embedURL
	}

	// Regex pour Spotify
	spotifyRegex := regexp.MustCompile(`(?i)spotify\.com\/(track|album|playlist)\/([a-zA-Z0-9]{22})`)
	if matches := spotifyRegex.FindStringSubmatch(content); len(matches) > 2 {
		itemType := matches[1]
		itemID := matches[2]
		embedURL := fmt.Sprintf("https://open.spotify.com/embed/%s/%s", itemType, itemID)
		embeds.Spotify = &embedURL
	}

	// Retourner nil si aucun embed trouvé
	if embeds.YouTube == nil && embeds.Spotify == nil {
		return nil
	}

	return embeds
}
