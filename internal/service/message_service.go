package service

import (
	"context"
	"fmt"
	"time"
	"tuno_backend/internal/domain"

	"github.com/google/uuid"
)

type SendMessageCommand struct {
	UserID  string
	GroupID string
	Content string
	Type    domain.MessageType
}

func (c SendMessageCommand) CommandName() string {
	return "SendMessageCommand"
}

type SendMessageResult struct {
	Message   domain.Message
	MemberIDs []string
}

type MessageService struct {
	groupRepo   domain.GroupRepository
	messageRepo domain.MessageRepository
}

func NewMessageService(groupRepo domain.GroupRepository, messageRepo domain.MessageRepository) *MessageService {
	return &MessageService{
		groupRepo:   groupRepo,
		messageRepo: messageRepo,
	}
}

func (s *MessageService) SendMessage(ctx context.Context, cmd interface{}) (interface{}, error) {
	command, ok := cmd.(SendMessageCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Validate input
	if command.GroupID == "" || command.Content == "" {
		return nil, fmt.Errorf("group_id and content are required")
	}

	// 2. Verify membership
	// Check if user is a member of the group.
	// We can use GetGroupDetail which checks membership implicitly.
	// Or we can just try to get members and check if user is in list, but that's inefficient.
	// Let's use GetGroupDetail for now as it validates membership.
	_, err := s.groupRepo.GetGroupDetail(command.GroupID, command.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify membership: %w", err)
	}

	// 3. Create Message
	msg := domain.Message{
		ID:        uuid.New().String(),
		GroupID:   command.GroupID,
		SenderID:  &command.UserID,
		Content:   command.Content,
		Type:      command.Type,
		CreatedAt: time.Now(),
	}

	// 4. Persist Message
	// This will assign sequence number
	err = s.messageRepo.Save(&msg)
	if err != nil {
		return nil, fmt.Errorf("failed to save message: %w", err)
	}

	// 5. Get Group Members for fan-out
	memberIDs, err := s.groupRepo.GetGroupMemberIDs(command.GroupID)
	if err != nil {
		return nil, fmt.Errorf("failed to get group members: %w", err)
	}

	return SendMessageResult{
		Message:   msg,
		MemberIDs: memberIDs,
	}, nil
}
