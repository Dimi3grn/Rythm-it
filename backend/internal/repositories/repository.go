package repositories

import (
	"database/sql"
)

// Repository interface de base pour tous les repositories
type Repository interface {
	// Create crée une nouvelle entité
	Create(entity interface{}) error

	// FindByID trouve une entité par son ID
	FindByID(id uint, dest interface{}) error

	// Update met à jour une entité
	Update(entity interface{}) error

	// Delete supprime une entité
	Delete(id uint) error

	// FindAll récupère toutes les entités
	FindAll(dest interface{}) error
}

// BaseRepository structure de base avec connexion DB
type BaseRepository struct {
	DB *sql.DB
}

// NewBaseRepository crée un nouveau repository de base
func NewBaseRepository(db *sql.DB) *BaseRepository {
	return &BaseRepository{
		DB: db,
	}
}

// Transaction helper pour exécuter une fonction dans une transaction
func (r *BaseRepository) Transaction(fn func(*sql.Tx) error) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}

	err = fn(tx)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

// Exists vérifie si un enregistrement existe
func (r *BaseRepository) Exists(query string, args ...interface{}) (bool, error) {
	var exists bool
	err := r.DB.QueryRow(query, args...).Scan(&exists)
	if err != nil && err != sql.ErrNoRows {
		return false, err
	}
	return exists, nil
}

// Count compte le nombre d'enregistrements
func (r *BaseRepository) Count(table string, where string, args ...interface{}) (int64, error) {
	var count int64
	query := "SELECT COUNT(*) FROM " + table
	if where != "" {
		query += " WHERE " + where
	}

	err := r.DB.QueryRow(query, args...).Scan(&count)
	return count, err
}
