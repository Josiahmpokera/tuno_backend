package handler

import (
	"net/http"
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
