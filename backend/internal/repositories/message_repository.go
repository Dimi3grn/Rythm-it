package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"strings"
)

// MessageRepository interface pour les opérations CRUD sur les messages
type MessageRepository interface {
	Create(message *models.Message) error
	FindByID(id uint) (*models.Message, error)
	FindByThreadID(threadID uint, params models.PaginationParams, sortBy string) ([]*models.Message, int64, error)
	FindByUserID(userID uint, params models.PaginationParams) ([]*models.Message, int64, error)
	Update(message *models.Message) error
	Delete(id uint) error
	CountByThreadID(threadID uint) (int64, error)
	GetPopularityScore(messageID uint) (int, error)
	GetUserVote(userID uint, messageID uint) (string, error)
	SetUserVote(userID uint, messageID uint, vote string) error
	GetMessageVoteCounts(messageID uint) (fireCount int, skipCount int, err error)
	GetMessagesWithVotes(threadID uint, userID *uint, params models.PaginationParams, sortBy string) ([]*models.Message, int64, error)
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

// Create crée un nouveau message avec meilleure gestion d'erreur
func (r *messageRepository) Create(message *models.Message) error {
	if message == nil {
		return fmt.Errorf("message ne peut pas être nil")
	}

	if message.Content == "" {
		return fmt.Errorf("le contenu du message ne peut pas être vide")
	}

	if message.ThreadID == 0 {
		return fmt.Errorf("thread_id est requis")
	}

	if message.UserID == 0 {
		return fmt.Errorf("user_id est requis")
	}

	query := `
		INSERT INTO messages (content, thread_id, user_id, created_at, updated_at) 
		VALUES (?, ?, ?, NOW(), NOW())
	`

	result, err := r.DB.Exec(query, message.Content, message.ThreadID, message.UserID)
	if err != nil {
		// Vérifier les erreurs spécifiques de la base de données
		if strings.Contains(err.Error(), "foreign key constraint fails") {
			if strings.Contains(err.Error(), "thread_id") {
				return fmt.Errorf("thread non trouvé: %w", err)
			}
			if strings.Contains(err.Error(), "user_id") {
				return fmt.Errorf("utilisateur non trouvé: %w", err)
			}
		}
		return fmt.Errorf("erreur création message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID message: %w", err)
	}

	message.ID = uint(id)
	return nil
}

// FindByID trouve un message par son ID avec meilleure gestion d'erreur
func (r *messageRepository) FindByID(id uint) (*models.Message, error) {
	if id == 0 {
		return nil, fmt.Errorf("id invalide")
	}

	query := `
		SELECT m.id, m.content, m.thread_id, m.user_id, m.created_at, m.updated_at,
			   u.id, u.username, u.email, u.profile_pic
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.id = ?
	`

	message := &models.Message{}
	var author models.User

	err := r.DB.QueryRow(query, id).Scan(
		&message.ID, &message.Content, &message.ThreadID, &message.UserID,
		&message.CreatedAt, &message.UpdatedAt,
		&author.ID, &author.Username, &author.Email, &author.ProfilePic,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("message ID %d non trouvé", id)
		}
		return nil, fmt.Errorf("erreur récupération message: %w", err)
	}

	message.Author = &author
	return message, nil
}

// FindByThreadID récupère les messages d'un thread avec pagination
func (r *messageRepository) FindByThreadID(threadID uint, params models.PaginationParams, sortBy string) ([]*models.Message, int64, error) {
	// Validation des paramètres
	models.ValidatePagination(&params)

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM messages WHERE thread_id = ?"
	var total int64
	err := r.DB.QueryRow(countQuery, threadID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage messages: %w", err)
	}

	// Construire la clause ORDER BY
	orderClause := r.buildOrderClause(sortBy)

	// Récupérer les messages avec l'auteur
	offset := (params.Page - 1) * params.PerPage
	query := fmt.Sprintf(`
		SELECT m.id, m.content, m.thread_id, m.user_id, m.created_at, m.updated_at,
			   u.id, u.username, u.email, u.profile_pic
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.thread_id = ?
		%s
		LIMIT ? OFFSET ?
	`, orderClause)

	rows, err := r.DB.Query(query, threadID, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération messages: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{Author: &models.User{}}
		err := rows.Scan(
			&message.ID, &message.Content, &message.ThreadID, &message.UserID, &message.CreatedAt, &message.UpdatedAt,
			&message.Author.ID, &message.Author.Username, &message.Author.Email, &message.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan message: %w", err)
		}

		// Calculer le score de popularité
		score, scoreErr := r.GetPopularityScore(message.ID)
		if scoreErr == nil {
			message.PopularityScore = score
		}

		messages = append(messages, message)
	}

	return messages, total, nil
}

