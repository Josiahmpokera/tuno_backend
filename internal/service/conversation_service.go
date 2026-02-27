package service

import (
	"context"
	"fmt"
	"time"
	"tuno_backend/internal/domain"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// --- Commands ---

type StartConversationCommand struct {
	CommandID string
	User1ID   string
	User2ID   string
}

func (c StartConversationCommand) CommandName() string {
	return "StartConversation"
}

type SendDirectMessageCommand struct {
	CommandID      string
	SenderID       string
	ConversationID string
	Content        string
	Type           domain.MessageType
}

func (c SendDirectMessageCommand) CommandName() string {
	return "SendDirectMessage"
}

type MarkMessagesReadCommand struct {
	CommandID      string
	UserID         string
	ConversationID string
}

func (c MarkMessagesReadCommand) CommandName() string {
	return "MarkMessagesRead"
}

// --- Events ---

type ConversationStartedEvent struct {
	ConversationID string
	User1ID        string
	User2ID        string
	CreatedAt      time.Time
}

func (e ConversationStartedEvent) EventName() string {
	return "ConversationStartedEvent"
}

type DirectMessageSentEvent struct {
	MessageID      string
	ConversationID string
	SenderID       string
	Content        string
	Type           domain.MessageType
	RecipientID    string
	CreatedAt      time.Time
}

func (e DirectMessageSentEvent) EventName() string {
	return "DirectMessageSentEvent"
}

type MessagesReadEvent struct {
	ConversationID string
	ReaderID       string // The user who read the messages
	PartnerID      string // The user whose messages were read (to be notified)
	ReadAt         time.Time
}

func (e MessagesReadEvent) EventName() string {
	return "MessagesReadEvent"
}

// --- Command Handlers ---

type StartConversationHandler struct {
	repo        domain.ConversationRepository
	groupRepo   domain.GroupRepository
	eventBus    EventBus
	redisClient *redis.Client
}

func NewStartConversationHandler(
	repo domain.ConversationRepository,
	groupRepo domain.GroupRepository,
	eventBus EventBus,
	redisClient *redis.Client,
) *StartConversationHandler {
	return &StartConversationHandler{
		repo:        repo,
		groupRepo:   groupRepo,
		eventBus:    eventBus,
		redisClient: redisClient,
	}
}

func (h *StartConversationHandler) Handle(ctx context.Context, cmd Command) (interface{}, error) {
	c, ok := cmd.(StartConversationCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Idempotency Check
	idempotencyKey := fmt.Sprintf("cmd:%s", c.CommandID)
	set, err := h.redisClient.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if !set {
		return nil, fmt.Errorf("duplicate command")
	}

	// 2. Validate Users
	if c.User1ID == c.User2ID {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("cannot start conversation with self")
	}

	// 3. Check for Shared Active Group (WhatsApp Style)
	hasCommon, err := h.groupRepo.HasCommonActiveGroup(c.User1ID, c.User2ID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to check common groups: %w", err)
	}
	if !hasCommon {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("users do not share a common active group")
	}

	// 4. Check if Conversation Exists
	existing, err := h.repo.FindByUsers(c.User1ID, c.User2ID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to check existing conversation: %w", err)
	}
	if existing != nil {
		// Return existing ID
		return existing.ID, nil
	}

	// 5. Create Event
	conversationID := uuid.New().String()
	event := ConversationStartedEvent{
		ConversationID: conversationID,
		User1ID:        c.User1ID,
		User2ID:        c.User2ID,
		CreatedAt:      time.Now(),
	}

	// 6. Dispatch Event (Synchronous)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to publish conversation started event: %w", err)
	}

	return conversationID, nil
}

type SendDirectMessageHandler struct {
	convRepo    domain.ConversationRepository
	dmRepo      domain.DirectMessageRepository
	groupRepo   domain.GroupRepository
	eventBus    EventBus
	redisClient *redis.Client
}

func NewSendDirectMessageHandler(
	convRepo domain.ConversationRepository,
	dmRepo domain.DirectMessageRepository,
	groupRepo domain.GroupRepository,
	eventBus EventBus,
	redisClient *redis.Client,
) *SendDirectMessageHandler {
	return &SendDirectMessageHandler{
		convRepo:    convRepo,
		dmRepo:      dmRepo,
		groupRepo:   groupRepo,
		eventBus:    eventBus,
		redisClient: redisClient,
	}
}

func (h *SendDirectMessageHandler) Handle(ctx context.Context, cmd Command) (interface{}, error) {
	c, ok := cmd.(SendDirectMessageCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Idempotency Check
	idempotencyKey := fmt.Sprintf("cmd:%s", c.CommandID)
	set, err := h.redisClient.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if !set {
		return nil, fmt.Errorf("duplicate command")
	}

	// 2. Validate Conversation Exists and Users Share Group
	conversation, err := h.convRepo.FindByID(c.ConversationID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to find conversation: %w", err)
	}
	if conversation == nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("conversation not found")
	}

	// Determine Recipient ID
	var recipientID string
	if conversation.User1ID == c.SenderID {
		recipientID = conversation.User2ID
	} else if conversation.User2ID == c.SenderID {
		recipientID = conversation.User1ID
	} else {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("sender is not a participant in this conversation")
	}

	// Check Shared Active Group (Recheck Rule)
	hasCommon, err := h.groupRepo.HasCommonActiveGroup(c.SenderID, recipientID)
	if err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to check common groups: %w", err)
	}
	if !hasCommon {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("cannot send message: no common active group shared with recipient")
	}

	// 3. Create Event
	event := DirectMessageSentEvent{
		MessageID:      uuid.New().String(),
		ConversationID: c.ConversationID,
		SenderID:       c.SenderID,
		RecipientID:    recipientID,
		Content:        c.Content,
		Type:           c.Type,
		CreatedAt:      time.Now(),
	}

	// 4. Dispatch Event (Synchronous)
	if err := h.eventBus.Publish(ctx, event); err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to publish direct message sent event: %w", err)
	}

	return nil, nil
}

