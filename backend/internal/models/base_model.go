package models

import (
	"errors"
	"fmt"
	"rythmitbackend/internal/utils"
	"strings"
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

// UserProfile modèle pour les profils utilisateur personnalisés
type UserProfile struct {
	BaseModel
	UserID      uint    `json:"user_id" db:"user_id" validate:"required"`
	DisplayName *string `json:"display_name" db:"display_name" validate:"omitempty,max=100"`
	AvatarImage *string `json:"avatar_image" db:"avatar_image" validate:"omitempty"`
	BannerImage *string `json:"banner_image" db:"banner_image" validate:"omitempty"`
}

// Thread modèle fil de discussion musical
type Thread struct {
	BaseModel
	Title       string  `json:"title" db:"title" validate:"required,min=5,max=200"`
	Description string  `json:"description" db:"desc_" validate:"required,min=10"`
	ImageURL    *string `json:"image_url" db:"image_url" validate:"omitempty"`
	State       string  `json:"state" db:"state" validate:"oneof=ouvert fermé archivé"`
	Visibility  string  `json:"visibility" db:"visibility" validate:"oneof=public privé"`
	UserID      uint    `json:"user_id" db:"user_id"`
	Author      *User   `json:"author,omitempty"`
	Tags        []*Tag  `json:"tags,omitempty"`
	FireCount   int     `json:"fire_count"` // Compteur Fire 🔥
	SkipCount   int     `json:"skip_count"` // Compteur Skip ⏭️
}

// Message modèle pour les messages
type Message struct {
	BaseModel
	Content         string         `json:"content" db:"content" validate:"required,min=1,max=5000,nohtml"`
	ImageURL        *string        `json:"image_url" db:"image_url" validate:"omitempty"`
	ThreadID        uint           `json:"thread_id" db:"thread_id" validate:"required"`
	UserID          uint           `json:"user_id" db:"user_id" validate:"required"`
	Author          *User          `json:"author,omitempty"`
	PopularityScore int            `json:"popularity_score"` // Fire - Skip
	UserVote        *string        `json:"user_vote,omitempty" validate:"omitempty,oneof=fire skip neutral"`
	Embeds          *MessageEmbeds `json:"embeds,omitempty" validate:"omitempty,dive"`
}

// MessageEmbeds embeds YouTube/Spotify dans les messages
type MessageEmbeds struct {
	YouTube *string `json:"youtube,omitempty" validate:"omitempty,url,youtube_url"`
	Spotify *string `json:"spotify,omitempty" validate:"omitempty,url,spotify_url"`
}

// ValidateMessageEmbeds valide les embeds d'un message
func ValidateMessageEmbeds(embeds *MessageEmbeds) error {
	if embeds == nil {
		return nil
	}

	// Vérifier qu'au moins un embed est présent
	if embeds.YouTube == nil && embeds.Spotify == nil {
		return errors.New("au moins un embed (YouTube ou Spotify) doit être présent")
	}

	// Vérifier les URLs YouTube
	if embeds.YouTube != nil {
		if !strings.Contains(*embeds.YouTube, "youtube.com") && !strings.Contains(*embeds.YouTube, "youtu.be") {
			return errors.New("URL YouTube invalide")
		}
	}

	// Vérifier les URLs Spotify
	if embeds.Spotify != nil {
		if !strings.Contains(*embeds.Spotify, "spotify.com") {
			return errors.New("URL Spotify invalide")
		}
	}

	return nil
}

// ValidateMessage valide un message complet
func ValidateMessage(msg *Message) error {
	if msg == nil {
		return errors.New("message ne peut pas être nil")
	}

	// Valider la structure de base
	validationErrors := utils.ValidateStruct(msg)
	if len(validationErrors) > 0 {
		// Convertir les erreurs de validation en une seule erreur
		var errMsgs []string
		for _, err := range validationErrors {
			errMsgs = append(errMsgs, err.Message)
		}
		return fmt.Errorf("erreurs de validation: %s", strings.Join(errMsgs, "; "))
	}

	// Valider les embeds si présents
	if err := ValidateMessageEmbeds(msg.Embeds); err != nil {
		return err
	}

	// Vérifier que le contenu n'est pas vide après nettoyage
	content := strings.TrimSpace(msg.Content)
	if content == "" {
		return errors.New("le contenu du message ne peut pas être vide")
	}

	// Vérifier que le contenu n'est pas trop court
	if len(content) < 2 {
		return errors.New("le contenu du message est trop court")
	}

	// Vérifier que le contenu n'est pas trop long
	if len(content) > 5000 {
		return errors.New("le contenu du message est trop long (max 5000 caractères)")
	}

	// Vérifier que le vote est valide si présent
	if msg.UserVote != nil {
		switch *msg.UserVote {
		case "fire", "skip", "neutral":
			// OK
		default:
			return errors.New("vote invalide")
		}
	}

	return nil
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

// Battle modèle pour les battles de rap

// BattleState représente les différents états possibles d'une battle
const (
	BattleStateActive    = "active"
	BattleStateFinished  = "finished"
	BattleStateCancelled = "cancelled"
)

// ValidateBattleState vérifie si l'état de la battle est valide
func ValidateBattleState(state string) bool {
	switch state {
	case BattleStateActive, BattleStateFinished, BattleStateCancelled:
		return true
	default:
		return false
	}
}

// PaginationParams paramètres de pagination
type PaginationParams struct {
	Page    int    `json:"page"`
	PerPage int    `json:"per_page"`
	Sort    string `json:"sort"`
	Order   string `json:"order"`
}

// DefaultPagination retourne les paramètres de pagination par défaut
func DefaultPagination() PaginationParams {
	return PaginationParams{
		Page:    1,
		PerPage: 10,
		Sort:    "id",
		Order:   "DESC",
	}
}

// ValidatePagination valide et normalise les paramètres de pagination
func ValidatePagination(params *PaginationParams) {
	if params.Page < 1 {
		params.Page = 1
	}

	if params.PerPage < 1 || params.PerPage > 100 {
		params.PerPage = 10
	}

	if params.Order != "ASC" && params.Order != "DESC" {
		params.Order = "DESC"
	}
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
