package service

import (
	"context"
	"fmt"
	"time"
	"tuno_backend/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Commands

type CreateGroupCommand struct {
	CommandID           string
	Name                string
	PhotoURL            string
	ContributionAmount  float64
	RotationFrequency   domain.RotationFrequency
	CustomFrequencyDays int
	CreatorID           string
	Invitees            []string // List of phone numbers
	Timestamp           int64    // Unix timestamp
}

func (c CreateGroupCommand) CommandName() string {
	return "CreateGroupCommand"
}

type AddGroupMemberByPhoneCommand struct {
	CommandID    string
	GroupID      string
	PhoneNumber  string
	AddedByID    string
	Timestamp    int64
}

func (c AddGroupMemberByPhoneCommand) CommandName() string {
	return "AddGroupMemberByPhoneCommand"
}

// Command Handler

type CreateGroupHandler struct {
	eventBus    EventBus
	userRepo    domain.UserRepository
	redisClient *redis.Client
}

func NewCreateGroupHandler(eventBus EventBus, userRepo domain.UserRepository, redisClient *redis.Client) *CreateGroupHandler {
	return &CreateGroupHandler{eventBus: eventBus, userRepo: userRepo, redisClient: redisClient}
}

func (h *CreateGroupHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(CreateGroupCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 0. Idempotency Check (Atomic SETNX)
	idempotencyKey := fmt.Sprintf("cmd:%s", c.CommandID)
	// Try to acquire lock/mark as processing
	success, err := h.redisClient.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !success {
		return nil, fmt.Errorf("duplicate command: request already processed or in progress")
	}

	// 1. Validation (Command Side)
	if c.Name == "" {
		return nil, fmt.Errorf("group name is required")
	}
	if c.ContributionAmount <= 0 {
		return nil, fmt.Errorf("contribution amount must be positive")
	}
	if c.CreatorID == "" {
		return nil, fmt.Errorf("creator id is required")
	}

	// 2. Generate Group ID (Server Only)
	groupID := uuid.New().String()
	// Generate a simple invite link (in production, use a more robust short-link service)
	inviteLink := fmt.Sprintf("https://tuno.app/join/%s", uuid.New().String()[:8])

	// 3. Prepare Members List (In-Memory, No DB Write yet)
	var members []domain.GroupMember

	// Add Creator as Admin
	creatorMember := domain.GroupMember{
		ID:       uuid.New().String(),
		GroupID:  groupID,
		UserID:   c.CreatorID,
		Role:     domain.RoleAdmin,
		Status:   domain.StatusActive, // Creator is automatically active
		JoinedAt: time.Now(),
	}
	members = append(members, creatorMember)

	// Handle Invitees
	invitedUsers := []string{}
	failedInvitees := []string{}

	for _, phone := range c.Invitees {
		// Check if user exists
		user, err := h.userRepo.FindByPhoneNumber(phone)
		if err != nil {
			// Log error but continue
			failedInvitees = append(failedInvitees, phone)
			continue
		}
		if user == nil {
			// User not found in DB
			failedInvitees = append(failedInvitees, phone)
			continue
		}

		// Add as Member (Invited)
		member := domain.GroupMember{
			ID:       uuid.New().String(),
			GroupID:  groupID,
			UserID:   user.ID,
			Role:     domain.RoleMember,
			Status:   domain.StatusInvited,
			JoinedAt: time.Now(),
		}
		members = append(members, member)
		invitedUsers = append(invitedUsers, phone)
	}

	// 4. Create Immutable Event (Source of Truth)
	event := domain.GroupCreatedEvent{
		GroupID:             groupID,
		Name:                c.Name,
		PhotoURL:            c.PhotoURL,
		ContributionAmount:  c.ContributionAmount,
		RotationFrequency:   c.RotationFrequency,
		CustomFrequencyDays: c.CustomFrequencyDays,
		CreatorID:           c.CreatorID,
		InviteLink:          inviteLink,
		Members:             members,
		CreatedAt:           time.Now(),
	}

	// 5. Emit Event (Sync or Async)
	// This event handler will perform the DB writes in a transaction
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// If publishing fails, we should ideally delete the idempotency key so client can retry.
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to publish group created event: %w", err)
	}

	// 6. Return Result (ACK only after event is committed/processed)
	return map[string]interface{}{
		"group_id":        groupID,
		"invite_link":     inviteLink,
		"invited_count":   len(invitedUsers),
		"invited_users":   invitedUsers,
		"failed_invitees": failedInvitees,
		"message":         "Group created successfully",
	}, nil
}

type AddGroupMemberByPhoneHandler struct {
	eventBus    EventBus
	userRepo    domain.UserRepository
	groupRepo   domain.GroupRepository
	redisClient *redis.Client
}

func NewAddGroupMemberByPhoneHandler(
	eventBus EventBus,
	userRepo domain.UserRepository,
	groupRepo domain.GroupRepository,
	redisClient *redis.Client,
) *AddGroupMemberByPhoneHandler {
	return &AddGroupMemberByPhoneHandler{
		eventBus:    eventBus,
		userRepo:    userRepo,
		groupRepo:   groupRepo,
		redisClient: redisClient,
	}
}

func (h *AddGroupMemberByPhoneHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(AddGroupMemberByPhoneCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	idempotencyKey := fmt.Sprintf("cmd:%s", c.CommandID)
	success, err := h.redisClient.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to check idempotency: %w", err)
	}
	if !success {
		return nil, fmt.Errorf("duplicate command: request already processed or in progress")
	}

	if c.GroupID == "" || c.PhoneNumber == "" || c.AddedByID == "" {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("group_id, phone_number, and added_by_id are required")
	}

	addedByMember, err := h.groupRepo.GetMember(c.GroupID, c.AddedByID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to verify admin membership: %w", err)
	}
	if addedByMember == nil || addedByMember.Status != domain.StatusActive {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("forbidden: not an active group member")
	}
	if addedByMember.Role != domain.RoleAdmin {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("forbidden: admin role required")
	}

	user, err := h.userRepo.FindByPhoneNumber(c.PhoneNumber)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to check phone number availability: %w", err)
	}
	if user == nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("phone number is not registered on Tuno")
	}

	existing, err := h.groupRepo.GetMember(c.GroupID, user.ID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to check existing membership: %w", err)
	}
	if existing != nil && existing.Status != domain.StatusLeft {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("user already in group")
	}

	now := time.Now()
	event := domain.GroupMemberAddedEvent{
		GroupID:    c.GroupID,
		MemberID:   uuid.New().String(),
		UserID:     user.ID,
		AddedByID:  c.AddedByID,
		Role:       domain.RoleMember,
		Status:     domain.StatusActive,
		JoinedAt:   now,
		OccurredAt: now,
	}

	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to publish group member added event: %w", err)
	}

	return map[string]interface{}{
		"group_id":     c.GroupID,
		"user_id":      user.ID,
		"phone_number": user.PhoneNumber,
		"status":       string(domain.StatusActive),
		"message":      "Member added successfully",
	}, nil
}

