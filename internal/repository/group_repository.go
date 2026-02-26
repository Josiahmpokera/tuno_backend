package repository

import (
	"context"
	"fmt"
	"tuno_backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresGroupRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresGroupRepository(pool *pgxpool.Pool) domain.GroupRepository {
	return &PostgresGroupRepository{pool: pool}
}

func (r *PostgresGroupRepository) Create(group *domain.Group) error {
	query := `
		INSERT INTO groups (id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err := r.pool.Exec(context.Background(), query,
		group.ID,
		group.Name,
		group.PhotoURL,
		group.ContributionAmount,
		group.RotationFrequency,
		group.CustomFrequencyDays,
		group.CreatorID,
		group.InviteLink,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepository) AddMember(member *domain.GroupMember) error {
	query := `
		INSERT INTO group_members (id, group_id, user_id, role, status, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(context.Background(), query,
		member.ID,
		member.GroupID,
		member.UserID,
		member.Role,
		member.Status,
		member.JoinedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepository) FindByID(id string) (*domain.Group, error) {
	query := `
		SELECT id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at
		FROM groups
		WHERE id = $1
	`
	row := r.pool.QueryRow(context.Background(), query, id)

	var group domain.Group
	err := row.Scan(
		&group.ID,
		&group.Name,
		&group.PhotoURL,
		&group.ContributionAmount,
		&group.RotationFrequency,
		&group.CustomFrequencyDays,
		&group.CreatorID,
		&group.InviteLink,
		&group.CreatedAt,
		&group.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find group by id: %w", err)
	}
	return &group, nil
}

func (r *PostgresGroupRepository) CreateGroupWithMembers(group *domain.Group, members []domain.GroupMember, initialMessage *domain.Message) error {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// 1. Insert Group
	groupQuery := `
		INSERT INTO groups (id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	`
	_, err = tx.Exec(ctx, groupQuery,
		group.ID,
		group.Name,
		group.PhotoURL,
		group.ContributionAmount,
		group.RotationFrequency,
		group.CustomFrequencyDays,
		group.CreatorID,
		group.InviteLink,
		group.CreatedAt,
		group.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create group: %w", err)
	}

	// 2. Insert Members
	memberQuery := `
		INSERT INTO group_members (id, group_id, user_id, role, status, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	for _, member := range members {
		_, err = tx.Exec(ctx, memberQuery,
			member.ID,
			member.GroupID,
			member.UserID,
			member.Role,
			member.Status,
			member.JoinedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to add member: %w", err)
		}
	}

	// 3. Insert System Message (if provided)
	if initialMessage != nil {
		messageQuery := `
			INSERT INTO messages (id, group_id, sender_id, content, type, created_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`
		_, err = tx.Exec(ctx, messageQuery,
			initialMessage.ID,
			initialMessage.GroupID,
			initialMessage.SenderID,
			initialMessage.Content,
			initialMessage.Type,
			initialMessage.CreatedAt,
		)
		if err != nil {
			return fmt.Errorf("failed to insert system message: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *PostgresGroupRepository) GetGroupsByUserID(userID string) ([]domain.GroupWithRole, error) {
	query := `
		SELECT 
			g.id, g.name, g.photo_url, g.contribution_amount, g.rotation_frequency, 
			g.custom_frequency_days, g.creator_id, g.invite_link, g.created_at, g.updated_at,
			gm.role
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE gm.user_id = $1
		ORDER BY g.created_at DESC
	`
	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.GroupWithRole
	for rows.Next() {
		var gr domain.GroupWithRole
		err := rows.Scan(
			&gr.ID,
			&gr.Name,
			&gr.PhotoURL,
			&gr.ContributionAmount,
			&gr.RotationFrequency,
			&gr.CustomFrequencyDays,
			&gr.CreatorID,
			&gr.InviteLink,
			&gr.CreatedAt,
			&gr.UpdatedAt,
			&gr.Role,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group row: %w", err)
		}
		groups = append(groups, gr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return groups, nil
}

func (r *PostgresGroupRepository) GetGroupDetail(groupID, userID string) (*domain.GroupDetail, error) {
	ctx := context.Background()

	// 1. Get Group info and verify membership
	query := `
		SELECT 
			g.id, g.name, g.photo_url, g.contribution_amount, g.rotation_frequency, 
			g.custom_frequency_days, g.creator_id, g.invite_link, g.created_at, g.updated_at,
			gm.role,
			(SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'ACTIVE') as member_count
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE g.id = $1 AND gm.user_id = $2
	`
	row := r.pool.QueryRow(ctx, query, groupID, userID)

	var detail domain.GroupDetail
	err := row.Scan(
		&detail.ID,
		&detail.Name,
		&detail.PhotoURL,
		&detail.ContributionAmount,
		&detail.RotationFrequency,
		&detail.CustomFrequencyDays,
		&detail.CreatorID,
		&detail.InviteLink,
		&detail.CreatedAt,
		&detail.UpdatedAt,
		&detail.Role,
		&detail.MembersCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Or specific error for "Not Found" or "Not Member"
		}
		return nil, fmt.Errorf("failed to fetch group details: %w", err)
	}

	// 2. Get Current Round (Active or Pending)
	roundQuery := `
		SELECT id, group_id, round_number, recipient_id, start_date, end_date, status, created_at
		FROM rounds
		WHERE group_id = $1 AND status IN ('PENDING', 'ACTIVE')
		ORDER BY round_number ASC
		LIMIT 1
	`
	roundRow := r.pool.QueryRow(ctx, roundQuery, groupID)

	var round domain.Round
	err = roundRow.Scan(
		&round.ID,
		&round.GroupID,
		&round.RoundNumber,
		&round.RecipientID,
		&round.StartDate,
		&round.EndDate,
		&round.Status,
		&round.CreatedAt,
	)
	if err != nil {
		if err != pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to fetch current round: %w", err)
		}
		// No active/pending round found, current_round remains nil
	} else {
		detail.CurrentRound = &round
	}

	return &detail, nil
}

func (r *PostgresGroupRepository) GetGroupMemberIDs(groupID string) ([]string, error) {
	query := `
		SELECT user_id
		FROM group_members
		WHERE group_id = $1 AND status = 'ACTIVE'
	`
	rows, err := r.pool.Query(context.Background(), query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}
	defer rows.Close()

	var memberIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan member id: %w", err)
		}
		memberIDs = append(memberIDs, userID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return memberIDs, nil
}

func (r *PostgresGroupRepository) HasCommonActiveGroup(user1ID, user2ID string) (bool, error) {
	// Find group where:
	// - both users are members
	// - neither has left (status != 'LEFT')
	// - For now, we allow INVITED status too, as long as they are in the group structure.
	// - If you want stricter rules (must be ACTIVE), change to status = 'ACTIVE'.
	// WhatsApp allows messaging if you share a group, even if one is just added.
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_members gm1
			JOIN group_members gm2 ON gm1.group_id = gm2.group_id
			WHERE gm1.user_id = $1 AND gm1.status != 'LEFT'
			  AND gm2.user_id = $2 AND gm2.status != 'LEFT'
		)
	`
	var exists bool
	err := r.pool.QueryRow(context.Background(), query, user1ID, user2ID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check common group: %w", err)
	}
	return exists, nil
}

func (r *PostgresGroupRepository) GetGroupMembers(groupID string) ([]domain.GroupMemberView, error) {
	query := `
		SELECT 
			gm.user_id,
			u.name,
			u.photo_url,
			gm.role,
			gm.status,
			gm.joined_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND gm.status != 'LEFT'
	`
	rows, err := r.pool.Query(context.Background(), query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch group members: %w", err)
	}
	defer rows.Close()

	var members []domain.GroupMemberView
	for rows.Next() {
		var m domain.GroupMemberView
		if err := rows.Scan(
			&m.UserID,
			&m.Name,
			&m.PhotoURL,
			&m.Role,
			&m.Status,
			&m.JoinedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}
	return members, nil
}
