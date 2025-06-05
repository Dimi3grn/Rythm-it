package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"time"
)

// UserRepository interface définissant les opérations sur les utilisateurs
type UserRepository interface {
	// CRUD operations
	Create(user *models.User) error
	FindByID(id uint) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	FindByUsername(username string) (*models.User, error)
	Update(user *models.User) error
	Delete(id uint) error

	// Existence checks
	ExistsByEmail(email string) (bool, error)
	ExistsByUsername(username string) (bool, error)

	// Additional methods
	UpdateLastConnection(userID uint) error
	IncrementMessageCount(userID uint) error
	IncrementThreadCount(userID uint) error
}

// userRepository implémentation concrète
type userRepository struct {
	*BaseRepository
}

// NewUserRepository crée une nouvelle instance du repository utilisateur
func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create insère un nouvel utilisateur en base
func (r *userRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, email, password, is_admin, profile_pic, biography, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	result, err := r.DB.Exec(query,
		user.Username,
		user.Email,
		user.Password,
		user.IsAdmin,
		user.ProfilePic,
		user.Biography,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur création utilisateur: %w", err)
	}

	// Récupérer l'ID généré
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID utilisateur: %w", err)
	}

	user.ID = uint(id)
	return nil
}

// FindByID trouve un utilisateur par son ID
func (r *userRepository) FindByID(id uint) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, is_admin, profile_pic, biography, 
		       last_connection, message_count, thread_count, created_at, updated_at
		FROM users 
		WHERE id = ?
	`

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
		&user.ProfilePic,
		&user.Biography,
		&user.LastConnection,
		&user.MessageCount,
		&user.ThreadCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur avec ID %d non trouvé", id)
		}
		return nil, fmt.Errorf("erreur recherche utilisateur par ID: %w", err)
	}

	return user, nil
}

// FindByEmail trouve un utilisateur par son email
func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, is_admin, profile_pic, biography,
		       last_connection, message_count, thread_count, created_at, updated_at
		FROM users 
		WHERE email = ?
	`

	err := r.DB.QueryRow(query, email).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
		&user.ProfilePic,
		&user.Biography,
		&user.LastConnection,
		&user.MessageCount,
		&user.ThreadCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur avec email %s non trouvé", email)
		}
		return nil, fmt.Errorf("erreur recherche utilisateur par email: %w", err)
	}

	return user, nil
}

// FindByUsername trouve un utilisateur par son nom d'utilisateur
func (r *userRepository) FindByUsername(username string) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, email, password, is_admin, profile_pic, biography,
		       last_connection, message_count, thread_count, created_at, updated_at
		FROM users 
		WHERE username = ?
	`

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.IsAdmin,
		&user.ProfilePic,
		&user.Biography,
		&user.LastConnection,
		&user.MessageCount,
		&user.ThreadCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("utilisateur avec username %s non trouvé", username)
		}
		return nil, fmt.Errorf("erreur recherche utilisateur par username: %w", err)
	}

	return user, nil
}

// Update met à jour un utilisateur existant
func (r *userRepository) Update(user *models.User) error {
	query := `
		UPDATE users 
		SET username = ?, email = ?, password = ?, is_admin = ?, 
		    profile_pic = ?, biography = ?, updated_at = ?
		WHERE id = ?
	`

	user.UpdatedAt = time.Now()

	result, err := r.DB.Exec(query,
		user.Username,
		user.Email,
		user.Password,
		user.IsAdmin,
		user.ProfilePic,
		user.Biography,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return fmt.Errorf("erreur mise à jour utilisateur: %w", err)
	}

	// Vérifier qu'une ligne a été affectée
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucun utilisateur trouvé avec ID %d", user.ID)
	}

	return nil
}

// Delete supprime un utilisateur (soft delete non implémenté pour l'instant)
func (r *userRepository) Delete(id uint) error {
	query := `DELETE FROM users WHERE id = ?`

	result, err := r.DB.Exec(query, id)
	if err != nil {
		return fmt.Errorf("erreur suppression utilisateur: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucun utilisateur trouvé avec ID %d", id)
	}

	return nil
}

// ExistsByEmail vérifie si un email existe déjà
func (r *userRepository) ExistsByEmail(email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = ?)`

	var exists bool
	err := r.DB.QueryRow(query, email).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur vérification existence email: %w", err)
	}

	return exists, nil
}

// ExistsByUsername vérifie si un nom d'utilisateur existe déjà
func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE username = ?)`

	var exists bool
	err := r.DB.QueryRow(query, username).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur vérification existence username: %w", err)
	}

	return exists, nil
}

// UpdateLastConnection met à jour la dernière connexion d'un utilisateur
func (r *userRepository) UpdateLastConnection(userID uint) error {
	query := `UPDATE users SET last_connection = ?, updated_at = ? WHERE id = ?`

	now := time.Now()
	result, err := r.DB.Exec(query, now, now, userID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour dernière connexion: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour connexion: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé pour mise à jour connexion", userID)
	}

	return nil
}

// IncrementMessageCount incrémente le compteur de messages d'un utilisateur
func (r *userRepository) IncrementMessageCount(userID uint) error {
	query := `UPDATE users SET message_count = message_count + 1, updated_at = ? WHERE id = ?`

	result, err := r.DB.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("erreur incrémentation compteur messages: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification incrémentation messages: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé pour incrémentation messages", userID)
	}

	return nil
}

// IncrementThreadCount incrémente le compteur de threads d'un utilisateur
func (r *userRepository) IncrementThreadCount(userID uint) error {
	query := `UPDATE users SET thread_count = thread_count + 1, updated_at = ? WHERE id = ?`

	result, err := r.DB.Exec(query, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("erreur incrémentation compteur threads: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification incrémentation threads: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("utilisateur ID %d non trouvé pour incrémentation threads", userID)
	}

	return nil
}
