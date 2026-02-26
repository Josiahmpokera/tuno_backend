package domain

import "time"

type MessageType string

const (
	MessageTypeText   MessageType = "TEXT"
	MessageTypeSystem MessageType = "SYSTEM"
	MessageTypeImage  MessageType = "IMAGE"
)

type Message struct {
	ID        string      `json:"id" db:"id"`
	GroupID   string      `json:"group_id" db:"group_id"`
	SenderID  *string     `json:"sender_id" db:"sender_id"` // Nullable for system messages
	Content   string      `json:"content" db:"content"`
	Type      MessageType `json:"type" db:"type"`
	Sequence  int64       `json:"sequence_number" db:"sequence_number"`
	CreatedAt time.Time   `json:"created_at" db:"created_at"`
}

type MessageRepository interface {
	Save(message *Message) error
	GetByGroupID(groupID string, limit, offset int) ([]Message, error)
}