// GetMessagesWithVotes récupère les messages avec les votes de l'utilisateur
func (r *messageRepository) GetMessagesWithVotes(threadID uint, userID *uint, params models.PaginationParams, sortBy string) ([]*models.Message, int64, error) {
	// D'abord récupérer les messages normalement
	messages, total, err := r.FindByThreadID(threadID, params, sortBy)
	if err != nil {
		return nil, 0, err
	}

	// Si un utilisateur est connecté, récupérer ses votes
	if userID != nil {
		for _, message := range messages {
			vote, voteErr := r.GetUserVote(*userID, message.ID)
			if voteErr == nil && vote != models.VoteNeutral {
				message.UserVote = &vote
			}
		}
	}

	return messages, total, nil
}

// FindByUserID récupère les messages d'un utilisateur
func (r *messageRepository) FindByUserID(userID uint, params models.PaginationParams) ([]*models.Message, int64, error) {
	models.ValidatePagination(&params)

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM messages WHERE user_id = ?"
	var total int64
	err := r.DB.QueryRow(countQuery, userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage messages utilisateur: %w", err)
	}

	// Récupérer les messages
	offset := (params.Page - 1) * params.PerPage
	query := `
		SELECT m.id, m.content, m.thread_id, m.user_id, m.created_at, m.updated_at,
			   u.id, u.username, u.email, u.profile_pic
		FROM messages m
		JOIN users u ON m.user_id = u.id
		WHERE m.user_id = ?
		ORDER BY m.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, userID, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération messages utilisateur: %w", err)
	}
	defer rows.Close()

	var messages []*models.Message
	for rows.Next() {
		message := &models.Message{Author: &models.User{}}
		err := rows.Scan(
			&message.ID, &message.Content, &message.ThreadID, &message.UserID, &message.CreatedAt, &message.UpdatedAt,
			&message.Author.ID, &message.Author.Username, &message.Author.Email, &message.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan message utilisateur: %w", err)
		}

		messages = append(messages, message)
	}

	return messages, total, nil
}

// Update met à jour un message avec meilleure gestion d'erreur
func (r *messageRepository) Update(message *models.Message) error {
	if message == nil {
		return fmt.Errorf("message ne peut pas être nil")
	}

	if message.ID == 0 {
		return fmt.Errorf("id invalide")
	}

	if message.Content == "" {
		return fmt.Errorf("le contenu du message ne peut pas être vide")
	}

	// Vérifier que le message existe
	exists, err := r.messageExists(message.ID)
	if err != nil {
		return fmt.Errorf("erreur vérification existence message: %w", err)
	}
	if !exists {
		return fmt.Errorf("message ID %d non trouvé", message.ID)
	}

	query := `
		UPDATE messages 
		SET content = ?, updated_at = NOW()
		WHERE id = ?
	`

	result, err := r.DB.Exec(query, message.Content, message.ID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour message: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("message ID %d non trouvé pour mise à jour", message.ID)
	}

	return nil
}

// messageExists vérifie si un message existe
func (r *messageRepository) messageExists(id uint) (bool, error) {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM messages WHERE id = ?)"
	err := r.DB.QueryRow(query, id).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur vérification existence message: %w", err)
	}
	return exists, nil
}

// Delete supprime un message
func (r *messageRepository) Delete(id uint) error {
	// Les votes seront supprimés automatiquement par CASCADE
	query := "DELETE FROM messages WHERE id = ?"

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression message: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("message ID %d non trouvé pour suppression", id)
	}

	return nil
}

// CountByThreadID compte les messages dans un thread
func (r *messageRepository) CountByThreadID(threadID uint) (int64, error) {
	query := "SELECT COUNT(*) FROM messages WHERE thread_id = ?"

	var count int64
	err := r.DB.QueryRow(query, threadID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage messages thread: %w", err)
	}

	return count, nil
}

// GetPopularityScore calcule le score de popularité (Fire - Skip)
func (r *messageRepository) GetPopularityScore(messageID uint) (int, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN state = 'fire' THEN 1 ELSE 0 END), 0) as fire_count,
			COALESCE(SUM(CASE WHEN state = 'skip' THEN 1 ELSE 0 END), 0) as skip_count
		FROM message_votes 
		WHERE message_id = ?
	`

	var fireCount, skipCount int
	err := r.DB.QueryRow(query, messageID).Scan(&fireCount, &skipCount)
	if err != nil {
		return 0, fmt.Errorf("erreur calcul score popularité: %w", err)
	}

	return fireCount - skipCount, nil
}