type MarkMessagesReadHandler struct {
	repo        domain.DirectMessageRepository
	convRepo    domain.ConversationRepository
	eventBus    EventBus
	redisClient *redis.Client
}

func NewMarkMessagesReadHandler(
	repo domain.DirectMessageRepository,
	convRepo domain.ConversationRepository,
	eventBus EventBus,
	redisClient *redis.Client,
) *MarkMessagesReadHandler {
	return &MarkMessagesReadHandler{
		repo:        repo,
		convRepo:    convRepo,
		eventBus:    eventBus,
		redisClient: redisClient,
	}
}

func (h *MarkMessagesReadHandler) Handle(ctx context.Context, cmd Command) (interface{}, error) {
	c, ok := cmd.(MarkMessagesReadCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Idempotency Check
	idempotencyKey := fmt.Sprintf("cmd:%s", c.CommandID)
	set, err := h.redisClient.SetNX(ctx, idempotencyKey, "processing", 24*time.Hour).Result()
	if err != nil {
		return nil, fmt.Errorf("idempotency check failed: %w", err)
	}
	if !set {
		return nil, fmt.Errorf("duplicate command")
	}

	// 2. Mark as Read
	if err := h.repo.MarkMessagesAsRead(c.ConversationID, c.UserID); err != nil {
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to mark messages as read: %w", err)
	}

	// 3. Find Partner ID (to notify them)
	conv, err := h.convRepo.FindByID(c.ConversationID)
	if err != nil {
		// Log error but continue? No, if we can't find conversation, something is wrong.
		// But messages are already marked.
		// Let's just log and skip notification.
		// For now, return error to be safe, though partial failure is bad.
		// But idempotency protects retry.
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to find conversation for notification: %w", err)
	}

	var partnerID string
	if conv.User1ID == c.UserID {
		partnerID = conv.User2ID
	} else {
		partnerID = conv.User1ID
	}

	// 4. Create Event
	event := MessagesReadEvent{
		ConversationID: c.ConversationID,
		ReaderID:       c.UserID,
		PartnerID:      partnerID,
		ReadAt:         time.Now(),
	}

	// 5. Publish Event
	if err := h.eventBus.Publish(ctx, event); err != nil {
		// Event failed, but DB updated. This is inconsistent.
		// But we can't rollback DB easily here without transaction manager across services.
		// For now, we rely on retry or just log.
		// We delete idempotency key so client can retry.
		h.redisClient.Del(ctx, idempotencyKey)
		return nil, fmt.Errorf("failed to publish messages read event: %w", err)
	}

	return nil, nil
}

// --- Event Handlers (Projections) ---

type ConversationStartedEventHandler struct {
	repo domain.ConversationRepository
}

func NewConversationStartedEventHandler(repo domain.ConversationRepository) *ConversationStartedEventHandler {
	return &ConversationStartedEventHandler{repo: repo}
}

func (h *ConversationStartedEventHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(ConversationStartedEvent)
	if !ok {
		return nil // Ignore other events
	}

	conversation := &domain.Conversation{
		ID:        e.ConversationID,
		User1ID:   e.User1ID,
		User2ID:   e.User2ID,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.CreatedAt,
	}

	return h.repo.Create(conversation)
}

type DirectMessageSentEventHandler struct {
	repo domain.DirectMessageRepository
}

func NewDirectMessageSentEventHandler(repo domain.DirectMessageRepository) *DirectMessageSentEventHandler {
	return &DirectMessageSentEventHandler{repo: repo}
}

func (h *DirectMessageSentEventHandler) Handle(ctx context.Context, event domain.Event) error {
	e, ok := event.(DirectMessageSentEvent)
	if !ok {
		return nil
	}

	msg := &domain.DirectMessage{
		ID:             e.MessageID,
		ConversationID: e.ConversationID,
		SenderID:       e.SenderID,
		Content:        e.Content,
		Type:           e.Type,
		CreatedAt:      e.CreatedAt,
	}

	return h.repo.Save(msg)
}

// --- Query Handlers ---

type GetConversationsHandler struct {
	repo domain.ConversationRepository
}

func NewGetConversationsHandler(repo domain.ConversationRepository) *GetConversationsHandler {
	return &GetConversationsHandler{repo: repo}
}

func (h *GetConversationsHandler) Handle(userID string) ([]domain.Conversation, error) {
	return h.repo.GetUserConversations(userID)
}

type GetDirectMessagesHandler struct {
	repo domain.DirectMessageRepository
}

func NewGetDirectMessagesHandler(repo domain.DirectMessageRepository) *GetDirectMessagesHandler {
	return &GetDirectMessagesHandler{repo: repo}
}

func (h *GetDirectMessagesHandler) Handle(conversationID string, limit, offset int) ([]domain.DirectMessage, error) {
	return h.repo.GetByConversationID(conversationID, limit, offset)
}
