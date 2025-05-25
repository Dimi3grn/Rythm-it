package models

import (
	"time"
)

// BaseModel structure de base pour tous les modèles
type BaseModel struct {
	ID        uint      `json:"id" db:"id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// User modèle utilisateur Rythmit
type User struct {
	BaseModel
	Username        string     `json:"username" db:"username" validate:"required,min=3,max=30"`
	Email           string     `json:"email" db:"email" validate:"required,email"`
	Password        string     `json:"-" db:"password"` // jamais exposé en JSON
	IsAdmin         bool       `json:"is_admin" db:"is_admin"`
	ProfilePic      *string    `json:"profile_pic" db:"profile_pic"`
	Biography       *string    `json:"biography" db:"biography"`
	LastConnection  *time.Time `json:"last_connection" db:"last_connection"`
	MessageCount    int        `json:"message_count" db:"message_count"`
	ThreadCount     int        `json:"thread_count" db:"thread_count"`
	FavoriteGenres  []string   `json:"favorite_genres,omitempty"`  // Spécifique Rythmit
	FavoriteArtists []string   `json:"favorite_artists,omitempty"` // Spécifique Rythmit
}

// Thread modèle fil de discussion musical
type Thread struct {
	BaseModel
	Title       string `json:"title" db:"title" validate:"required,min=5,max=200"`
	Description string `json:"description" db:"desc_" validate:"required,min=10"`
	State       string `json:"state" db:"state" validate:"oneof=ouvert fermé archivé"`
	Visibility  string `json:"visibility" db:"visibility" validate:"oneof=public privé"`
	UserID      uint   `json:"user_id" db:"user_id"`
	Author      *User  `json:"author,omitempty"`
	Tags        []Tag  `json:"tags,omitempty"`
	FireCount   int    `json:"fire_count"` // Compteur Fire 🔥
	SkipCount   int    `json:"skip_count"` // Compteur Skip ⏭️
}

// Message modèle message dans un thread
type Message struct {
	BaseModel
	Content         string         `json:"content" db:"content" validate:"required,min=1,max=5000"`
	ThreadID        uint           `json:"thread_id" db:"thread_id"`
	UserID          uint           `json:"user_id" db:"user_id"`
	Author          *User          `json:"author,omitempty"`
	PopularityScore int            `json:"popularity_score"`    // Fire - Skip
	UserVote        *string        `json:"user_vote,omitempty"` // "fire", "skip" ou null
	Embeds          *MessageEmbeds `json:"embeds,omitempty"`
}

// MessageEmbeds embeds YouTube/Spotify dans les messages
type MessageEmbeds struct {
	YouTube *string `json:"youtube,omitempty"`
	Spotify *string `json:"spotify,omitempty"`
}

// Tag modèle pour les tags musicaux
type Tag struct {
	ID   uint   `json:"id" db:"tag_id"`
	Name string `json:"name" db:"name" validate:"required,min=2,max=50"`
	Type string `json:"type"` // "genre", "artist", "album"
}

// LikedDisliked modèle pour les votes Fire/Skip
type LikedDisliked struct {
	UserID    uint   `json:"user_id" db:"user_id"`
	MessageID uint   `json:"message_id" db:"message_id"`
	State     string `json:"state" db:"state" validate:"oneof=fire skip neutral"`
}

// Battle modèle pour les battles musicales
type Battle struct {
	BaseModel
	Title      string         `json:"title" validate:"required"`
	Options    []BattleOption `json:"options" validate:"required,len=2"`
	TotalVotes int            `json:"total_votes"`
	Status     string         `json:"status" validate:"oneof=active ended"`
	EndDate    time.Time      `json:"end_date"`
	CreatorID  uint           `json:"creator_id"`
}

// BattleOption option dans une battle (artiste, album, etc.)
type BattleOption struct {
	ID    uint   `json:"id"`
	Name  string `json:"name" validate:"required"`
	Image string `json:"image"`
	Votes int    `json:"votes"`
}

// Constants pour les états
const (
	ThreadStateOpen     = "ouvert"
	ThreadStateClosed   = "fermé"
	ThreadStateArchived = "archivé"

	VisibilityPublic  = "public"
	VisibilityPrivate = "privé"

	VoteFire    = "fire"
	VoteSkip    = "skip"
	VoteNeutral = "neutral"

	BattleStatusActive = "active"
	BattleStatusEnded  = "ended"
)
