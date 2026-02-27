package domain

import "time"

// Conversation represents a private chat between two users.
type Conversation struct {
	ID        string    `json:"id" db:"id"`
	User1ID   string    `json:"user1_id" db:"user1_id"`
	User2ID   string    `json:"user2_id" db:"user2_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DirectMessage represents a message within a private conversation.
type DirectMessage struct {
	ID             string      `json:"id" db:"id"`
	ConversationID string      `json:"conversation_id" db:"conversation_id"`
	SenderID       string      `json:"sender_id" db:"sender_id"`
	Content        string      `json:"content" db:"content"`
	Type           MessageType `json:"type" db:"type"`
	Sequence       int64       `json:"sequence_number" db:"sequence_number"`
	Status         string      `json:"status" db:"status"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
}

// ConversationRepository defines the interface for managing conversations.
type ConversationRepository interface {
	// FindByUsers finds an existing conversation between two users.
	// Order of user1ID and user2ID should not matter.
	FindByUsers(user1ID, user2ID string) (*Conversation, error)

	// Create creates a new conversation.
	Create(conversation *Conversation) error

	// GetUserConversations retrieves all conversations for a user.
	GetUserConversations(userID string) ([]Conversation, error)

	// FindByID retrieves a conversation by its ID.
	FindByID(id string) (*Conversation, error)
}

// DirectMessageRepository defines the interface for managing direct messages.
type DirectMessageRepository interface {
	Save(message *DirectMessage) error
	GetByConversationID(conversationID string, limit, offset int) ([]DirectMessage, error)
	UpdateStatus(messageID string, status string) error
	MarkMessagesAsRead(conversationID, userID string) error
}
