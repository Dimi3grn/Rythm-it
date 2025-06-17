package repositories

import (
	"database/sql"
	"fmt"
)

// LikeRepository interface pour la gestion des likes
type LikeRepository interface {
	LikeThread(userID, threadID uint) error
	UnlikeThread(userID, threadID uint) error
	IsThreadLikedByUser(userID, threadID uint) (bool, error)
	GetThreadLikesCount(threadID uint) (int, error)
	GetUserLikedThreads(userID uint) ([]uint, error)
	UpdateThreadLikesCount(threadID uint) error
}

type likeRepository struct {
	*BaseRepository
}

// NewLikeRepository crée une nouvelle instance du repository
func NewLikeRepository(db *sql.DB) LikeRepository {
	return &likeRepository{
		BaseRepository: &BaseRepository{DB: db},
	}
}

// LikeThread ajoute un like à un thread par un utilisateur
func (r *likeRepository) LikeThread(userID, threadID uint) error {
	// Vérifier si l'utilisateur a déjà liké ce thread
	liked, err := r.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return fmt.Errorf("erreur vérification like existant: %w", err)
	}

	if liked {
		return fmt.Errorf("utilisateur a déjà liké ce thread")
	}

	// Transaction pour ajouter le like et mettre à jour le compteur
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	// Insérer le like
	query := "INSERT INTO thread_likes (user_id, thread_id) VALUES (?, ?)"
	_, err = tx.Exec(query, userID, threadID)
	if err != nil {
		return fmt.Errorf("erreur insertion like: %w", err)
	}

	// Mettre à jour le compteur de likes du thread
	updateQuery := "UPDATE threads SET likes_count = likes_count + 1 WHERE id = ?"
	_, err = tx.Exec(updateQuery, threadID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour compteur likes: %w", err)
	}

	return tx.Commit()
}

// UnlikeThread supprime un like d'un thread par un utilisateur
func (r *likeRepository) UnlikeThread(userID, threadID uint) error {
	// Vérifier si l'utilisateur a bien liké ce thread
	liked, err := r.IsThreadLikedByUser(userID, threadID)
	if err != nil {
		return fmt.Errorf("erreur vérification like existant: %w", err)
	}

	if !liked {
		return fmt.Errorf("utilisateur n'a pas liké ce thread")
	}

	// Transaction pour supprimer le like et mettre à jour le compteur
	tx, err := r.DB.Begin()
	if err != nil {
		return fmt.Errorf("erreur début transaction: %w", err)
	}
	defer tx.Rollback()

	// Supprimer le like
	query := "DELETE FROM thread_likes WHERE user_id = ? AND thread_id = ?"
	_, err = tx.Exec(query, userID, threadID)
	if err != nil {
		return fmt.Errorf("erreur suppression like: %w", err)
	}

	// Mettre à jour le compteur de likes du thread
	updateQuery := "UPDATE threads SET likes_count = GREATEST(likes_count - 1, 0) WHERE id = ?"
	_, err = tx.Exec(updateQuery, threadID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour compteur likes: %w", err)
	}

	return tx.Commit()
}

// IsThreadLikedByUser vérifie si un utilisateur a liké un thread
func (r *likeRepository) IsThreadLikedByUser(userID, threadID uint) (bool, error) {
	query := "SELECT COUNT(*) FROM thread_likes WHERE user_id = ? AND thread_id = ?"

	var count int
	err := r.DB.QueryRow(query, userID, threadID).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification like: %w", err)
	}

	return count > 0, nil
}

// GetThreadLikesCount récupère le nombre de likes d'un thread
func (r *likeRepository) GetThreadLikesCount(threadID uint) (int, error) {
	query := "SELECT likes_count FROM threads WHERE id = ?"

	var count int
	err := r.DB.QueryRow(query, threadID).Scan(&count)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("thread non trouvé")
		}
		return 0, fmt.Errorf("erreur récupération compteur likes: %w", err)
	}

	return count, nil
}

// GetUserLikedThreads récupère la liste des threads likés par un utilisateur
func (r *likeRepository) GetUserLikedThreads(userID uint) ([]uint, error) {
	query := "SELECT thread_id FROM thread_likes WHERE user_id = ? ORDER BY created_at DESC"

	rows, err := r.DB.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération threads likés: %w", err)
	}
	defer rows.Close()

	var threadIDs []uint
	for rows.Next() {
		var threadID uint
		err := rows.Scan(&threadID)
		if err != nil {
			return nil, fmt.Errorf("erreur scan thread ID: %w", err)
		}
		threadIDs = append(threadIDs, threadID)
	}

	return threadIDs, nil
}

// UpdateThreadLikesCount met à jour le compteur de likes d'un thread en comptant les likes réels
func (r *likeRepository) UpdateThreadLikesCount(threadID uint) error {
	query := `
		UPDATE threads 
		SET likes_count = (
			SELECT COUNT(*) 
			FROM thread_likes 
			WHERE thread_id = ?
		) 
		WHERE id = ?
	`

	_, err := r.DB.Exec(query, threadID, threadID)
	if err != nil {
		return fmt.Errorf("erreur mise à jour compteur likes: %w", err)
	}

	return nil
}
