package service

import (
	"context"
	"fmt"
	"time"
	"tuno_backend/internal/auth"
	"tuno_backend/internal/domain"

	"github.com/google/uuid"
)

// Commands

type SendOtpCommand struct {
	PhoneNumber string
}

func (c SendOtpCommand) CommandName() string {
	return "SendOtpCommand"
}

type VerifyOTPCommand struct {
	PhoneNumber string
	OTP         string
}

func (c VerifyOTPCommand) CommandName() string {
	return "VerifyOTPCommand"
}

// Handlers

// SendOtpHandler
type SendOtpHandler struct {
	otpService *OtpService
}

func NewSendOtpHandler(otpService *OtpService) *SendOtpHandler {
	return &SendOtpHandler{otpService: otpService}
}

func (h *SendOtpHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(SendOtpCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	otp, err := h.otpService.GenerateOTP(ctx, c.PhoneNumber)
	if err != nil {
		return nil, err
	}

	// For development, return the OTP directly
	return map[string]string{
		"message": "OTP sent successfully",
		"otp":     otp,
	}, nil
}

type VerifyOTPHandler struct {
	repo       domain.UserRepository
	otpService *OtpService
	jwtService *auth.JWTService
}

func NewVerifyOTPHandler(repo domain.UserRepository, otpService *OtpService, jwtService *auth.JWTService) *VerifyOTPHandler {
	return &VerifyOTPHandler{repo: repo, otpService: otpService, jwtService: jwtService}
}

func (h *VerifyOTPHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(VerifyOTPCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Verify OTP
	if err := h.otpService.VerifyOTP(ctx, c.PhoneNumber, c.OTP); err != nil {
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// 2. Check if user exists
	user, err := h.repo.FindByPhoneNumber(c.PhoneNumber)
	if err != nil {
		return nil, err
	}

	var isNewUser bool

	if user == nil {
		// Create new user
		isNewUser = true
		user = &domain.User{
			ID:          uuid.New().String(),
			PhoneNumber: c.PhoneNumber,
			Name:        "", // Empty name for new users
			PhotoURL:    "",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		if err := h.repo.Create(user); err != nil {
			return nil, err
		}
	} else {
		isNewUser = false
	}

	// 3. Generate JWT Token
	token, err := h.jwtService.GenerateToken(user.ID, user.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// 4. Return result
	return map[string]interface{}{
		"user_id":     user.ID,
		"token":       token,
		"is_new_user": isNewUser,
		"user":        user,
	}, nil
}
