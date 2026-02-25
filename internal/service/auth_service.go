package service

import (
	"context"
	"fmt"
	"time"
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

type RegisterUserCommand struct {
	PhoneNumber string
	Name        string
	PhotoURL    string
	OTP         string
}

func (c RegisterUserCommand) CommandName() string {
	return "RegisterUserCommand"
}

type LoginUserCommand struct {
	PhoneNumber string
	OTP         string
}

func (c LoginUserCommand) CommandName() string {
	return "LoginUserCommand"
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

type RegisterUserHandler struct {
	repo       domain.UserRepository
	otpService *OtpService
}

func NewRegisterUserHandler(repo domain.UserRepository, otpService *OtpService) *RegisterUserHandler {
	return &RegisterUserHandler{repo: repo, otpService: otpService}
}

func (h *RegisterUserHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(RegisterUserCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Verify OTP
	if err := h.otpService.VerifyOTP(ctx, c.PhoneNumber, c.OTP); err != nil {
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// 2. Validate Business Rules
	existingUser, err := h.repo.FindByPhoneNumber(c.PhoneNumber)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, fmt.Errorf("user with phone number %s already exists", c.PhoneNumber)
	}

	// 3. Create User Entity
	user := &domain.User{
		ID:          uuid.New().String(),
		PhoneNumber: c.PhoneNumber,
		Name:        c.Name,
		PhotoURL:    c.PhotoURL,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// 4. Persist
	if err := h.repo.Create(user); err != nil {
		return nil, err
	}

	// 5. Return result (Token generation would happen here in a real app)
	return map[string]string{
		"user_id": user.ID,
		"token":   "mock-jwt-token-for-" + user.ID, // TODO: Implement real JWT
	}, nil
}

type LoginUserHandler struct {
	repo       domain.UserRepository
	otpService *OtpService
}

func NewLoginUserHandler(repo domain.UserRepository, otpService *OtpService) *LoginUserHandler {
	return &LoginUserHandler{repo: repo, otpService: otpService}
}

func (h *LoginUserHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(LoginUserCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Verify OTP
	if err := h.otpService.VerifyOTP(ctx, c.PhoneNumber, c.OTP); err != nil {
		return nil, fmt.Errorf("OTP verification failed: %w", err)
	}

	// 2. Find User
	user, err := h.repo.FindByPhoneNumber(c.PhoneNumber)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 3. Return Token
	return map[string]string{
		"user_id": user.ID,
		"token":   "mock-jwt-token-for-" + user.ID, // TODO: Implement real JWT
	}, nil
}
