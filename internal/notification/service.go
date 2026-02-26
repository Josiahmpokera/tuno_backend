package notification

import (
	"context"
	"encoding/json"

	"tuno_backend/internal/domain"
	"tuno_backend/internal/service"
	"tuno_backend/internal/websocket"
	"tuno_backend/pkg/logger"

	"go.uber.org/zap"
)

// NotificationService handles real-time notifications via WebSocket
type NotificationService struct {
	hub *websocket.Hub
}

func NewNotificationService(hub *websocket.Hub) *NotificationService {
	return &NotificationService{
		hub: hub,
	}
}

// HandleDirectMessageSent handles the DirectMessageSentEvent
// It sends the message to the recipient via WebSocket if they are online.
func (s *NotificationService) HandleDirectMessageSent(ctx context.Context, event domain.Event) error {
	e, ok := event.(service.DirectMessageSentEvent)
	if !ok {
		return nil
	}

	// Run in goroutine to not block the event bus (and thus the HTTP response)
	go func() {
		// Create payload matching what the frontend expects
		payload := map[string]interface{}{
			"type":            "direct_message",
			"id":              e.MessageID,
			"conversation_id": e.ConversationID,
			"sender_id":       e.SenderID,
			"content":         e.Content,
			"message_type":    e.Type,
			"created_at":      e.CreatedAt,
			// sequence_number is not in event, but frontend might need it.
			// Ideally, event should carry it, but for now this is "optimistic" push.
			// If strict ordering is needed, frontend should rely on HTTP response or fetch.
			// However, for real-time chat, usually we want to show it immediately.
		}

		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			logger.Error("Failed to marshal notification payload", zap.Error(err))
			return
		}

		// Send to Recipient
		s.hub.BroadcastToUsers([]string{e.RecipientID}, jsonPayload)
		
		// Send back to Sender (for multi-device sync and real-time confirmation)
		// This makes the chat feel "live" even if the sender is on multiple devices or to confirm the server processed it.
		// Frontend should handle deduplication if it optimistically added the message.
		s.hub.BroadcastToUsers([]string{e.SenderID}, jsonPayload)
		
		logger.Info("Notification sent via WebSocket", 
			zap.String("recipient_id", e.RecipientID),
			zap.String("sender_id", e.SenderID),
			zap.String("conversation_id", e.ConversationID))
	}()

	return nil
}
