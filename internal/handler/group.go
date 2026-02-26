package handler

import (
	"net/http"
	"strconv"
	"tuno_backend/internal/domain"
	"tuno_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type GroupHandler struct {
	bus         *service.CommandBus
	repo        domain.GroupRepository
	messageRepo domain.MessageRepository
}

func NewGroupHandler(bus *service.CommandBus, repo domain.GroupRepository, messageRepo domain.MessageRepository) *GroupHandler {
	return &GroupHandler{bus: bus, repo: repo, messageRepo: messageRepo}
}

type CreateGroupRequest struct {
	CommandID           string                   `json:"command_id" binding:"required"`
	Name                string                   `json:"name" binding:"required"`
	PhotoURL            string                   `json:"photo_url"`
	ContributionAmount  float64                  `json:"contribution_amount" binding:"required"`
	RotationFrequency   domain.RotationFrequency `json:"rotation_frequency" binding:"required"`
	CustomFrequencyDays int                      `json:"custom_frequency_days"`
	Invitees            []string                 `json:"invitees"` // List of phone numbers
	Timestamp           int64                    `json:"timestamp" binding:"required"`
}

func (h *GroupHandler) CreateGroup(c *gin.Context) {
	// 1. Authenticate (Get UserID from Context set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Invalid user ID format"})
		return
	}

	// 2. Parse input
	var req CreateGroupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 3. Validate input (Basic)
	if req.RotationFrequency == domain.FrequencyCustom && req.CustomFrequencyDays <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "custom_frequency_days must be positive for CUSTOM frequency"})
		return
	}

	// 4. Create command
	cmd := service.CreateGroupCommand{
		CommandID:           req.CommandID,
		Name:                req.Name,
		PhotoURL:            req.PhotoURL,
		ContributionAmount:  req.ContributionAmount,
		RotationFrequency:   req.RotationFrequency,
		CustomFrequencyDays: req.CustomFrequencyDays,
		CreatorID:           userIDStr,
		Invitees:            req.Invitees,
		Timestamp:           req.Timestamp,
	}

	// 5. Dispatch command
	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 6. Respond
	c.JSON(http.StatusCreated, res)
}

type SendMessageRequest struct {
	Content string             `json:"content" binding:"required"`
	Type    domain.MessageType `json:"type"`
}

func (h *GroupHandler) SendMessage(c *gin.Context) {
	// 1. Authenticate (Get UserID from Context set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Invalid user ID format"})
		return
	}

	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Group ID is required"})
		return
	}

	// 2. Parse input
	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Type == "" {
		req.Type = domain.MessageTypeText
	}

	// 3. Create command
	cmd := service.SendMessageCommand{
		UserID:  userIDStr,
		GroupID: groupID,
		Content: req.Content,
		Type:    req.Type,
	}

	// 4. Dispatch command
	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 5. Respond
	c.JSON(http.StatusCreated, res)
}

func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	// 1. Authenticate (Get UserID from Context set by AuthMiddleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Invalid user ID format"})
		return
	}

	// 2. Query Repository
	groups, err := h.repo.GetGroupsByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch groups: " + err.Error()})
		return
	}

	// 3. Respond
	if groups == nil {
		// Return empty array instead of null
		groups = []domain.GroupWithRole{}
	}
	c.JSON(http.StatusOK, groups)
}

func (h *GroupHandler) GetGroupDetails(c *gin.Context) {
	groupID := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	detail, err := h.repo.GetGroupDetail(groupID, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group details: " + err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Group not found or access denied"})
		return
	}

	c.JSON(http.StatusOK, detail)
}

func (h *GroupHandler) GetGroupMessages(c *gin.Context) {
	groupID := c.Param("id")
	// Verify membership implicitly by checking if user can see group details first?
	// Or trust the repo/service layer. For now, we can check group membership via GetGroupDetail
	// or assume the query handles it (MessageRepo usually doesn't check membership unless we join).
	// Ideally, we should check membership first.

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr := userID.(string)

	// Simple membership check
	_, err := h.repo.GetGroupDetail(groupID, userIDStr)
	if err != nil || err == nil && false { // Wait, GetGroupDetail returns nil if not found/not member
		// But wait, GetGroupDetail query in repo returns rows only if member.
	}
	// Let's call GetGroupDetail to verify access
	detail, err := h.repo.GetGroupDetail(groupID, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify group access: " + err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Pagination
	limit := 50
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	offset := 0
	if o := c.Query("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	messages, err := h.messageRepo.GetByGroupID(groupID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages: " + err.Error()})
		return
	}
	if messages == nil {
		messages = []domain.Message{}
	}

	c.JSON(http.StatusOK, messages)
}

func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	// 1. Authenticate
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: User not authenticated"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error: Invalid user ID format"})
		return
	}

	// 2. Verify Membership
	detail, err := h.repo.GetGroupDetail(groupID, userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify group access: " + err.Error()})
		return
	}
	if detail == nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: You are not a member of this group"})
		return
	}

	// 3. Get Members
	members, err := h.repo.GetGroupMembers(groupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch group members: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, members)
}
