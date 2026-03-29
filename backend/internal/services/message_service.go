package services

import (
	"fmt"
	"rythmitbackend/internal/models"
	"rythmitbackend/internal/repositories"
)

// MessageService interface pour la logique métier des messages directs
type MessageService interface {
	// Conversations
	GetOrCreateConversation(user1ID, user2ID uint) (*models.Conversation, error)
	GetUserConversations(userID uint) ([]*models.ConversationWithDetails, error)
	GetConversation(conversationID, userID uint) (*models.Conversation, error)
	DeleteConversation(conversationID, userID uint) error

	// Messages
	SendMessage(senderID, receiverID uint, content string) (*models.DirectMessage, error)
	GetConversationMessages(conversationID, userID uint, limit, offset int) ([]*models.DirectMessage, error)
	MarkConversationAsRead(conversationID, userID uint) error
	GetUnreadCount(userID uint) (int, error)

	// Présence
	UpdateTypingStatus(conversationID, userID uint, isTyping bool) error
	UpdatePresence(conversationID, userID uint) error

	// Vérifications
	CanAccessConversation(conversationID, userID uint) (bool, error)
	AreFriends(user1ID, user2ID uint) (bool, error)
}

// messageService implémentation concrète
type messageService struct {
	messageRepo    repositories.DirectMessageRepository
	friendshipRepo repositories.FriendshipRepository
}

// NewMessageService crée une nouvelle instance du service
func NewMessageService(messageRepo repositories.DirectMessageRepository, friendshipRepo repositories.FriendshipRepository) MessageService {
	return &messageService{
		messageRepo:    messageRepo,
		friendshipRepo: friendshipRepo,
	}
}

// GetOrCreateConversation récupère ou crée une conversation
func (s *messageService) GetOrCreateConversation(user1ID, user2ID uint) (*models.Conversation, error) {
	// Vérifier que les utilisateurs sont amis
	areFriends, err := s.friendshipRepo.AreFriends(user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification amitié: %w", err)
	}

	if !areFriends {
		return nil, fmt.Errorf("les utilisateurs doivent être amis pour démarrer une conversation")
	}

	return s.messageRepo.GetOrCreateConversation(user1ID, user2ID)
}

// GetUserConversations récupère toutes les conversations d'un utilisateur
func (s *messageService) GetUserConversations(userID uint) ([]*models.ConversationWithDetails, error) {
	return s.messageRepo.GetUserConversations(userID)
}

// GetConversation récupère une conversation spécifique
func (s *messageService) GetConversation(conversationID, userID uint) (*models.Conversation, error) {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, fmt.Errorf("accès refusé à cette conversation")
	}

	return s.messageRepo.GetConversationByID(conversationID)
}

// SendMessage envoie un message
func (s *messageService) SendMessage(senderID, receiverID uint, content string) (*models.DirectMessage, error) {
	// Validation du contenu
	if content == "" {
		return nil, fmt.Errorf("le message ne peut pas être vide")
	}

	if len(content) > 2000 {
		return nil, fmt.Errorf("le message est trop long (max 2000 caractères)")
	}

	// Vérifier que les utilisateurs sont amis
	areFriends, err := s.friendshipRepo.AreFriends(senderID, receiverID)
	if err != nil {
		return nil, fmt.Errorf("erreur vérification amitié: %w", err)
	}

	if !areFriends {
		return nil, fmt.Errorf("vous devez être amis pour envoyer des messages")
	}

	// Récupérer ou créer la conversation
	conversation, err := s.messageRepo.GetOrCreateConversation(senderID, receiverID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération conversation: %w", err)
	}

	// Créer le message
	message := &models.DirectMessage{
		ConversationID: conversation.ID,
		SenderID:       senderID,
		ReceiverID:     receiverID,
		Content:        content,
	}

	err = s.messageRepo.CreateMessage(message)
	if err != nil {
		return nil, fmt.Errorf("erreur création message: %w", err)
	}

	return message, nil
}

// GetConversationMessages récupère les messages d'une conversation
func (s *messageService) GetConversationMessages(conversationID, userID uint, limit, offset int) ([]*models.DirectMessage, error) {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return nil, err
	}

	if !canAccess {
		return nil, fmt.Errorf("accès refusé à cette conversation")
	}

	// Limites de pagination
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	return s.messageRepo.GetConversationMessages(conversationID, limit, offset)
}

// MarkConversationAsRead marque tous les messages d'une conversation comme lus
func (s *messageService) MarkConversationAsRead(conversationID, userID uint) error {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}

	if !canAccess {
		return fmt.Errorf("accès refusé à cette conversation")
	}

	return s.messageRepo.MarkConversationAsRead(conversationID, userID)
}

// GetUnreadCount récupère le nombre de messages non lus
func (s *messageService) GetUnreadCount(userID uint) (int, error) {
	return s.messageRepo.GetUnreadCount(userID)
}

// UpdateTypingStatus met à jour le statut "en train d'écrire"
func (s *messageService) UpdateTypingStatus(conversationID, userID uint, isTyping bool) error {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}

	if !canAccess {
		return fmt.Errorf("accès refusé à cette conversation")
	}

	return s.messageRepo.UpdatePresence(conversationID, userID, isTyping)
}

// UpdatePresence met à jour la présence d'un utilisateur
func (s *messageService) UpdatePresence(conversationID, userID uint) error {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}

	if !canAccess {
		return fmt.Errorf("accès refusé à cette conversation")
	}

	return s.messageRepo.UpdatePresence(conversationID, userID, false)
}

// CanAccessConversation vérifie si un utilisateur peut accéder à une conversation
func (s *messageService) CanAccessConversation(conversationID, userID uint) (bool, error) {
	conversation, err := s.messageRepo.GetConversationByID(conversationID)
	if err != nil {
		return false, err
	}

	// L'utilisateur doit être l'un des participants
	if conversation.User1ID != userID && conversation.User2ID != userID {
		return false, nil
	}

	return true, nil
}

// AreFriends vérifie si deux utilisateurs sont amis
func (s *messageService) AreFriends(user1ID, user2ID uint) (bool, error) {
	return s.friendshipRepo.AreFriends(user1ID, user2ID)
}

// DeleteConversation supprime une conversation
func (s *messageService) DeleteConversation(conversationID, userID uint) error {
	// Vérifier l'accès
	canAccess, err := s.CanAccessConversation(conversationID, userID)
	if err != nil {
		return err
	}

	if !canAccess {
		return fmt.Errorf("accès refusé à cette conversation")
	}

	return s.messageRepo.DeleteConversation(conversationID)
}
