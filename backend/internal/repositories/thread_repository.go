package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
)

// ThreadRepository interface pour les opérations CRUD sur les threads
type ThreadRepository interface {
	Create(thread *models.Thread) error
	FindByID(id uint) (*models.Thread, error)
	FindAll(params models.PaginationParams) ([]*models.Thread, int64, error)
	FindByUserID(userID uint) ([]*models.Thread, error)
	FindPublicThreads(params models.PaginationParams) ([]*models.Thread, int64, error)
	Update(thread *models.Thread) error
	Delete(id uint) error
	UpdateState(id uint, state string) error
	AttachTags(threadID uint, tagIDs []uint) error
	DetachTags(threadID uint) error
	GetThreadTags(threadID uint) ([]*models.Tag, error)
	FindByTag(tagID uint, params models.PaginationParams) ([]*models.Thread, int64, error)
	Search(query string, params models.PaginationParams) ([]*models.Thread, int64, error)
	Transaction(fn func(*sql.Tx) error) error
}

// threadRepository implémentation concrète
type threadRepository struct {
	*BaseRepository
}

// NewThreadRepository crée une nouvelle instance du repository
func NewThreadRepository(db *sql.DB) ThreadRepository {
	return &threadRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create crée un nouveau thread
func (r *threadRepository) Create(thread *models.Thread) error {
	query := `
		INSERT INTO threads (title, desc_, state, visibility, user_id, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, NOW(), NOW())
	`

	result, err := r.DB.Exec(query, thread.Title, thread.Description, thread.State, thread.Visibility, thread.UserID)
	if err != nil {
		return fmt.Errorf("erreur création thread: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID thread: %w", err)
	}

	thread.ID = uint(id)
	return nil
}

// FindByID trouve un thread par son ID avec l'auteur
func (r *threadRepository) FindByID(id uint) (*models.Thread, error) {
	query := `
		SELECT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN users u ON t.user_id = u.id
		WHERE t.id = ?
	`

	thread := &models.Thread{Author: &models.User{}}
	err := r.DB.QueryRow(query, id).Scan(
		&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
		&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("thread ID %d non trouvé", id)
		}
		return nil, fmt.Errorf("erreur récupération thread: %w", err)
	}

	// Charger les tags du thread
	tags, err := r.GetThreadTags(thread.ID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags: %w", err)
	}
	thread.Tags = tags

	return thread, nil
}

// FindPublicThreads récupère les threads publics avec pagination
func (r *threadRepository) FindPublicThreads(params models.PaginationParams) ([]*models.Thread, int64, error) {
	// Validation des paramètres
	models.ValidatePagination(&params)

	// Compter le total
	countQuery := "SELECT COUNT(*) FROM threads WHERE visibility = 'public' AND state != 'archivé'"
	var total int64
	err := r.DB.QueryRow(countQuery).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage threads: %w", err)
	}

	// Récupérer les threads avec l'auteur
	offset := (params.Page - 1) * params.PerPage
	query := `
		SELECT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN users u ON t.user_id = u.id
		WHERE t.visibility = 'public' AND t.state != 'archivé'
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération threads: %w", err)
	}
	defer rows.Close()

	var threads []*models.Thread
	for rows.Next() {
		thread := &models.Thread{Author: &models.User{}}
		err := rows.Scan(
			&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan thread: %w", err)
		}

		// Charger les tags pour chaque thread
		tags, err := r.GetThreadTags(thread.ID)
		if err != nil {
			// Log l'erreur mais continue
			tags = []*models.Tag{}
		}
		thread.Tags = tags

		threads = append(threads, thread)
	}

	return threads, total, nil
}

// FindAll récupère tous les threads (admin)
func (r *threadRepository) FindAll(params models.PaginationParams) ([]*models.Thread, int64, error) {
	models.ValidatePagination(&params)

	// Compter le total
	var total int64
	err := r.DB.QueryRow("SELECT COUNT(*) FROM threads").Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage threads: %w", err)
	}

	// Récupérer les threads
	offset := (params.Page - 1) * params.PerPage
	query := `
		SELECT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN users u ON t.user_id = u.id
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération threads: %w", err)
	}
	defer rows.Close()

	var threads []*models.Thread
	for rows.Next() {
		thread := &models.Thread{Author: &models.User{}}
		err := rows.Scan(
			&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan thread: %w", err)
		}

		tags, _ := r.GetThreadTags(thread.ID)
		thread.Tags = tags
		threads = append(threads, thread)
	}

	return threads, total, nil
}

