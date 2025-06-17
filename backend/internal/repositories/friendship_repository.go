package repositories

import (
	"database/sql"
	"fmt"
	"rythmitbackend/internal/models"
	"time"
)

// FriendshipRepository interface pour les opérations sur les amitiés
type FriendshipRepository interface {
	// Demandes d'amitié
	SendFriendRequest(requesterID, addresseeID uint) error
	AcceptFriendRequest(requesterID, addresseeID uint) error
	RejectFriendRequest(requesterID, addresseeID uint) error
	BlockUser(blockerID, blockedID uint) error
	UnblockUser(blockerID, blockedID uint) error

	// Récupération des données
	GetFriendshipStatus(userID1, userID2 uint) (*string, error)
	GetFriends(userID uint) ([]*models.Friend, error)
	GetFriendRequests(userID uint) ([]*models.FriendRequest, error)
	GetSentRequests(userID uint) ([]*models.FriendRequest, error)

	// Recherche d'utilisateurs
	SearchUsers(query string, currentUserID uint, limit int) ([]*models.UserSearchResult, error)
	GetMutualFriends(userID1, userID2 uint) ([]*models.Friend, error)
	GetMutualFriendsCount(userID1, userID2 uint) (int, error)

	// Statistiques
	GetFriendsCount(userID uint) (int, error)
	GetPendingRequestsCount(userID uint) (int, error)

	// Vérifications
	AreFriends(userID1, userID2 uint) (bool, error)
	HasPendingRequest(requesterID, addresseeID uint) (bool, error)
	IsBlocked(blockerID, blockedID uint) (bool, error)

	// Gestion
	RemoveFriend(userID1, userID2 uint) error
	GetFriendship(userID1, userID2 uint) (*models.Friendship, error)
}

// friendshipRepository implémentation concrète
type friendshipRepository struct {
	*BaseRepository
}

// NewFriendshipRepository crée une nouvelle instance du repository
func NewFriendshipRepository(db *sql.DB) FriendshipRepository {
	return &friendshipRepository{
		BaseRepository: NewBaseRepository(db),
	}
}

// SendFriendRequest envoie une demande d'amitié
func (r *friendshipRepository) SendFriendRequest(requesterID, addresseeID uint) error {
	// Vérifier qu'il n'y a pas déjà une relation
	existing, err := r.GetFriendship(requesterID, addresseeID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("erreur vérification relation existante: %w", err)
	}

	if existing != nil {
		return fmt.Errorf("une relation existe déjà entre ces utilisateurs")
	}

	query := `
		INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
	`

	_, err = r.DB.Exec(query, requesterID, addresseeID, models.FriendshipStatusPending)
	if err != nil {
		return fmt.Errorf("erreur envoi demande d'amitié: %w", err)
	}

	return nil
}

// AcceptFriendRequest accepte une demande d'amitié
func (r *friendshipRepository) AcceptFriendRequest(requesterID, addresseeID uint) error {
	query := `
		UPDATE friendships 
		SET status = ?, updated_at = NOW()
		WHERE requester_id = ? AND addressee_id = ? AND status = ?
	`

	result, err := r.DB.Exec(query, models.FriendshipStatusAccepted, requesterID, addresseeID, models.FriendshipStatusPending)
	if err != nil {
		return fmt.Errorf("erreur acceptation demande d'amitié: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification acceptation: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucune demande d'amitié en attente trouvée")
	}

	return nil
}

// RejectFriendRequest rejette une demande d'amitié
func (r *friendshipRepository) RejectFriendRequest(requesterID, addresseeID uint) error {
	query := `
		DELETE FROM friendships 
		WHERE requester_id = ? AND addressee_id = ? AND status = ?
	`

	result, err := r.DB.Exec(query, requesterID, addresseeID, models.FriendshipStatusPending)
	if err != nil {
		return fmt.Errorf("erreur rejet demande d'amitié: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("erreur vérification rejet: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("aucune demande d'amitié en attente trouvée")
	}

	return nil
}

// BlockUser bloque un utilisateur
func (r *friendshipRepository) BlockUser(blockerID, blockedID uint) error {
	// Supprimer toute relation existante
	r.RemoveFriend(blockerID, blockedID)

	query := `
		INSERT INTO friendships (requester_id, addressee_id, status, created_at, updated_at)
		VALUES (?, ?, ?, NOW(), NOW())
		ON DUPLICATE KEY UPDATE status = ?, updated_at = NOW()
	`

	_, err := r.DB.Exec(query, blockerID, blockedID, models.FriendshipStatusBlocked, models.FriendshipStatusBlocked)
	if err != nil {
		return fmt.Errorf("erreur blocage utilisateur: %w", err)
	}

	return nil
}

// UnblockUser débloque un utilisateur
func (r *friendshipRepository) UnblockUser(blockerID, blockedID uint) error {
	query := `
		DELETE FROM friendships 
		WHERE requester_id = ? AND addressee_id = ? AND status = ?
	`

	_, err := r.DB.Exec(query, blockerID, blockedID, models.FriendshipStatusBlocked)
	if err != nil {
		return fmt.Errorf("erreur déblocage utilisateur: %w", err)
	}

	return nil
}

// GetFriendshipStatus récupère le statut d'amitié entre deux utilisateurs
func (r *friendshipRepository) GetFriendshipStatus(userID1, userID2 uint) (*string, error) {
	query := `
		SELECT status FROM friendships 
		WHERE (requester_id = ? AND addressee_id = ?) 
		   OR (requester_id = ? AND addressee_id = ?)
		LIMIT 1
	`

	var status string
	err := r.DB.QueryRow(query, userID1, userID2, userID2, userID1).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Aucune relation
		}
		return nil, fmt.Errorf("erreur récupération statut amitié: %w", err)
	}

	return &status, nil
}

