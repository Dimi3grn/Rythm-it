package models

import (
	"time"
)

// Conversation représente une conversation entre deux utilisateurs
type Conversation struct {
	ID            uint       `json:"id" db:"id"`
	User1ID       uint       `json:"user1_id" db:"user1_id"`
	User2ID       uint       `json:"user2_id" db:"user2_id"`
	LastMessageID *uint      `json:"last_message_id" db:"last_message_id"`
	LastMessageAt *time.Time `json:"last_message_at" db:"last_message_at"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	User1       *User          `json:"user1,omitempty"`
	User2       *User          `json:"user2,omitempty"`
	LastMessage *DirectMessage `json:"last_message,omitempty"`
}

// DirectMessage représente un message direct entre deux utilisateurs
type DirectMessage struct {
	ID             uint       `json:"id" db:"id"`
	ConversationID uint       `json:"conversation_id" db:"conversation_id"`
	SenderID       uint       `json:"sender_id" db:"sender_id"`
	ReceiverID     uint       `json:"receiver_id" db:"receiver_id"`
	Content        string     `json:"content" db:"content"`
	IsRead         bool       `json:"is_read" db:"is_read"`
	ReadAt         *time.Time `json:"read_at" db:"read_at"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// Relations (chargées séparément)
	Sender   *User `json:"sender,omitempty"`
	Receiver *User `json:"receiver,omitempty"`
}

// ConversationPresence représente la présence d'un utilisateur dans une conversation
type ConversationPresence struct {
	ID             uint      `json:"id" db:"id"`
	ConversationID uint      `json:"conversation_id" db:"conversation_id"`
	UserID         uint      `json:"user_id" db:"user_id"`
	IsTyping       bool      `json:"is_typing" db:"is_typing"`
	LastSeenAt     time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

// ConversationWithDetails représente une conversation avec toutes ses informations
type ConversationWithDetails struct {
	Conversation
	OtherUser       *User  `json:"other_user"`
	UnreadCount     int    `json:"unread_count"`
	LastMessageText string `json:"last_message_text"`
	IsOnline        bool   `json:"is_online"`
	IsTyping        bool   `json:"is_typing"`
}

// WebSocketMessage représente un message WebSocket
type WebSocketMessage struct {
	Type           string                 `json:"type"` // "message", "typing", "read", "status"
	ConversationID uint                   `json:"conversation_id,omitempty"`
	Message        *DirectMessage         `json:"message,omitempty"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Timestamp      time.Time              `json:"timestamp"`
}

