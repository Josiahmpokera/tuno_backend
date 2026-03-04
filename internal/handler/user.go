package handler

import (
	"net/http"
	"time"
	"tuno_backend/internal/domain"
	"tuno_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	bus  *service.CommandBus
	repo domain.UserRepository
}

func NewUserHandler(bus *service.CommandBus, repo domain.UserRepository) *UserHandler {
	return &UserHandler{bus: bus, repo: repo}
}

type UpdateProfileRequest struct {
	Name     string `json:"name"`
	PhotoURL string `json:"photo_url"`
}

type RegisteredContactsRequest struct {
	PhoneNumbers []string `json:"phone_numbers" binding:"required"`
}

type RegisteredContactView struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phone_number"`
	Name        string `json:"name"`
	PhotoURL    string `json:"photo_url"`
	IsOnline    bool   `json:"is_online"`
	LastSeenAt  string `json:"last_seen_at"`
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	user, err := h.repo.FindByID(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user"})
		return
	}
	if user == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) GetRegisteredContacts(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req RegisteredContactsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(req.PhoneNumbers) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone_numbers is required"})
		return
	}
	if len(req.PhoneNumbers) > 2000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "phone_numbers exceeds maximum size"})
		return
	}

	users, err := h.repo.FindByPhoneNumbers(req.PhoneNumbers)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to lookup contacts"})
		return
	}

	registered := make([]RegisteredContactView, 0, len(users))
	for _, u := range users {
		lastSeen := ""
		if !u.LastSeenAt.IsZero() {
			lastSeen = u.LastSeenAt.UTC().Format(time.RFC3339)
		}
		registered = append(registered, RegisteredContactView{
			ID:          u.ID,
			PhoneNumber: u.PhoneNumber,
			Name:        u.Name,
			PhotoURL:    u.PhotoURL,
			IsOnline:    u.IsOnline,
			LastSeenAt:  lastSeen,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"registered_contacts": registered,
	})
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.UpdateProfileCommand{
		UserID:   userID,
		Name:     req.Name,
		PhotoURL: req.PhotoURL,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