// GetFriends récupère la liste des amis d'un utilisateur
func (r *friendshipRepository) GetFriends(userID uint) ([]*models.Friend, error) {
	query := `
		SELECT DISTINCT
			u.id, u.username, u.profile_pic, u.last_seen,
			f.created_at as friendship_date
		FROM friendships f
		JOIN users u ON (
			(f.requester_id = ? AND f.addressee_id = u.id) OR
			(f.addressee_id = ? AND f.requester_id = u.id)
		)
		WHERE f.status = ? AND u.id != ?
		ORDER BY u.username ASC
	`

	rows, err := r.DB.Query(query, userID, userID, models.FriendshipStatusAccepted, userID)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération amis: %w", err)
	}
	defer rows.Close()

	var friends []*models.Friend
	for rows.Next() {
		friend := &models.Friend{}
		var lastSeen sql.NullTime

		err := rows.Scan(
			&friend.ID, &friend.Username, &friend.Avatar, &lastSeen,
			&friend.FriendshipDate,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan ami: %w", err)
		}

		if lastSeen.Valid {
			friend.LastSeen = &lastSeen.Time
		}

		// Déterminer le statut en ligne (logique simple basée sur last_seen)
		if friend.LastSeen != nil {
			timeSince := time.Since(*friend.LastSeen)
			if timeSince < 5*time.Minute {
				friend.OnlineStatus = "online"
			} else if timeSince < 30*time.Minute {
				friend.OnlineStatus = "away"
			} else {
				friend.OnlineStatus = "offline"
			}
		} else {
			friend.OnlineStatus = "offline"
		}

		// Récupérer le nombre d'amis mutuels
		mutualCount, _ := r.GetMutualFriendsCount(userID, friend.ID)
		friend.MutualFriends = mutualCount

		friends = append(friends, friend)
	}

	return friends, nil
}

