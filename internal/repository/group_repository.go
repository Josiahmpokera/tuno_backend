package repository

import (
	"context"
	"fmt"
	"time"
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
		INSERT INTO groups (id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at, description, currency, is_started, start_date, total_rounds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
		group.Description,
		group.Currency,
		group.IsStarted,
		group.StartDate,
		group.TotalRounds,
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

func (r *PostgresGroupRepository) GetMember(groupID, userID string) (*domain.GroupMember, error) {
	query := `
		SELECT id, group_id, user_id, role, status, joined_at
		FROM group_members
		WHERE group_id = $1 AND user_id = $2
		LIMIT 1
	`
	row := r.pool.QueryRow(context.Background(), query, groupID, userID)

	var member domain.GroupMember
	err := row.Scan(
		&member.ID,
		&member.GroupID,
		&member.UserID,
		&member.Role,
		&member.Status,
		&member.JoinedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get group member: %w", err)
	}
	return &member, nil
}

func (r *PostgresGroupRepository) GetGroupMemberIDs(groupID string) ([]string, error) {
	query := `SELECT user_id FROM group_members WHERE group_id = $1 AND status = 'ACTIVE'`
	rows, err := r.pool.Query(context.Background(), query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group member IDs: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan member ID: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PostgresGroupRepository) FindByID(id string) (*domain.Group, error) {
	query := `
		SELECT id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at, description, currency, is_started, start_date, total_rounds
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
		&group.Description,
		&group.Currency,
		&group.IsStarted,
		&group.StartDate,
		&group.TotalRounds,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find group: %w", err)
	}
	return &group, nil
}

func (r *PostgresGroupRepository) GetGroupDetail(groupID, userID string) (*domain.GroupDetail, error) {
	ctx := context.Background()

	// 1. Get Group info + User Role + Member Count
	query := `
		SELECT 
			g.id, g.name, g.photo_url, g.contribution_amount, g.rotation_frequency, 
			g.custom_frequency_days, g.creator_id, g.invite_link, g.created_at, g.updated_at,
			g.description, g.currency, g.is_started, g.start_date, g.total_rounds,
			gm.role,
			(SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'ACTIVE') as members_count
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE g.id = $1 AND gm.user_id = $2
	`
	var gd domain.GroupDetail
	err := r.pool.QueryRow(ctx, query, groupID, userID).Scan(
		&gd.ID, &gd.Name, &gd.PhotoURL, &gd.ContributionAmount, &gd.RotationFrequency,
		&gd.CustomFrequencyDays, &gd.CreatorID, &gd.InviteLink, &gd.CreatedAt, &gd.UpdatedAt,
		&gd.Description, &gd.Currency, &gd.IsStarted, &gd.StartDate, &gd.TotalRounds,
		&gd.Role,
		&gd.MembersCount,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found or access denied")
		}
		return nil, fmt.Errorf("failed to get group detail: %w", err)
	}

	// 2. Get Current Round (if any)
	roundQuery := `
		SELECT id, group_id, round_number, recipient_id, start_date, end_date, status, created_at
		FROM rounds
		WHERE group_id = $1 AND status = 'ACTIVE'
		ORDER BY round_number ASC
		LIMIT 1
	`
	var round domain.Round
	err = r.pool.QueryRow(ctx, roundQuery, groupID).Scan(
		&round.ID, &round.GroupID, &round.RoundNumber, &round.RecipientID,
		&round.StartDate, &round.EndDate, &round.Status, &round.CreatedAt,
	)
	if err == nil {
		gd.CurrentRound = &round
	} else if err != pgx.ErrNoRows {
		// Log error but continue? Or fail?
		// For now, fail
		return nil, fmt.Errorf("failed to get current round: %w", err)
	}

	return &gd, nil
}

func (r *PostgresGroupRepository) GetGroupsByUserID(userID string) ([]domain.GroupWithRole, error) {
	query := `
		SELECT 
			g.id, g.name, g.photo_url, g.contribution_amount, g.rotation_frequency, 
			g.custom_frequency_days, g.creator_id, g.invite_link, g.created_at, g.updated_at,
			g.description, g.currency, g.is_started, g.start_date, g.total_rounds,
			gm.role,
			m.id, m.content, m.type, m.sender_id, m.created_at
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		LEFT JOIN LATERAL (
			SELECT id, content, type, sender_id, created_at
			FROM messages
			WHERE group_id = g.id
			ORDER BY created_at DESC
			LIMIT 1
		) m ON true
		WHERE gm.user_id = $1
		ORDER BY COALESCE(m.created_at, g.created_at) DESC
	`
	rows, err := r.pool.Query(context.Background(), query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user groups: %w", err)
	}
	defer rows.Close()

	var groups []domain.GroupWithRole
	for rows.Next() {
		var gr domain.GroupWithRole
		var msgID, msgContent, msgType, msgSenderID *string
		var msgCreatedAt *time.Time

		err := rows.Scan(
			&gr.ID, &gr.Name, &gr.PhotoURL, &gr.ContributionAmount, &gr.RotationFrequency,
			&gr.CustomFrequencyDays, &gr.CreatorID, &gr.InviteLink, &gr.CreatedAt, &gr.UpdatedAt,
			&gr.Description, &gr.Currency, &gr.IsStarted, &gr.StartDate, &gr.TotalRounds,
			&gr.Role,
			&msgID, &msgContent, &msgType, &msgSenderID, &msgCreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}

		if msgID != nil {
			gr.LastMessage = &domain.Message{
				ID:        *msgID,
				GroupID:   gr.ID,
				Content:   *msgContent,
				Type:      domain.MessageType(*msgType),
				SenderID:  msgSenderID,
				CreatedAt: *msgCreatedAt,
			}
		}

		groups = append(groups, gr)
	}
	return groups, nil
}

func (r *PostgresGroupRepository) CreateGroupWithMembers(group *domain.Group, members []domain.GroupMember, systemMessage *domain.Message) error {
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	// Insert Group
	_, err = tx.Exec(context.Background(), `
		INSERT INTO groups (id, name, photo_url, contribution_amount, rotation_frequency, custom_frequency_days, creator_id, invite_link, created_at, updated_at, description, currency, is_started, start_date, total_rounds)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`, group.ID, group.Name, group.PhotoURL, group.ContributionAmount, group.RotationFrequency, group.CustomFrequencyDays, group.CreatorID, group.InviteLink, group.CreatedAt, group.UpdatedAt, group.Description, group.Currency, group.IsStarted, group.StartDate, group.TotalRounds)
	if err != nil {
		return fmt.Errorf("failed to insert group: %w", err)
	}

	// Insert Members
	for _, member := range members {
		_, err = tx.Exec(context.Background(), `
			INSERT INTO group_members (id, group_id, user_id, role, status, joined_at)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, member.ID, member.GroupID, member.UserID, member.Role, member.Status, member.JoinedAt)
		if err != nil {
			return fmt.Errorf("failed to insert member: %w", err)
		}
	}

	// Insert System Message
	_, err = tx.Exec(context.Background(), `
		INSERT INTO messages (id, group_id, sender_id, content, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, systemMessage.ID, systemMessage.GroupID, systemMessage.SenderID, systemMessage.Content, systemMessage.Type, systemMessage.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert system message: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepository) AddMemberWithSystemMessage(member *domain.GroupMember, systemMessage *domain.Message) error {
	tx, err := r.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), `
		INSERT INTO group_members (id, group_id, user_id, role, status, joined_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, member.ID, member.GroupID, member.UserID, member.Role, member.Status, member.JoinedAt)
	if err != nil {
		return fmt.Errorf("failed to insert member: %w", err)
	}

	_, err = tx.Exec(context.Background(), `
		INSERT INTO messages (id, group_id, sender_id, content, type, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, systemMessage.ID, systemMessage.GroupID, systemMessage.SenderID, systemMessage.Content, systemMessage.Type, systemMessage.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert system message: %w", err)
	}

	if err := tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	return nil
}

func (r *PostgresGroupRepository) GetGroupMembers(groupID string) ([]domain.GroupMemberView, error) {
	ctx := context.Background()

	query := `
		SELECT u.id, u.name, u.photo_url, gm.role, gm.status, gm.joined_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND gm.status = 'ACTIVE'
		ORDER BY gm.joined_at ASC
	`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []domain.GroupMemberView
	for rows.Next() {
		var m domain.GroupMemberView
		if err := rows.Scan(&m.UserID, &m.Name, &m.PhotoURL, &m.Role, &m.Status, &m.JoinedAt); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}

	return members, nil
}

func (r *PostgresGroupRepository) GetGroupMembersPaginated(groupID string, limit, offset int) ([]domain.GroupMemberView, int, error) {
	return r.getGroupMembersWithPagination(groupID, limit, offset)
}

func (r *PostgresGroupRepository) getGroupMembersWithPagination(groupID string, limit, offset int) ([]domain.GroupMemberView, int, error) {
	ctx := context.Background()

	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND gm.status = 'ACTIVE'
	`
	var totalCount int
	err := r.pool.QueryRow(ctx, countQuery, groupID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get member count: %w", err)
	}

	// Get paginated members
	query := `
		SELECT u.id, u.name, u.photo_url, gm.role, gm.status, gm.joined_at
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		WHERE gm.group_id = $1 AND gm.status = 'ACTIVE'
		ORDER BY gm.joined_at ASC
	`

	// Add pagination if limit > 0
	if limit > 0 {
		query += ` LIMIT $2 OFFSET $3`
	}

	var rows pgx.Rows
	if limit > 0 {
		rows, err = r.pool.Query(ctx, query, groupID, limit, offset)
	} else {
		rows, err = r.pool.Query(ctx, query, groupID)
	}

	if err != nil {
		return nil, 0, fmt.Errorf("failed to get group members: %w", err)
	}
	defer rows.Close()

	var members []domain.GroupMemberView
	for rows.Next() {
		var m domain.GroupMemberView
		if err := rows.Scan(&m.UserID, &m.Name, &m.PhotoURL, &m.Role, &m.Status, &m.JoinedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}

	return members, totalCount, nil
}

func (r *PostgresGroupRepository) HasCommonActiveGroup(user1ID, user2ID string) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1
			FROM group_members gm1
			JOIN group_members gm2 ON gm1.group_id = gm2.group_id
			WHERE gm1.user_id = $1 AND gm2.user_id = $2
			AND gm1.status = 'ACTIVE' AND gm2.status = 'ACTIVE'
		)
	`
	var exists bool
	err := r.pool.QueryRow(context.Background(), query, user1ID, user2ID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check common group: %w", err)
	}
	return exists, nil
}

func (r *PostgresGroupRepository) GetGroupHome(groupID, userID string) (*domain.GroupHomeResponse, error) {
	ctx := context.Background()
	resp := &domain.GroupHomeResponse{}

	// 1. Get Group Details & User Role & Stats
	// We need to count total active members for stats and pot calculation
	groupQuery := `
		SELECT 
			g.id, g.name, g.photo_url, g.contribution_amount, g.rotation_frequency, 
			g.custom_frequency_days, g.creator_id, g.invite_link, g.created_at, g.updated_at,
			g.description, g.currency, g.is_started, g.start_date, g.total_rounds,
			gm.role,
			(SELECT COUNT(*) FROM group_members WHERE group_id = g.id AND status = 'ACTIVE') as total_members
		FROM groups g
		JOIN group_members gm ON g.id = gm.group_id
		WHERE g.id = $1 AND gm.user_id = $2
	`
	var totalMembers int
	err := r.pool.QueryRow(ctx, groupQuery, groupID, userID).Scan(
		&resp.Group.ID, &resp.Group.Name, &resp.Group.PhotoURL, &resp.Group.ContributionAmount, &resp.Group.RotationFrequency,
		&resp.Group.CustomFrequencyDays, &resp.Group.CreatorID, &resp.Group.InviteLink, &resp.Group.CreatedAt, &resp.Group.UpdatedAt,
		&resp.Group.Description, &resp.Group.Currency, &resp.Group.IsStarted, &resp.Group.StartDate, &resp.Group.TotalRounds,
		&resp.Group.Role,
		&totalMembers,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("group not found or user not member")
		}
		return nil, fmt.Errorf("failed to get group details: %w", err)
	}
	resp.Stats.TotalMembers = totalMembers

	// 2. Get Current Round
	// We prefer ACTIVE, then PENDING (next), then COMPLETED (last)
	roundQuery := `
		SELECT id, round_number, start_date, end_date, status, recipient_id
		FROM rounds
		WHERE group_id = $1
		ORDER BY 
			CASE status 
				WHEN 'ACTIVE' THEN 1 
				WHEN 'PENDING' THEN 2 
				WHEN 'COMPLETED' THEN 3 
				ELSE 4 
			END ASC, 
			round_number ASC
		LIMIT 1
	`
	var roundID string
	var roundRecipientID *string
	resp.CurrentRound = &domain.RoundDetail{}
	err = r.pool.QueryRow(ctx, roundQuery, groupID).Scan(
		&roundID,
		&resp.CurrentRound.RoundNumber,
		&resp.CurrentRound.StartDate,
		&resp.CurrentRound.EndDate,
		&resp.CurrentRound.Status,
		&roundRecipientID,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			resp.CurrentRound = nil
		} else {
			return nil, fmt.Errorf("failed to get current round: %w", err)
		}
	}

	if resp.CurrentRound != nil {
		resp.CurrentRound.PotAmount = float64(totalMembers) * resp.Group.ContributionAmount

		// Fetch Receiver details if assigned
		if roundRecipientID != nil {
			userQuery := `SELECT id, name, photo_url FROM users WHERE id = $1`
			var receiver domain.UserShort
			err := r.pool.QueryRow(ctx, userQuery, roundRecipientID).Scan(&receiver.UserID, &receiver.Name, &receiver.PhotoURL)
			if err == nil {
				resp.CurrentRound.Receiver = &receiver
			}
		}

		// Calculate Collected Amount
		collectedQuery := `SELECT COALESCE(SUM(amount), 0) FROM contributions WHERE round_id = $1 AND status = 'PAID'`
		err := r.pool.QueryRow(ctx, collectedQuery, roundID).Scan(&resp.CurrentRound.CollectedAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate collected amount: %w", err)
		}
	}

	// 3. Get Members and their status for the *current round* (if exists)
	// Even if no round exists, we list members
	membersQuery := `
		SELECT 
			u.id, u.name, u.photo_url, u.phone_number,
			gm.role, gm.joined_at,
			r.round_number as assigned_round,
			c.status as payment_status
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		LEFT JOIN rounds r ON r.recipient_id = u.id AND r.group_id = gm.group_id
		LEFT JOIN contributions c ON c.user_id = u.id AND c.round_id = $3
		WHERE gm.group_id = $1 AND gm.status = 'ACTIVE'
		ORDER BY r.round_number ASC NULLS LAST, gm.joined_at ASC
	`
	// If no round, pass nil or empty UUID for round_id binding, but postgres requires UUID
	// If roundID is empty, we can pass a dummy UUID or handle it in query.
	// Easier: if roundID is empty, pass a non-existent UUID
	queryRoundID := roundID
	if queryRoundID == "" {
		queryRoundID = "00000000-0000-0000-0000-000000000000"
	}

	rows, err := r.pool.Query(ctx, membersQuery, groupID, userID, queryRoundID) // Wait, $2 is userID not used in this query?
	// Ah, I copied $2 from previous.
	// Actually membersQuery needs groupID ($1) and roundID ($3). But I can just use $1 and $2.

	membersQuery = `
		SELECT 
			u.id, u.name, u.photo_url, u.phone_number,
			gm.role, gm.joined_at,
			r.round_number as assigned_round,
			COALESCE(c.status, 'UNPAID') as payment_status
		FROM group_members gm
		JOIN users u ON gm.user_id = u.id
		LEFT JOIN rounds r ON r.recipient_id = u.id AND r.group_id = gm.group_id
		LEFT JOIN contributions c ON c.user_id = u.id AND c.round_id = $2
		WHERE gm.group_id = $1 AND gm.status = 'ACTIVE'
		ORDER BY r.round_number ASC NULLS LAST, gm.joined_at ASC
	`

	rows, err = r.pool.Query(ctx, membersQuery, groupID, queryRoundID)
	if err != nil {
		return nil, fmt.Errorf("failed to get members: %w", err)
	}
	defer rows.Close()

	var members []domain.MemberRoundStatus
	paidCount := 0
	unpaidCount := 0

	for rows.Next() {
		var m domain.MemberRoundStatus
		var assignedRound *int
		var paymentStatus string

		err := rows.Scan(
			&m.UserID, &m.Name, &m.PhotoURL, &m.PhoneNumber,
			&m.Role, &m.JoinedAt,
			&assignedRound,
			&paymentStatus,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}

		if assignedRound != nil {
			m.RoundNumber = *assignedRound
		}

		m.PaymentStatus = domain.ContributionStatus(paymentStatus)

		// If user is receiver of current round, status is EXEMPT
		if roundRecipientID != nil && m.UserID == *roundRecipientID {
			m.PaymentStatus = domain.ContributionExempt
		}

		// Stats logic
		if m.PaymentStatus == domain.ContributionPaid {
			paidCount++
		} else if m.PaymentStatus == domain.ContributionUnpaid {
			unpaidCount++
		}

		members = append(members, m)
	}
	resp.Members = members
	resp.Stats.PaidMembersCount = paidCount
	resp.Stats.UnpaidMembersCount = unpaidCount
	if totalMembers > 0 {
		resp.Stats.CompletionPercentage = (float64(paidCount) / float64(totalMembers)) * 100
	}

	return resp, nil
}

func (r *PostgresGroupRepository) GetGroupSchedule(groupID string, filterStatus *domain.RoundStatus) (*domain.GroupScheduleResponse, error) {
	ctx := context.Background()

	// 1. Fetch Rounds with Receiver
	roundsQuery := `
		SELECT 
			r.id, r.round_number, r.start_date, r.end_date, r.status,
			rec.id, rec.name, rec.photo_url
		FROM rounds r
		LEFT JOIN users rec ON r.recipient_id = rec.id
		WHERE r.group_id = $1
	`
	args := []interface{}{groupID}
	if filterStatus != nil {
		roundsQuery += ` AND r.status = $2`
		args = append(args, *filterStatus)
	}
	roundsQuery += ` ORDER BY r.round_number ASC`

	rows, err := r.pool.Query(ctx, roundsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get rounds: %w", err)
	}
	defer rows.Close()

	var rounds []domain.RoundScheduleItem
	roundIDs := []string{}

	// Map roundID to index in rounds slice for easy member population
	roundMap := make(map[string]int)

	for rows.Next() {
		var rItem domain.RoundScheduleItem
		var rID string
		var recID, recName, recPhoto *string

		err := rows.Scan(
			&rID, &rItem.RoundNumber, &rItem.StartDate, &rItem.EndDate, &rItem.Status,
			&recID, &recName, &recPhoto,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan round: %w", err)
		}

		if recID != nil {
			rItem.Receiver = &domain.UserShort{
				UserID:   *recID,
				Name:     *recName,
				PhotoURL: *recPhoto,
			}
		}

		rounds = append(rounds, rItem)
		roundIDs = append(roundIDs, rID)
		roundMap[rID] = len(rounds) - 1
	}
	rows.Close()

	if len(rounds) == 0 {
		return &domain.GroupScheduleResponse{Rounds: []domain.RoundScheduleItem{}}, nil
	}

	// 2. Fetch Members and Payment Status for ALL rounds
	// This query joins rounds, members, and contributions
	// We want to list ALL members for EACH round
	// So we need Cross Join of Members * Rounds, then left join Contributions
	// But fetching that much data might be heavy.
	// Optimization: Only fetch members for specific roundIDs

	membersQuery := `
		SELECT 
			r.id as round_id,
			u.id, u.name, u.photo_url,
			COALESCE(c.status, 'UNPAID') as payment_status
		FROM rounds r
		CROSS JOIN group_members gm
		JOIN users u ON gm.user_id = u.id
		LEFT JOIN contributions c ON c.round_id = r.id AND c.user_id = u.id
		WHERE r.group_id = $1 AND gm.group_id = $1 AND gm.status = 'ACTIVE'
	`
	if filterStatus != nil {
		membersQuery += ` AND r.status = $2`
		// args already has status at index 1
	}
	membersQuery += ` ORDER BY r.round_number, u.name`

	// args is reused
	rows, err = r.pool.Query(ctx, membersQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get round members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var roundID string
		var m domain.MemberPaymentStatus
		var paymentStatus string

		err := rows.Scan(&roundID, &m.UserID, &m.Name, &m.PhotoURL, &paymentStatus)
		if err != nil {
			return nil, fmt.Errorf("failed to scan member status: %w", err)
		}

		m.PaymentStatus = domain.ContributionStatus(paymentStatus)

		// Find which round this belongs to
		idx, ok := roundMap[roundID]
		if ok {
			// Check if receiver
			if rounds[idx].Receiver != nil && rounds[idx].Receiver.UserID == m.UserID {
				m.IsReceiver = true
				m.PaymentStatus = domain.ContributionExempt
			}
			rounds[idx].Members = append(rounds[idx].Members, m)
		}
	}

	return &domain.GroupScheduleResponse{Rounds: rounds}, nil
}
