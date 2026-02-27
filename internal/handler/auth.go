package handler

import (
	"net/http"
	"tuno_backend/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	bus *service.CommandBus
}

func NewAuthHandler(bus *service.CommandBus) *AuthHandler {
	return &AuthHandler{bus: bus}
}

// SendOtp Request DTO
type SendOtpRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
}

// VerifyOTP Request DTO
type VerifyOTPRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	OTP         string `json:"otp" binding:"required"`
}

func (h *AuthHandler) SendOtp(c *gin.Context) {
	var req SendOtpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.SendOtpCommand{
		PhoneNumber: req.PhoneNumber,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}

// RegisterRequest DTO
type RegisterRequest struct {
	Name     string `json:"name" binding:"required"`
	PhotoURL string `json:"photo_url"`
}

// Register completes the user registration by setting name and photo
func (h *AuthHandler) Register(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var req RegisterRequest
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

func (h *AuthHandler) VerifyOTP(c *gin.Context) {
	var req VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.VerifyOTPCommand{
		PhoneNumber: req.PhoneNumber,
		OTP:         req.OTP,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
