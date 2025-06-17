package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"time"
)

// ProfileRepository interface définissant les opérations sur les profils utilisateur
type ProfileRepository interface {
	// CRUD operations
	Create(profile *models.UserProfile) error
	FindByUserID(userID uint) (*models.UserProfile, error)
	Update(profile *models.UserProfile) error
	Delete(userID uint) error

	// Convenience methods
	CreateOrUpdate(profile *models.UserProfile) error
	ExistsByUserID(userID uint) (bool, error)
}

// profileRepository implémentation concrète
type profileRepository struct {
	*BaseRepository
}

// NewProfileRepository crée une nouvelle instance du repository profil
func NewProfileRepository(db *sql.DB) ProfileRepository {
	return &profileRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// Create insère un nouveau profil utilisateur en base
func (r *profileRepository) Create(profile *models.UserProfile) error {
	query := `
		INSERT INTO user_profiles (user_id, display_name, avatar_image, banner_image, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`

	now := time.Now()
	profile.CreatedAt = now
	profile.UpdatedAt = now

	result, err := r.DB.Exec(query,
		profile.UserID,
		profile.DisplayName,
		profile.AvatarImage,
		profile.BannerImage,
		profile.CreatedAt,
		profile.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("erreur création profil utilisateur: %w", err)
	}

	// Récupérer l'ID généré
	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("erreur récupération ID profil: %w", err)
	}

	profile.ID = uint(id)
	return nil
}

// FindByUserID trouve un profil par l'ID utilisateur
func (r *profileRepository) FindByUserID(userID uint) (*models.UserProfile, error) {
	profile := &models.UserProfile{}
	query := `
		SELECT id, user_id, display_name, avatar_image, banner_image, created_at, updated_at
		FROM user_profiles 
		WHERE user_id = ?
	`

	err := r.DB.QueryRow(query, userID).Scan(
		&profile.ID,
		&profile.UserID,
		&profile.DisplayName,
		&profile.AvatarImage,
		&profile.BannerImage,
		&profile.CreatedAt,
		&profile.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("profil pour utilisateur ID %d non trouvé", userID)
		}
		return nil, fmt.Errorf("erreur recherche profil par userID: %w", err)
	}

	return profile, nil
}

// Update met à jour un profil utilisateur existant
func (r *profileRepository) Update(profile *models.UserProfile) error {
	query := `
		UPDATE user_profiles 
		SET display_name = ?, avatar_image = ?, banner_image = ?, updated_at = ?
		WHERE user_id = ?
	`

	profile.UpdatedAt = time.Now()

	result, err := r.DB.Exec(query,
		profile.DisplayName,
		profile.AvatarImage,
		profile.BannerImage,
		profile.UpdatedAt,
		profile.UserID,
	)

	if err != nil {
		return fmt.Errorf("erreur mise à jour profil: %w", err)
	}

	// Vérifier qu'une ligne a été affectée
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification mise à jour profil: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucun profil trouvé pour utilisateur ID %d", profile.UserID)
	}

	return nil
}

// Delete supprime un profil utilisateur
func (r *profileRepository) Delete(userID uint) error {
	query := `DELETE FROM user_profiles WHERE user_id = ?`

	result, err := r.DB.Exec(query, userID)
	if err != nil {
		return fmt.Errorf("erreur suppression profil: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification suppression profil: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucun profil trouvé pour utilisateur ID %d", userID)
	}

	return nil
}

// CreateOrUpdate crée un profil s'il n'existe pas, sinon le met à jour
func (r *profileRepository) CreateOrUpdate(profile *models.UserProfile) error {
	exists, err := r.ExistsByUserID(profile.UserID)
	if err != nil {
		return fmt.Errorf("erreur vérification existence profil: %w", err)
	}

	if exists {
		return r.Update(profile)
	} else {
		return r.Create(profile)
	}
}

// ExistsByUserID vérifie si un profil existe pour un utilisateur donné
func (r *profileRepository) ExistsByUserID(userID uint) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM user_profiles WHERE user_id = ?)`

	var exists bool
	err := r.DB.QueryRow(query, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("erreur vérification existence profil: %w", err)
	}

	return exists, nil
}