// FindByUserID trouve les threads d'un utilisateur
func (r *threadRepository) FindByUserID(userID uint) ([]*models.Thread, error) {
	query := `
		SELECT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN users u ON t.user_id = u.id
		WHERE t.user_id = ?
		ORDER BY t.created_at DESC
	`

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération threads utilisateur: %w", err)
	}
	defer rows.Close()

	var threads []*models.Thread
	for rows.Next() {
		thread := &models.Thread{Author: &models.User{}}
		err := rows.Scan(
			&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan thread: %w", err)
		}

		tags, _ := r.GetThreadTags(thread.ID)
		thread.Tags = tags
		threads = append(threads, thread)
	}

	return threads, nil
}

// Update met à jour un thread
func (r *threadRepository) Update(thread *models.Thread) error {
	query := `
		UPDATE threads 
		SET title = ?, desc_ = ?, state = ?, visibility = ?, updated_at = NOW()
		WHERE id = ?
	`

	_, err := r.DB.Exec(query, thread.Title, thread.Description, thread.State, thread.Visibility, thread.ID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour thread: %w", err)
	}

	return nil
}

// Delete supprime un thread
func (r *threadRepository) Delete(id uint) error {
	// Les suppressions en cascade sont gérées par la DB (messages, tags, etc.)
	query := "DELETE FROM threads WHERE id = ?"

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression thread: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("thread ID %d non trouvé pour suppression", id)
	}

	return nil
}

// UpdateState change l'état d'un thread
func (r *threadRepository) UpdateState(id uint, state string) error {
	query := "UPDATE threads SET state = ?, updated_at = NOW() WHERE id = ?"

	result, err := r.DB.Exec(query, state, id)
	if err != nil {
		return fmt.Errorf("erreur changement état thread: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification changement état: %w", err)
	}

	if affected == 0 {
		return fmt.Errorf("thread ID %d non trouvé pour changement état", id)
	}

	return nil
}

// AttachTags attache des tags à un thread
func (r *threadRepository) AttachTags(threadID uint, tagIDs []uint) error {
	if len(tagIDs) == 0 {
		return nil
	}

	// D'abord supprimer les anciens tags
	err := r.DetachTags(threadID)
	if err != nil {
		return fmt.Errorf("erreur suppression anciens tags: %w", err)
	}

	// Insérer les nouveaux tags
	query := "INSERT INTO thread_tags (thread_id, tag_id) VALUES (?, ?)"

	for _, tagID := range tagIDs {
		_, err := r.DB.Exec(query, threadID, tagID)
		if err != nil {
			return fmt.Errorf("erreur attachement tag %d: %w", tagID, err)
		}
	}

	return nil
}

// DetachTags supprime tous les tags d'un thread
func (r *threadRepository) DetachTags(threadID uint) error {
	query := "DELETE FROM thread_tags WHERE thread_id = ?"

	_, err := r.DB.Exec(query, threadID)
	if err != nil {
		return fmt.Errorf("erreur suppression tags thread: %w", err)
	}

	return nil
}