// GetFriendRequests récupère les demandes d'amitié reçues
func (r *friendshipRepository) GetFriendRequests(userID uint) ([]*models.FriendRequest, error) {
	query := `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at,
		       u.username, u.profile_pic
		FROM friendships f
		JOIN users u ON f.requester_id = u.id
		WHERE f.addressee_id = ? AND f.status = ?
		ORDER BY f.created_at DESC
	`

	rows, err := r.DB.Query(query, userID, models.FriendshipStatusPending)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération demandes d'amitié: %w", err)
	}
	defer rows.Close()

	var requests []*models.FriendRequest
	for rows.Next() {
		request := &models.FriendRequest{}
		var avatar sql.NullString

		err := rows.Scan(
			&request.ID, &request.RequesterID, &request.AddresseeID,
			&request.Status, &request.CreatedAt,
			&request.RequesterUsername, &avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan demande d'amitié: %w", err)
		}

		if avatar.Valid {
			request.RequesterAvatar = &avatar.String
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// GetSentRequests récupère les demandes d'amitié envoyées
func (r *friendshipRepository) GetSentRequests(userID uint) ([]*models.FriendRequest, error) {
	query := `
		SELECT f.id, f.requester_id, f.addressee_id, f.status, f.created_at,
		       u.username, u.profile_pic
		FROM friendships f
		JOIN users u ON f.addressee_id = u.id
		WHERE f.requester_id = ? AND f.status = ?
		ORDER BY f.created_at DESC
	`

	rows, err := r.DB.Query(query, userID, models.FriendshipStatusPending)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération demandes envoyées: %w", err)
	}
	defer rows.Close()

	var requests []*models.FriendRequest
	for rows.Next() {
		request := &models.FriendRequest{}
		var avatar sql.NullString

		err := rows.Scan(
			&request.ID, &request.RequesterID, &request.AddresseeID,
			&request.Status, &request.CreatedAt,
			&request.RequesterUsername, &avatar,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan demande envoyée: %w", err)
		}

		if avatar.Valid {
			request.RequesterAvatar = &avatar.String
		}

		requests = append(requests, request)
	}

	return requests, nil
}

// SearchUsers recherche des utilisateurs par nom d'utilisateur
func (r *friendshipRepository) SearchUsers(query string, currentUserID uint, limit int) ([]*models.UserSearchResult, error) {
	searchTerm := "%" + query + "%"

	sqlQuery := `
		SELECT DISTINCT u.id, u.username, u.profile_pic,
		       f.status as friendship_status
		FROM users u
		LEFT JOIN friendships f ON (
			(f.requester_id = ? AND f.addressee_id = u.id) OR
			(f.addressee_id = ? AND f.requester_id = u.id)
		)
		WHERE u.username LIKE ? AND u.id != ?
		ORDER BY u.username ASC
		LIMIT ?
	`

	rows, err := r.DB.Query(sqlQuery, currentUserID, currentUserID, searchTerm, currentUserID, limit)
	if err != nil {
		return nil, fmt.Errorf("erreur recherche utilisateurs: %w", err)
	}
	defer rows.Close()

	var users []*models.UserSearchResult
	for rows.Next() {
		user := &models.UserSearchResult{}
		var avatar sql.NullString
		var friendshipStatus sql.NullString

		err := rows.Scan(
			&user.ID, &user.Username, &avatar, &friendshipStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("erreur scan utilisateur: %w", err)
		}

		if avatar.Valid {
			user.Avatar = &avatar.String
		}

		if friendshipStatus.Valid {
			user.FriendshipStatus = &friendshipStatus.String
		}

		// Récupérer le nombre d'amis mutuels
		mutualCount, _ := r.GetMutualFriendsCount(currentUserID, user.ID)
		user.MutualFriends = mutualCount

		users = append(users, user)
	}

	return users, nil
}

// GetMutualFriends récupère les amis mutuels entre deux utilisateurs
func (r *friendshipRepository) GetMutualFriends(userID1, userID2 uint) ([]*models.Friend, error) {
	query := `
		SELECT DISTINCT u.id, u.username, u.profile_pic
		FROM users u
		WHERE u.id IN (
			SELECT CASE 
				WHEN f1.requester_id = ? THEN f1.addressee_id
				ELSE f1.requester_id
			END as friend_id
			FROM friendships f1
			WHERE ((f1.requester_id = ? AND f1.addressee_id != ?) OR 
				   (f1.addressee_id = ? AND f1.requester_id != ?))
			  AND f1.status = ?
		) AND u.id IN (
			SELECT CASE 
				WHEN f2.requester_id = ? THEN f2.addressee_id
				ELSE f2.requester_id
			END as friend_id
			FROM friendships f2
			WHERE ((f2.requester_id = ? AND f2.addressee_id != ?) OR 
				   (f2.addressee_id = ? AND f2.requester_id != ?))
			  AND f2.status = ?
		)
		ORDER BY u.username ASC
	`

	rows, err := r.DB.Query(query,
		userID1, userID1, userID2, userID1, userID2, models.FriendshipStatusAccepted,
		userID2, userID2, userID1, userID2, userID1, models.FriendshipStatusAccepted,
	)
	if err != nil {
		return nil, fmt.Errorf("erreur récupération amis mutuels: %w", err)
	}
	defer rows.Close()

	var friends []*models.Friend
	for rows.Next() {
		friend := &models.Friend{}
		var avatar sql.NullString

		err := rows.Scan(&friend.ID, &friend.Username, &avatar)
		if err != nil {
			return nil, fmt.Errorf("erreur scan ami mutuel: %w", err)
		}

		if avatar.Valid {
			friend.Avatar = &avatar.String
		}

		friends = append(friends, friend)
	}

	return friends, nil
}

// GetMutualFriendsCount compte les amis mutuels entre deux utilisateurs
func (r *friendshipRepository) GetMutualFriendsCount(userID1, userID2 uint) (int, error) {
	query := `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		WHERE u.id IN (
			SELECT CASE 
				WHEN f1.requester_id = ? THEN f1.addressee_id
				ELSE f1.requester_id
			END as friend_id
			FROM friendships f1
			WHERE ((f1.requester_id = ? AND f1.addressee_id != ?) OR 
				   (f1.addressee_id = ? AND f1.requester_id != ?))
			  AND f1.status = ?
		) AND u.id IN (
			SELECT CASE 
				WHEN f2.requester_id = ? THEN f2.addressee_id
				ELSE f2.requester_id
			END as friend_id
			FROM friendships f2
			WHERE ((f2.requester_id = ? AND f2.addressee_id != ?) OR 
				   (f2.addressee_id = ? AND f2.requester_id != ?))
			  AND f2.status = ?
		)
	`

	var count int
	err := r.DB.QueryRow(query,
		userID1, userID1, userID2, userID1, userID2, models.FriendshipStatusAccepted,
		userID2, userID2, userID1, userID2, userID1, models.FriendshipStatusAccepted,
	).Scan(&count)

	if err != nil {
		return 0, fmt.Errorf("erreur comptage amis mutuels: %w", err)
	}

	return count, nil
}

// GetFriendsCount compte le nombre d'amis d'un utilisateur
func (r *friendshipRepository) GetFriendsCount(userID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM friendships f
		WHERE ((f.requester_id = ? AND f.addressee_id != ?) OR 
			   (f.addressee_id = ? AND f.requester_id != ?))
		  AND f.status = ?
	`

	var count int
	err := r.DB.QueryRow(query, userID, userID, userID, userID, models.FriendshipStatusAccepted).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage amis: %w", err)
	}

	return count, nil
}

// GetPendingRequestsCount compte les demandes d'amitié en attente
func (r *friendshipRepository) GetPendingRequestsCount(userID uint) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM friendships
		WHERE addressee_id = ? AND status = ?
	`

	var count int
	err := r.DB.QueryRow(query, userID, models.FriendshipStatusPending).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("erreur comptage demandes en attente: %w", err)
	}

	return count, nil
}

