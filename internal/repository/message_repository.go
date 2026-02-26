package repository

import (
	"context"
	"fmt"
	"tuno_backend/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresMessageRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresMessageRepository(pool *pgxpool.Pool) domain.MessageRepository {
	return &PostgresMessageRepository{pool: pool}
}

func (r *PostgresMessageRepository) Save(message *domain.Message) error {
	query := `
		INSERT INTO messages (id, group_id, sender_id, content, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING sequence_number
	`
	err := r.pool.QueryRow(context.Background(), query,
		message.ID,
		message.GroupID,
		message.SenderID,
		message.Content,
		message.Type,
		message.CreatedAt,
	).Scan(&message.Sequence)

	if err != nil {
		return fmt.Errorf("failed to save message: %w", err)
	}
	return nil
}

func (r *PostgresMessageRepository) GetByGroupID(groupID string, limit, offset int) ([]domain.Message, error) {
	query := `
		SELECT id, group_id, sender_id, content, type, sequence_number, created_at
		FROM messages
		WHERE group_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(context.Background(), query, groupID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch messages: %w", err)
	}
	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var msg domain.Message
		err := rows.Scan(
			&msg.ID,
			&msg.GroupID,
			&msg.SenderID,
			&msg.Content,
			&msg.Type,
			&msg.Sequence,
			&msg.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan message: %w", err)
		}
		messages = append(messages, msg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return messages, nil
}
