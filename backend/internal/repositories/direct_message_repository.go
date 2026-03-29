package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"time"
)

// DirectMessageRepository interface pour les opérations sur les messages directs
type DirectMessageRepository interface {
	// Conversations
	GetOrCreateConversation(user1ID, user2ID uint) (*models.Conversation, error)
	GetConversationByID(conversationID uint) (*models.Conversation, error)
	GetUserConversations(userID uint) ([]*models.ConversationWithDetails, error)
	DeleteConversation(conversationID uint) error

	// Messages
	CreateMessage(message *models.DirectMessage) error
	GetConversationMessages(conversationID uint, limit, offset int) ([]*models.DirectMessage, error)
	MarkMessageAsRead(messageID uint) error
	MarkConversationAsRead(conversationID, userID uint) error
	GetUnreadCount(userID uint) (int, error)
	GetConversationUnreadCount(conversationID, userID uint) (int, error)

	// Présence
	UpdatePresence(conversationID, userID uint, isTyping bool) error
	GetPresence(conversationID, userID uint) (*models.ConversationPresence, error)
	RemovePresence(conversationID, userID uint) error
}

// directMessageRepository implémentation concrète
type directMessageRepository struct {
	*BaseRepository
}

// NewDirectMessageRepository crée une nouvelle instance du repository
func NewDirectMessageRepository(db *sql.DB) DirectMessageRepository {
	return &directMessageRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// GetOrCreateConversation récupère ou crée une conversation entre deux utilisateurs
func (r *directMessageRepository) GetOrCreateConversation(user1ID, user2ID uint) (*models.Conversation, error) {
	// S'assurer que user1ID < user2ID pour la requête
	if user1ID > user2ID {
		user1ID, user2ID = user2ID, user1ID
	}

	// Vérifier si la conversation existe
	query := `
		SELECT id, user1_id, user2_id, last_message_id, last_message_at, created_at, updated_at
		FROM conversations
		WHERE (user1_id = ? AND user2_id = ?) OR (user1_id = ? AND user2_id = ?)
		LIMIT 1
	`

	conversation := &models.Conversation{}
	err := r.DB.QueryRow(query, user1ID, user2ID, user2ID, user1ID).Scan(
		&conversation.ID, &conversation.User1ID, &conversation.User2ID,
		&conversation.LastMessageID, &conversation.LastMessageAt,
		&conversation.CreatedAt, &conversation.UpdatedAt,
	)

	if err == nil {
		return conversation, nil
	}

	if err != sql.ErrNoRows {
		return nil, fmt.Errorf("erreur vérification conversation: %w", err)
	}

	// Créer une nouvelle conversation
	insertQuery := `
		INSERT INTO conversations (user1_id, user2_id, created_at, updated_at)
		VALUES (?, ?, NOW(), NOW())
	`

	result, err := r.DB.Exec(insertQuery, user1ID, user2ID)
	if err != nil {
		return nil, fmt.Errorf("erreur création conversation: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("erreur récupération ID conversation: %w", err)
	}

	conversation.ID = uint(id)
	conversation.User1ID = user1ID
	conversation.User2ID = user2ID
	conversation.CreatedAt = time.Now()
	conversation.UpdatedAt = time.Now()

	return conversation, nil
}

// GetConversationByID récupère une conversation par son ID
func (r *directMessageRepository) GetConversationByID(conversationID uint) (*models.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, last_message_id, last_message_at, created_at, updated_at
		FROM conversations
		WHERE id = ?
	`

	conversation := &models.Conversation{}
	err := r.DB.QueryRow(query, conversationID).Scan(
		&conversation.ID, &conversation.User1ID, &conversation.User2ID,
		&conversation.LastMessageID, &conversation.LastMessageAt,
		&conversation.CreatedAt, &conversation.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversation non trouvée")
		}
		return nil, fmt.Errorf("erreur récupération conversation: %w", err)
	}

	return conversation, nil
}

// GetUserConversations récupère toutes les conversations d'un utilisateur
func (r *directMessageRepository) GetUserConversations(userID uint) ([]*models.ConversationWithDetails, error) {
	query := `
		SELECT 
			c.id, c.user1_id, c.user2_id, c.last_message_at, c.created_at, c.updated_at,
			u.id, u.username, u.profile_pic,
			dm.content as last_message_text,
			COUNT(CASE WHEN dm2.is_read = 0 AND dm2.receiver_id = ? THEN 1 END) as unread_count
		FROM conversations c
		INNER JOIN users u ON (
			CASE 
				WHEN c.user1_id = ? THEN c.user2_id = u.id
				ELSE c.user1_id = u.id
			END
		)
		LEFT JOIN direct_messages dm ON c.last_message_id = dm.id
		LEFT JOIN direct_messages dm2 ON c.id = dm2.conversation_id
		WHERE c.user1_id = ? OR c.user2_id = ?
		GROUP BY c.id, u.id, dm.content
		ORDER BY c.last_message_at DESC, c.created_at DESC
	`

	rows, err := r.DB.Query(query, userID, userID, userID, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération conversations: %w", err)
	}
	defer rows.Close()

	var conversations []*models.ConversationWithDetails
	for rows.Next() {
		conv := &models.ConversationWithDetails{}
		conv.OtherUser = &models.User{}
		
		var lastMessageText sql.NullString
		var profilePic sql.NullString
		
		err := rows.Scan(
			&conv.ID, &conv.User1ID, &conv.User2ID, &conv.LastMessageAt, 
			&conv.CreatedAt, &conv.UpdatedAt,
			&conv.OtherUser.ID, &conv.OtherUser.Username, &profilePic,
			&lastMessageText, &conv.UnreadCount,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan conversation: %w", err)
		}

		if lastMessageText.Valid {
			conv.LastMessageText = lastMessageText.String
		}
		
		if profilePic.Valid {
			conv.OtherUser.ProfilePic = &profilePic.String
		}

		// Vérifier la présence de l'autre utilisateur
		presence, _ := r.GetPresence(conv.ID, conv.OtherUser.ID)
		if presence != nil {
			conv.IsTyping = presence.IsTyping
			// Considérer en ligne si vu dans les 5 dernières minutes
			conv.IsOnline = time.Since(presence.LastSeenAt) < 5*time.Minute
		}

		conversations = append(conversations, conv)
	}

	return conversations, nil
}

// CreateMessage crée un nouveau message
func (r *directMessageRepository) CreateMessage(message *models.DirectMessage) error {
	query := `
		INSERT INTO direct_messages (conversation_id, sender_id, receiver_id, content, created_at, updated_at)
		VALUES (?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.DB.Exec(query, message.ConversationID, message.SenderID, message.ReceiverID, message.Content)
	if err != nil {
		return fmt.Errorf("erreur création message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID message: %w", err)
	}

	message.ID = uint(id)
	message.CreatedAt = time.Now()
	message.UpdatedAt = time.Now()

	// Mettre à jour la conversation
	updateQuery := `
		UPDATE conversations
		SET last_message_id = ?, last_message_at = NOW(), updated_at = NOW()
		WHERE id = ?
	`

	_, err = r.DB.Exec(updateQuery, message.ID, message.ConversationID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour conversation: %w", err)
	}

	return nil
}

// GetConversationMessages récupère les messages d'une conversation
func (r *directMessageRepository) GetConversationMessages(conversationID uint, limit, offset int) ([]*models.DirectMessage, error) {
	query := `
		SELECT id, conversation_id, sender_id, receiver_id, content, is_read, read_at, created_at, updated_at
		FROM direct_messages
		WHERE conversation_id = ?
		ORDER BY created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.DirectMessage
	for rows.Next() {
		message := &models.DirectMessage{}
		err := rows.Scan(
			&message.ID, &message.ConversationID, &message.SenderID, &message.ReceiverID,
			&message.Content, &message.IsRead, &message.ReadAt,
			&message.CreatedAt, &message.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan message: %w", err)
		}
		messages = append(messages, message)
	}

	// Inverser l'ordre pour avoir les plus anciens en premier
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// MarkMessageAsRead marque un message comme lu
func (r *directMessageRepository) MarkMessageAsRead(messageID uint) error {
	query := `
		UPDATE direct_messages
		SET is_read = TRUE, read_at = NOW()
		WHERE id = ? AND is_read = FALSE
	`

	_, err := r.DB.Exec(query, messageID)
	if err != nil {
		return fmt.Errorf("erreur marquage message lu: %w", err)
	}

	return nil
}

// MarkConversationAsRead marque tous les messages d'une conversation comme lus
func (r *directMessageRepository) MarkConversationAsRead(conversationID, userID uint) error {
	query := `
		UPDATE direct_messages
		SET is_read = TRUE, read_at = NOW()
		WHERE conversation_id = ? AND receiver_id = ? AND is_read = FALSE
	`

	_, err := r.DB.Exec(query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("erreur marquage conversation lue: %w", err)
	}

	return nil
}

// GetUnreadCount récupère le nombre total de messages non lus
func (r *directMessageRepository) GetUnreadCount(userID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM direct_messages
		WHERE receiver_id = ? AND is_read = FALSE
	`

	var count int
	err := r.DB.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage messages non lus: %w", err)
	}

	return count, nil
}

// GetConversationUnreadCount récupère le nombre de messages non lus dans une conversation
func (r *directMessageRepository) GetConversationUnreadCount(conversationID, userID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM direct_messages
		WHERE conversation_id = ? AND receiver_id = ? AND is_read = FALSE
	`

	var count int
	err := r.DB.QueryRow(query, conversationID, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage messages non lus: %w", err)
	}

	return count, nil
}

// UpdatePresence met à jour la présence d'un utilisateur dans une conversation
func (r *directMessageRepository) UpdatePresence(conversationID, userID uint, isTyping bool) error {
	query := `
		INSERT INTO conversation_presence (conversation_id, user_id, is_typing, last_seen_at, created_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE is_typing = ?, last_seen_at = NOW()
	`

	_, err := r.DB.Exec(query, conversationID, userID, isTyping, isTyping)
	if err != nil {
		return fmt.Errorf("erreur mise à jour présence: %w", err)
	}

	return nil
}

// GetPresence récupère la présence d'un utilisateur dans une conversation
func (r *directMessageRepository) GetPresence(conversationID, userID uint) (*models.ConversationPresence, error) {
	query := `
		SELECT id, conversation_id, user_id, is_typing, last_seen_at, created_at
		FROM conversation_presence
		WHERE conversation_id = ? AND user_id = ?
	`

	presence := &models.ConversationPresence{}
	err := r.DB.QueryRow(query, conversationID, userID).Scan(
		&presence.ID, &presence.ConversationID, &presence.UserID,
		&presence.IsTyping, &presence.LastSeenAt, &presence.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("erreur récupération présence: %w", err)
	}

	return presence, nil
}

// RemovePresence supprime la présence d'un utilisateur
func (r *directMessageRepository) RemovePresence(conversationID, userID uint) error {
	query := `DELETE FROM conversation_presence WHERE conversation_id = ? AND user_id = ?`

	_, err := r.DB.Exec(query, conversationID, userID)
	if err != nil {
		return fmt.Errorf("erreur suppression présence: %w", err)
	}

	return nil
}

// DeleteConversation supprime une conversation
func (r *directMessageRepository) DeleteConversation(conversationID uint) error {
	// Les messages et la présence seront supprimés automatiquement par CASCADE
	query := `DELETE FROM conversations WHERE id = ?`

	_, err := r.DB.Exec(query, conversationID)
	if err != nil {
		return fmt.Errorf("erreur suppression conversation: %w", err)
	}

	return nil
}