// GetThreadTags récupère les tags d'un thread
func (r *threadRepository) GetThreadTags(threadID uint) ([]*models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.type
		FROM tags t
		JOIN thread_tags tt ON t.id = tt.tag_id
		WHERE tt.thread_id = ?
		ORDER BY t.name
	`

	rows, err := r.DB.Query(query, threadID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération tags thread: %w", err)
	}
	defer rows.Close()

	var tags []*models.Tag
	for rows.Next() {
		tag := &models.Tag{}
		err := rows.Scan(&tag.ID, &tag.Name, &tag.Type)
		if err != nil {
			return nil, fmt.Errorf("erreur scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

// FindByTag trouve les threads par tag
func (r *threadRepository) FindByTag(tagID uint, params models.PaginationParams) ([]*models.Thread, int64, error) {
	models.ValidatePagination(&params)

	// Compter le total
	countQuery := `
		SELECT COUNT(DISTINCT t.id)
		FROM threads t
		JOIN thread_tags tt ON t.id = tt.thread_id
		WHERE tt.tag_id = ? AND t.visibility = 'public' AND t.state != 'archivé'
	`
	var total int64
	err := r.DB.QueryRow(countQuery, tagID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage threads par tag: %w", err)
	}

	// Récupérer les threads
	offset := (params.Page - 1) * params.PerPage
	query := `
		SELECT DISTINCT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN thread_tags tt ON t.id = tt.thread_id
		JOIN users u ON t.user_id = u.id
		WHERE tt.tag_id = ? AND t.visibility = 'public' AND t.state != 'archivé'
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(query, tagID, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur récupération threads par tag: %w", err)
	}
	defer rows.Close()

	var threads []*models.Thread
	for rows.Next() {
		thread := &models.Thread{Author: &models.User{}}
		err := rows.Scan(
			&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan thread: %w", err)
		}

		tags, _ := r.GetThreadTags(thread.ID)
		thread.Tags = tags
		threads = append(threads, thread)
	}

	return threads, total, nil
}

// Search recherche dans les threads par titre
func (r *threadRepository) Search(query string, params models.PaginationParams) ([]*models.Thread, int64, error) {
	models.ValidatePagination(&params)

	searchTerm := "%" + query + "%"

	// Compter le total
	countQuery := `
		SELECT COUNT(*)
		FROM threads t
		WHERE t.title LIKE ? AND t.visibility = 'public' AND t.state != 'archivé'
	`
	var total int64
	err := r.DB.QueryRow(countQuery, searchTerm).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur comptage recherche threads: %w", err)
	}

	// Récupérer les threads
	offset := (params.Page - 1) * params.PerPage
	searchQuery := `
		SELECT t.id, t.title, t.desc_, t.state, t.visibility, t.user_id, t.created_at, t.updated_at,
		       u.id, u.username, u.email, u.profile_pic
		FROM threads t
		JOIN users u ON t.user_id = u.id
		WHERE t.title LIKE ? AND t.visibility = 'public' AND t.state != 'archivé'
		ORDER BY t.created_at DESC
		LIMIT ? OFFSET ?
	`

	rows, err := r.DB.Query(searchQuery, searchTerm, params.PerPage, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("erreur recherche threads: %w", err)
	}
	defer rows.Close()

	var threads []*models.Thread
	for rows.Next() {
		thread := &models.Thread{Author: &models.User{}}
		err := rows.Scan(
			&thread.ID, &thread.Title, &thread.Description, &thread.State, &thread.Visibility, &thread.UserID, &thread.CreatedAt, &thread.UpdatedAt,
			&thread.Author.ID, &thread.Author.Username, &thread.Author.Email, &thread.Author.ProfilePic,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("erreur scan thread recherche: %w", err)
		}

		tags, _ := r.GetThreadTags(thread.ID)
		thread.Tags = tags
		threads = append(threads, thread)
	}

	return threads, total, nil
}

// Transaction exécute une transaction
func (r *threadRepository) Transaction(fn func(*sql.Tx) error) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur début de transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if err := tx.Rollback(); err != nil {
			return fmt.Errorf("erreur rollback transaction: %w", err)
		}
		return err
	}

	return tx.Commit()
}
