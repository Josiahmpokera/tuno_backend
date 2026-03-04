package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateIndexes creates all necessary database indexes for performance optimization
func CreateIndexes(pool *pgxpool.Pool) error {
	ctx := context.Background()
	
	indexQueries := []string{
		// Indexes for group_members table
		`CREATE INDEX IF NOT EXISTS idx_group_members_group_id ON group_members(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_group_members_user_id ON group_members(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_group_members_group_user ON group_members(group_id, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_group_members_status ON group_members(status)`,
		
		// Indexes for groups table
		`CREATE INDEX IF NOT EXISTS idx_groups_creator_id ON groups(creator_id)`,
		
		// Indexes for rounds table
		`CREATE INDEX IF NOT EXISTS idx_rounds_group_id ON rounds(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_status ON rounds(status)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_recipient_id ON rounds(recipient_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_group_status ON rounds(group_id, status)`,
		
		// Indexes for contributions table
		`CREATE INDEX IF NOT EXISTS idx_contributions_round_id ON contributions(round_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contributions_user_id ON contributions(user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_contributions_status ON contributions(status)`,
		`CREATE INDEX IF NOT EXISTS idx_contributions_round_user ON contributions(round_id, user_id)`,
		
		// Indexes for messages table
		`CREATE INDEX IF NOT EXISTS idx_messages_group_id ON messages(group_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages(sender_id)`,
		`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages(created_at)`,
		
		// Composite indexes for frequently joined queries
		`CREATE INDEX IF NOT EXISTS idx_group_members_composite ON group_members(group_id, status, user_id)`,
		`CREATE INDEX IF NOT EXISTS idx_rounds_composite ON rounds(group_id, status, round_number)`,
	}

	for _, query := range indexQueries {
		_, err := pool.Exec(ctx, query)
		if err != nil {
			return fmt.Errorf("failed to create index: %s, error: %w", query, err)
		}
	}

	return nil
}