// AreFriends vérifie si deux utilisateurs sont amis
func (r *friendshipRepository) AreFriends(userID1, userID2 uint) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM friendships
		WHERE ((requester_id = ? AND addressee_id = ?) OR 
			   (requester_id = ? AND addressee_id = ?))
		  AND status = ?
	`

	var count int
	err := r.DB.QueryRow(query, userID1, userID2, userID2, userID1, models.FriendshipStatusAccepted).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification amitié: %w", err)
	}

	return count > 0, nil
}

// HasPendingRequest vérifie s'il y a une demande d'amitié en attente
func (r *friendshipRepository) HasPendingRequest(requesterID, addresseeID uint) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM friendships
		WHERE requester_id = ? AND addressee_id = ? AND status = ?
	`

	var count int
	err := r.DB.QueryRow(query, requesterID, addresseeID, models.FriendshipStatusPending).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification demande en attente: %w", err)
	}

	return count > 0, nil
}

// IsBlocked vérifie si un utilisateur est bloqué
func (r *friendshipRepository) IsBlocked(blockerID, blockedID uint) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM friendships
		WHERE requester_id = ? AND addressee_id = ? AND status = ?
	`

	var count int
	err := r.DB.QueryRow(query, blockerID, blockedID, models.FriendshipStatusBlocked).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("erreur vérification blocage: %w", err)
	}

	return count > 0, nil
}

// RemoveFriend supprime une amitié
func (r *friendshipRepository) RemoveFriend(userID1, userID2 uint) error {
	query := `
		DELETE FROM friendships
		WHERE ((requester_id = ? AND addressee_id = ?) OR 
			   (requester_id = ? AND addressee_id = ?))
		  AND status = ?
	`

	_, err := r.DB.Exec(query, userID1, userID2, userID2, userID1, models.FriendshipStatusAccepted)
	if err != nil {
		return fmt.Errorf("erreur suppression amitié: %w", err)
	}

	return nil
}

// GetFriendship récupère une relation d'amitié entre deux utilisateurs
func (r *friendshipRepository) GetFriendship(userID1, userID2 uint) (*models.Friendship, error) {
	query := `
		SELECT id, requester_id, addressee_id, status, created_at, updated_at
		FROM friendships
		WHERE (requester_id = ? AND addressee_id = ?) OR 
			  (requester_id = ? AND addressee_id = ?)
		LIMIT 1
	`

	friendship := &models.Friendship{}
	err := r.DB.QueryRow(query, userID1, userID2, userID2, userID1).Scan(
		&friendship.ID, &friendship.RequesterID, &friendship.AddresseeID,
		&friendship.Status, &friendship.CreatedAt, &friendship.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, sql.ErrNoRows
		}
		return nil, fmt.Errorf("erreur récupération amitié: %w", err)
	}

	return friendship, nil
}
