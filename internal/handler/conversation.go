package handler

import (
	"net/http"
	"strconv"
	"tuno_backend/internal/domain"
	"tuno_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type ConversationHandler struct {
	bus  *service.CommandBus
	repo domain.ConversationRepository
	dmRepo domain.DirectMessageRepository
}

func NewConversationHandler(
	bus *service.CommandBus,
	repo domain.ConversationRepository,
	dmRepo domain.DirectMessageRepository,
) *ConversationHandler {
	return &ConversationHandler{
		bus:    bus,
		repo:   repo,
		dmRepo: dmRepo,
	}
}

type StartConversationRequest struct {
	CommandID   string `json:"command_id" binding:"required"`
	RecipientID string `json:"recipient_id" binding:"required"`
}

func (h *ConversationHandler) StartConversation(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req StartConversationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.StartConversationCommand{
		CommandID: req.CommandID,
		User1ID:   userID,
		User2ID:   req.RecipientID,
	}

	result, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	convID, ok := result.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unexpected result type"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"conversation_id": convID,
		"message":         "Conversation started successfully",
	})
}

type SendDirectMessageRequest struct {
	CommandID string             `json:"command_id" binding:"required"`
	Content   string             `json:"content" binding:"required"`
	Type      domain.MessageType `json:"type" binding:"required"`
}

func (h *ConversationHandler) SendMessage(c *gin.Context) {
	userID := c.GetString("user_id")
	convID := c.Param("id")

	var req SendDirectMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.SendDirectMessageCommand{
		CommandID:      req.CommandID,
		SenderID:       userID,
		ConversationID: convID,
		Content:        req.Content,
		Type:           req.Type,
	}

	_, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{"message": "Message sent"})
}

func (h *ConversationHandler) GetConversations(c *gin.Context) {
	userID := c.GetString("user_id")
	
	conversations, err := h.repo.GetUserConversations(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func (h *ConversationHandler) GetMessages(c *gin.Context) {
	convID := c.Param("id")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	messages, err := h.dmRepo.GetByConversationID(convID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