// GetUserVote récupère le vote d'un utilisateur pour un message
func (r *messageRepository) GetUserVote(userID uint, messageID uint) (string, error) {
	query := "SELECT state FROM message_votes WHERE user_id = ? AND message_id = ?"

	var vote string
	err := r.DB.QueryRow(query, userID, messageID).Scan(&vote)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.VoteNeutral, nil
		}
		return "", fmt.Errorf("erreur récupération vote: %w", err)
	}

	return vote, nil
}

// SetUserVote définit ou met à jour le vote d'un utilisateur
func (r *messageRepository) SetUserVote(userID uint, messageID uint, vote string) error {
	// Valider le vote
	if vote != models.VoteFire && vote != models.VoteSkip && vote != models.VoteNeutral {
		return fmt.Errorf("vote invalide: %s", vote)
	}

	// Utiliser UPSERT (INSERT ... ON DUPLICATE KEY UPDATE)
	query := `
		INSERT INTO message_votes (user_id, message_id, state, created_at, updated_at) 
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE 
		state = VALUES(state), 
		updated_at = NOW()
	`

	_, err := r.DB.Exec(query, userID, messageID, vote)
	if err != nil {
		return fmt.Errorf("erreur enregistrement vote: %w", err)
	}

	return nil
}

// GetMessageVoteCounts récupère les compteurs de votes pour un message
func (r *messageRepository) GetMessageVoteCounts(messageID uint) (fireCount int, skipCount int, err error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN state = 'fire' THEN 1 ELSE 0 END), 0) as fire_count,
			COALESCE(SUM(CASE WHEN state = 'skip' THEN 1 ELSE 0 END), 0) as skip_count
		FROM message_votes 
		WHERE message_id = ?
	`

	err = r.DB.QueryRow(query, messageID).Scan(&fireCount, &skipCount)
	if err != nil {
		return 0, 0, fmt.Errorf("erreur comptage votes: %w", err)
	}

	return fireCount, skipCount, nil
}

// buildOrderClause construit la clause ORDER BY selon le tri demandé
func (r *messageRepository) buildOrderClause(sortBy string) string {
	switch strings.ToLower(sortBy) {
	case "popularity", "popularité":
		// Tri par popularité (Fire - Skip) puis par date
		return `ORDER BY (
			SELECT COALESCE(SUM(CASE WHEN mv.state = 'fire' THEN 1 ELSE 0 END), 0) - 
			       COALESCE(SUM(CASE WHEN mv.state = 'skip' THEN 1 ELSE 0 END), 0)
			FROM message_votes mv 
			WHERE mv.message_id = m.id
		) DESC, m.created_at DESC`
	case "date", "chronologique", "":
		// Tri chronologique (par défaut)
		return "ORDER BY m.created_at DESC"
	default:
		// Fallback sur chronologique
		return "ORDER BY m.created_at DESC"
	}
}
