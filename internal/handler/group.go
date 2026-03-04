package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"
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

type AddGroupMemberByPhoneRequest struct {
	CommandID   string `json:"command_id" binding:"required"`
	PhoneNumber string `json:"phone_number" binding:"required"`
	Timestamp   int64  `json:"timestamp" binding:"required"`
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

func (h *GroupHandler) AddMemberByPhone(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

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

	var req AddGroupMemberByPhoneRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.AddGroupMemberByPhoneCommand{
		CommandID:   req.CommandID,
		GroupID:     groupID,
		PhoneNumber: req.PhoneNumber,
		AddedByID:   userIDStr,
		Timestamp:   req.Timestamp,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

func (h *GroupHandler) GetGroupMembers(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	// Parse pagination parameters
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 0 {
		limit = 50
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil || offset < 0 {
		offset = 0
	}

	// Get paginated members
	members, totalCount, err := h.repo.GetGroupMembersPaginated(groupID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate pagination metadata
	hasMore := (offset + len(members)) < totalCount
	nextOffset := offset + len(members)

	response := gin.H{
		"members": members,
		"pagination": gin.H{
			"total_count": totalCount,
			"limit":       limit,
			"offset":      offset,
			"has_more":    hasMore,
			"next_offset": nextOffset,
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetUserGroups(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	userIDStr, ok := userID.(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user ID"})
		return
	}

	groups, err := h.repo.GetGroupsByUserID(userIDStr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Create a well-structured response
	response := gin.H{
		"success": true,
		"data": gin.H{
			"groups":  groups,
			"count":   len(groups),
			"user_id": userIDStr,
		},
		"metadata": gin.H{
			"timestamp": time.Now().Format(time.RFC3339),
			"version":   "v1",
		},
	}

	c.JSON(http.StatusOK, response)
}

func (h *GroupHandler) GetGroupMessages(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 50
	}

	before := c.Query("before")
	var beforeTime *int64
	if before != "" {
		ts, err := strconv.ParseInt(before, 10, 64)
		if err == nil {
			beforeTime = &ts
		}
	}

	messages, err := h.messageRepo.GetMessages(groupID, limit, beforeTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, messages)
}

func (h *GroupHandler) GetGroupHome(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Parse field selection parameters
	fieldsParam := c.DefaultQuery("fields", "basic,round,members,stats")
	fields := strings.Split(fieldsParam, ",")

	// Create a map for quick lookup
	fieldSet := make(map[string]bool)
	for _, field := range fields {
		fieldSet[strings.TrimSpace(field)] = true
	}

	resp, err := h.repo.GetGroupHome(groupID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Filter response based on requested fields
	filteredResponse := gin.H{}

	if fieldSet["basic"] || fieldSet["all"] {
		filteredResponse["group"] = resp.Group
	}

	if (fieldSet["round"] || fieldSet["all"]) && resp.CurrentRound != nil {
		filteredResponse["current_round"] = resp.CurrentRound
	}

	if fieldSet["members"] || fieldSet["all"] {
		filteredResponse["members"] = resp.Members
	}

	if fieldSet["stats"] || fieldSet["all"] {
		filteredResponse["stats"] = resp.Stats
	}

	// If no specific fields requested, return full response
	if len(fieldSet) == 0 || fieldSet["all"] {
		c.JSON(http.StatusOK, resp)
		return
	}

	c.JSON(http.StatusOK, filteredResponse)
}

func (h *GroupHandler) GetGroupDetails(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	group, err := h.repo.GetGroupDetail(groupID, userID.(string))
	if err != nil {
		if strings.Contains(err.Error(), "not found") || strings.Contains(err.Error(), "access denied") {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, group)
}

type SendMessageRequest struct {
	Content string             `json:"content" binding:"required"`
	Type    domain.MessageType `json:"type"`
}

func (h *GroupHandler) SendMessage(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req SendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default to TEXT if not provided
	if req.Type == "" {
		req.Type = domain.MessageTypeText
	}

	cmd := service.SendMessageCommand{
		UserID:  userID.(string),
		GroupID: groupID,
		Content: req.Content,
		Type:    req.Type,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *GroupHandler) GetGroupSchedule(c *gin.Context) {
	groupID := c.Param("id")
	if groupID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "group_id is required"})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	// Verify membership
	member, err := h.repo.GetMember(groupID, userID.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check membership"})
		return
	}
	if member == nil || member.Status != domain.StatusActive {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied: You are not a member of this group"})
		return
	}

	statusStr := c.Query("status")
	var filterStatus *domain.RoundStatus
	if statusStr != "" {
		s := domain.RoundStatus(statusStr)
		filterStatus = &s
	}

	resp, err := h.repo.GetGroupSchedule(groupID, filterStatus)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}