// Event Handler (Projection)

type GroupCreatedEventHandler struct {
	groupRepo domain.GroupRepository
}

func NewGroupCreatedEventHandler(groupRepo domain.GroupRepository) *GroupCreatedEventHandler {
	return &GroupCreatedEventHandler{groupRepo: groupRepo}
}

func (h *GroupCreatedEventHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.GroupCreatedEvent)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	// Construct Group Domain Object
	group := &domain.Group{
		ID:                  e.GroupID,
		Name:                e.Name,
		PhotoURL:            e.PhotoURL,
		ContributionAmount:  e.ContributionAmount,
		RotationFrequency:   e.RotationFrequency,
		CustomFrequencyDays: e.CustomFrequencyDays,
		CreatorID:           e.CreatorID,
		InviteLink:          e.InviteLink,
		CreatedAt:           e.CreatedAt,
		UpdatedAt:           time.Now(), // Or same as CreatedAt
	}

	// Create System Message
	systemMessage := &domain.Message{
		ID:        uuid.New().String(),
		GroupID:   e.GroupID,
		SenderID:  nil, // System message
		Content:   "User created the group",
		Type:      domain.MessageTypeSystem,
		CreatedAt: time.Now(),
	}

	// Perform DB Write (Projection) in a single transaction
	if err := h.groupRepo.CreateGroupWithMembers(group, e.Members, systemMessage); err != nil {
		return fmt.Errorf("failed to project group state: %w", err)
	}

	return nil
}

type GroupMemberAddedEventHandler struct {
	groupRepo domain.GroupRepository
}

func NewGroupMemberAddedEventHandler(groupRepo domain.GroupRepository) *GroupMemberAddedEventHandler {
	return &GroupMemberAddedEventHandler{groupRepo: groupRepo}
}

func (h *GroupMemberAddedEventHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(domain.GroupMemberAddedEvent)
	if !ok {
		return fmt.Errorf("invalid event type")
	}

	member := &domain.GroupMember{
		ID:       e.MemberID,
		GroupID:  e.GroupID,
		UserID:   e.UserID,
		Role:     e.Role,
		Status:   e.Status,
		JoinedAt: e.JoinedAt,
	}

	systemMessage := &domain.Message{
		ID:        uuid.New().String(),
		GroupID:   e.GroupID,
		SenderID:  nil,
		Content:   "A new member was added",
		Type:      domain.MessageTypeSystem,
		CreatedAt: time.Now(),
	}

	if err := h.groupRepo.AddMemberWithSystemMessage(member, systemMessage); err != nil {
		return fmt.Errorf("failed to project group member added state: %w", err)
	}
	return nil
}
