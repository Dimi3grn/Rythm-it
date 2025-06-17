package models

import (
	"time"
)

// Friendship représente une relation d'amitié entre deux utilisateurs
type Friendship struct {
	ID          uint      `json:"id" db:"id"`
	RequesterID uint      `json:"requester_id" db:"requester_id"` // Utilisateur qui a envoyé la demande
	AddresseeID uint      `json:"addressee_id" db:"addressee_id"` // Utilisateur qui a reçu la demande
	Status      string    `json:"status" db:"status"`             // pending, accepted, blocked
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Requester *User `json:"requester,omitempty"`
	Addressee *User `json:"addressee,omitempty"`
}

// FriendshipStatus définit les statuts possibles d'une amitié
type FriendshipStatus string

const (
	FriendshipStatusPending  FriendshipStatus = "pending"
	FriendshipStatusAccepted FriendshipStatus = "accepted"
	FriendshipStatusBlocked  FriendshipStatus = "blocked"
)

// FriendRequest représente une demande d'amitié
type FriendRequest struct {
	ID          uint      `json:"id"`
	RequesterID uint      `json:"requester_id"`
	AddresseeID uint      `json:"addressee_id"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`

	// Informations de l'utilisateur qui a fait la demande
	RequesterUsername string  `json:"requester_username"`
	RequesterAvatar   *string `json:"requester_avatar"`
}

// Friend représente un ami avec ses informations et statut
type Friend struct {
	ID           uint       `json:"id"`
	Username     string     `json:"username"`
	Avatar       *string    `json:"avatar"`
	OnlineStatus string     `json:"online_status"` // online, away, offline
	LastSeen     *time.Time `json:"last_seen"`
	Activity     *string    `json:"activity"` // Ce que fait l'utilisateur actuellement

	// Statistiques d'amitié
	FriendshipDate time.Time `json:"friendship_date"`
	MutualFriends  int       `json:"mutual_friends"`
}

// UserSearchResult représente un utilisateur dans les résultats de recherche
type UserSearchResult struct {
	ID               uint    `json:"id"`
	Username         string  `json:"username"`
	Avatar           *string `json:"avatar"`
	FriendshipStatus *string `json:"friendship_status"` // null, pending, accepted, blocked
	MutualFriends    int     `json:"mutual_friends"`
}
