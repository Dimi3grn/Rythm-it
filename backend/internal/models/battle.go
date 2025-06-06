package models

import "time"

// Battle représente une battle de musique
type Battle struct {
	ID          uint            `json:"id"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	State       string          `json:"state"` // ex: 'active', 'finished'
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Options     []*BattleOption `json:"options"` // Les options (musiques) pour cette battle
}

// BattleOption représente une option de vote (musique) dans une battle
type BattleOption struct {
	ID        uint   `json:"id"`
	BattleID  uint   `json:"battle_id"`
	Title     string `json:"title"`      // Titre de la musique
	Artist    string `json:"artist"`     // Artiste
	MusicURL  string `json:"music_url"`  // URL ou chemin du fichier musique
	ImageURL  string `json:"image_url"`  // URL ou chemin de l'image associée
	VoteCount int    `json:"vote_count"` // Compte des votes pour cette option (calculé)
}

// BattleVote représente le vote d'un utilisateur pour une option dans une battle
type BattleVote struct {
	BattleID  uint      `json:"battle_id"`
	UserID    uint      `json:"user_id"`
	OptionID  uint      `json:"option_id"`
	CreatedAt time.Time `json:"created_at"`
}
