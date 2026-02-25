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

// Register Request DTO
type RegisterRequest struct {
	PhoneNumber string `json:"phone_number" binding:"required"`
	Name        string `json:"name" binding:"required"`
	PhotoURL    string `json:"photo_url"`
	OTP         string `json:"otp" binding:"required"`
}

// Login Request DTO
type LoginRequest struct {
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.RegisterUserCommand{
		PhoneNumber: req.PhoneNumber,
		Name:        req.Name,
		PhotoURL:    req.PhotoURL,
		OTP:         req.OTP,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		// In a real app, map errors to specific status codes (e.g., 409 Conflict)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := service.LoginUserCommand{
		PhoneNumber: req.PhoneNumber,
		OTP:         req.OTP,
	}

	res, err := h.bus.Dispatch(c.Request.Context(), cmd)
	if err != nil {
		// e.g., 404 Not Found or 401 Unauthorized
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, res)
}
