package repository

import (
	"context"
	"errors"
	"fmt"

	"tuno_backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresConversationRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresConversationRepository(pool *pgxpool.Pool) domain.ConversationRepository {
	return &PostgresConversationRepository{pool: pool}
}

func (r *PostgresConversationRepository) FindByUsers(user1ID, user2ID string) (*domain.Conversation, error) {
	// Query where (user1 = u1 AND user2 = u2) OR (user1 = u2 AND user2 = u1).
	query := `
		SELECT id, user1_id, user2_id, created_at, updated_at
		FROM conversations
		WHERE (user1_id = $1 AND user2_id = $2) OR (user1_id = $2 AND user2_id = $1)
		LIMIT 1
	`

	var c domain.Conversation
	err := r.pool.QueryRow(context.Background(), query, user1ID, user2ID).Scan(
		&c.ID, &c.User1ID, &c.User2ID, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}

	return &c, nil
}

func (r *PostgresConversationRepository) FindByID(id string) (*domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, created_at, updated_at
		FROM conversations
		WHERE id = $1
	`

	var c domain.Conversation
	err := r.pool.QueryRow(context.Background(), query, id).Scan(
		&c.ID, &c.User1ID, &c.User2ID, &c.CreatedAt, &c.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}

	return &c, nil
}

func (r *PostgresConversationRepository) Create(c *domain.Conversation) error {
	query := `
		INSERT INTO conversations (id, user1_id, user2_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (user1_id, user2_id) DO UPDATE SET updated_at = EXCLUDED.updated_at
		RETURNING id, created_at, updated_at
	`

	// Ensure consistent ordering to match a potential UNIQUE index on (least(id), greatest(id))
	// Or if the UNIQUE index is just (user1_id, user2_id), we must ensure we always insert in a specific order?
	// The schema defined earlier was: UNIQUE(user1_id, user2_id).
	// To prevent duplicates like (A, B) and (B, A), we should probably sort them.

	u1, u2 := c.User1ID, c.User2ID
	if u1 > u2 {
		u1, u2 = u2, u1
	}
	// Update the struct to match the stored order
	c.User1ID = u1
	c.User2ID = u2

	err := r.pool.QueryRow(context.Background(), query,
		c.ID, c.User1ID, c.User2ID, c.CreatedAt, c.UpdatedAt,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create conversation: %w", err)
	}

	return nil
}

func (r *PostgresConversationRepository) GetUserConversations(userID string) ([]domain.Conversation, error) {
	query := `
		SELECT id, user1_id, user2_id, created_at, updated_at
		FROM conversations
		WHERE user1_id = $1 OR user2_id = $1
		ORDER BY updated_at DESC
	`

	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query conversations: %w", err)
	}
	defer rows.Close()

	var conversations []domain.Conversation
	for rows.Next() {
		var c domain.Conversation
		if err := rows.Scan(&c.ID, &c.User1ID, &c.User2ID, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan conversation: %w", err)
		}
		conversations = append(conversations, c)
	}

	return conversations, nil
}

// --- Direct Message Repository ---

type PostgresDirectMessageRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresDirectMessageRepository(pool *pgxpool.Pool) domain.DirectMessageRepository {
	return &PostgresDirectMessageRepository{pool: pool}
}

func (r *PostgresDirectMessageRepository) Save(msg *domain.DirectMessage) error {
	// 1. Calculate next sequence number for this conversation
	// Using a subquery to get MAX(sequence_number) + 1
	// If no messages exist, start at 1.

	// We also set initial status to 'SENT'.

	query := `
		INSERT INTO direct_messages (id, conversation_id, sender_id, content, type, created_at, sequence_number, status)
		VALUES (
			$1, $2, $3, $4, $5, $6,
			(SELECT COALESCE(MAX(sequence_number), 0) + 1 FROM direct_messages WHERE conversation_id = $2),
			'SENT'
		)
		RETURNING sequence_number
	`

	err := r.pool.QueryRow(context.Background(), query,
		msg.ID, msg.ConversationID, msg.SenderID, msg.Content, msg.Type, msg.CreatedAt,
	).Scan(&msg.Sequence)

	if err != nil {
		return fmt.Errorf("failed to save direct message: %w", err)
	}

	// Also update the conversation's updated_at timestamp
	updateQuery := `UPDATE conversations SET updated_at = $1 WHERE id = $2`
	_, _ = r.pool.Exec(context.Background(), updateQuery, msg.CreatedAt, msg.ConversationID)

	return nil
}

func (r *PostgresDirectMessageRepository) GetByConversationID(conversationID string, limit, offset int) ([]domain.DirectMessage, error) {
	query := `
		SELECT id, conversation_id, sender_id, content, type, sequence_number, created_at, status
		FROM direct_messages
		WHERE conversation_id = $1
		ORDER BY sequence_number DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(context.Background(), query, conversationID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query direct messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.DirectMessage
	for rows.Next() {
		var m domain.DirectMessage
		// Check if status is NULL (for old messages if any, though schema default helps)
		var status *string
		if err := rows.Scan(
			&m.ID, &m.ConversationID, &m.SenderID, &m.Content, &m.Type, &m.Sequence, &m.CreatedAt, &status,
		); err != nil {
			return nil, fmt.Errorf("failed to scan direct message: %w", err)
		}
		if status != nil {
			m.Status = *status
		} else {
			m.Status = "SENT" // Default fallback
		}
		messages = append(messages, m)
	}
	return messages, nil
}
