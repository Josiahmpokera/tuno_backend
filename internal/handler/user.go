package handler

import (
	"net/http"
	"tuno_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	bus *service.CommandBus
}

func NewUserHandler(bus *service.CommandBus) *UserHandler {
	return &UserHandler{bus: bus}
}

type UpdateProfileRequest struct {
	UserID   string `json:"user_id" binding:"required"` // In real app, get from token
	Name     string `json:"name"`
	PhotoURL string `json:"photo_url"`
}

func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.UpdateProfileCommand{
		UserID:   req.UserID,
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
