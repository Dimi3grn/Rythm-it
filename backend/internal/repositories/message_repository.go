package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
)

// MessageRepository interface pour les opérations sur les messages dans les threads (commentaires)
type MessageRepository interface {
	// CRUD de base
	Create(message *models.Message) error
	FindByID(id uint) (*models.Message, error)
	Update(message *models.Message) error
	Delete(id uint) error

	// Récupération
	FindByThreadID(threadID uint, params models.PaginationParams, orderBy string) ([]*models.Message, int, error)
	FindByUserID(userID uint, params models.PaginationParams) ([]*models.Message, int, error)
	GetMessagesWithVotes(threadID uint, userID *uint, params models.PaginationParams, orderBy string) ([]*models.Message, int, error)

	// Comptage
	CountByThreadID(threadID uint) (int, error)

	// Votes
	SetUserVote(messageID, userID uint, voteType string) error
	GetUserVote(messageID, userID uint) (*string, error)
	GetMessageVoteCounts(messageID uint) (upvotes int, downvotes int, error error)
	GetPopularityScore(messageID uint) (int, error)
}

// messageRepository implémentation concrète
type messageRepository struct {
	*BaseRepository
}

// NewMessageRepository crée une nouvelle instance du repository
func NewMessageRepository(db *sql.DB) MessageRepository {
	return &messageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create crée un nouveau message
func (r *messageRepository) Create(message *models.Message) error {
	return fmt.Errorf("MessageRepository.Create not implemented yet - TODO")
}

// FindByID récupère un message par son ID
func (r *messageRepository) FindByID(id uint) (*models.Message, error) {
	return nil, fmt.Errorf("MessageRepository.FindByID not implemented yet - TODO")
}

// Update met à jour un message
func (r *messageRepository) Update(message *models.Message) error {
	return fmt.Errorf("MessageRepository.Update not implemented yet - TODO")
}

// Delete supprime un message
func (r *messageRepository) Delete(id uint) error {
	return fmt.Errorf("MessageRepository.Delete not implemented yet - TODO")
}

// FindByThreadID récupère les messages d'un thread
func (r *messageRepository) FindByThreadID(threadID uint, params models.PaginationParams, orderBy string) ([]*models.Message, int, error) {
	return nil, 0, fmt.Errorf("MessageRepository.FindByThreadID not implemented yet - TODO")
}

// FindByUserID récupère les messages d'un utilisateur
func (r *messageRepository) FindByUserID(userID uint, params models.PaginationParams) ([]*models.Message, int, error) {
	return nil, 0, fmt.Errorf("MessageRepository.FindByUserID not implemented yet - TODO")
}

// GetMessagesWithVotes récupère les messages avec les votes
func (r *messageRepository) GetMessagesWithVotes(threadID uint, userID *uint, params models.PaginationParams, orderBy string) ([]*models.Message, int, error) {
	return nil, 0, fmt.Errorf("MessageRepository.GetMessagesWithVotes not implemented yet - TODO")
}

// CountByThreadID compte les messages dans un thread
func (r *messageRepository) CountByThreadID(threadID uint) (int, error) {
	return 0, fmt.Errorf("MessageRepository.CountByThreadID not implemented yet - TODO")
}

// SetUserVote définit le vote d'un utilisateur sur un message
func (r *messageRepository) SetUserVote(messageID, userID uint, voteType string) error {
	return fmt.Errorf("MessageRepository.SetUserVote not implemented yet - TODO")
}

// GetUserVote récupère le vote d'un utilisateur sur un message
func (r *messageRepository) GetUserVote(messageID, userID uint) (*string, error) {
	return nil, fmt.Errorf("MessageRepository.GetUserVote not implemented yet - TODO")
}

// GetMessageVoteCounts récupère le nombre de votes d'un message
func (r *messageRepository) GetMessageVoteCounts(messageID uint) (upvotes int, downvotes int, error error) {
	return 0, 0, fmt.Errorf("MessageRepository.GetMessageVoteCounts not implemented yet - TODO")
}

// GetPopularityScore calcule le score de popularité d'un message
func (r *messageRepository) GetPopularityScore(messageID uint) (int, error) {
	return 0, fmt.Errorf("MessageRepository.GetPopularityScore not implemented yet - TODO")
}
