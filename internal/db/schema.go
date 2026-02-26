package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func InitSchema(pool *pgxpool.Pool) error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY,
			phone_number VARCHAR(20) UNIQUE NOT NULL,
			name VARCHAR(100),
			photo_url TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS groups (
			id UUID PRIMARY KEY,
			name VARCHAR(100) NOT NULL,
			photo_url TEXT,
			contribution_amount DECIMAL(15, 2) NOT NULL,
			rotation_frequency VARCHAR(50) NOT NULL,
			custom_frequency_days INT,
			creator_id UUID NOT NULL REFERENCES users(id),
			invite_link VARCHAR(255) UNIQUE,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS group_members (
			id UUID PRIMARY KEY,
			group_id UUID NOT NULL REFERENCES groups(id),
			user_id UUID NOT NULL REFERENCES users(id),
			role VARCHAR(20) NOT NULL,
			status VARCHAR(20) NOT NULL,
			joined_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(group_id, user_id)
		);`,
		`CREATE TABLE IF NOT EXISTS messages (
			id UUID PRIMARY KEY,
			group_id UUID NOT NULL REFERENCES groups(id),
			sender_id UUID REFERENCES users(id), -- Nullable for system messages
			content TEXT NOT NULL,
			type VARCHAR(50) NOT NULL,
			sequence_number BIGSERIAL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
		`CREATE TABLE IF NOT EXISTS rounds (
			id UUID PRIMARY KEY,
			group_id UUID NOT NULL REFERENCES groups(id),
			round_number INT NOT NULL,
			recipient_id UUID REFERENCES users(id),
			start_date TIMESTAMP WITH TIME ZONE NOT NULL,
			end_date TIMESTAMP WITH TIME ZONE NOT NULL,
			status VARCHAR(20) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(group_id, round_number)
		);`,
		`CREATE TABLE IF NOT EXISTS conversations (
			id UUID PRIMARY KEY,
			user1_id UUID NOT NULL REFERENCES users(id),
			user2_id UUID NOT NULL REFERENCES users(id),
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(user1_id, user2_id)
		);`,
		`CREATE TABLE IF NOT EXISTS direct_messages (
			id UUID PRIMARY KEY,
			conversation_id UUID NOT NULL REFERENCES conversations(id),
			sender_id UUID NOT NULL REFERENCES users(id),
			content TEXT NOT NULL,
			type VARCHAR(50) NOT NULL DEFAULT 'TEXT',
			sequence_number BIGSERIAL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);`,
	}

	for _, query := range queries {
		_, err := pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("failed to execute migration query: %s, error: %w", query, err)
		}
	}

	// Migrations for existing tables
	alterQueries := []string{
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS sequence_number BIGSERIAL;`,
		`ALTER TABLE messages ADD COLUMN IF NOT EXISTS sender_id UUID REFERENCES users(id);`, // Ensure sender_id exists and has correct reference
		`ALTER TABLE direct_messages ADD COLUMN IF NOT EXISTS sequence_number BIGINT DEFAULT 0;`,
		`ALTER TABLE direct_messages ADD COLUMN IF NOT EXISTS status VARCHAR(20) DEFAULT 'SENT';`,
	}

	for _, query := range alterQueries {
		_, err := pool.Exec(context.Background(), query)
		if err != nil {
			return fmt.Errorf("failed to execute alter query: %s, error: %w", query, err)
		}
	}

	return nil
}
