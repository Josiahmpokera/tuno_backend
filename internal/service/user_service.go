package service

import (
	"context"
	"fmt"
	"time"
	"tuno_backend/internal/domain"
)

// Commands

type UpdateProfileCommand struct {
	UserID   string
	Name     string
	PhotoURL string
}

func (c UpdateProfileCommand) CommandName() string {
	return "UpdateProfileCommand"
}

// Handlers

type UpdateProfileHandler struct {
	repo domain.UserRepository
}

func NewUpdateProfileHandler(repo domain.UserRepository) *UpdateProfileHandler {
	return &UpdateProfileHandler{repo: repo}
}

func (h *UpdateProfileHandler) Handle(ctx context.Context, cmd interface{}) (interface{}, error) {
	c, ok := cmd.(UpdateProfileCommand)
	if !ok {
		return nil, fmt.Errorf("invalid command type")
	}

	// 1. Find User
	user, err := h.repo.FindByID(c.UserID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, fmt.Errorf("user not found")
	}

	// 2. Update Fields
	if c.Name != "" {
		user.Name = c.Name
	}
	if c.PhotoURL != "" {
		user.PhotoURL = c.PhotoURL
	}
	user.UpdatedAt = time.Now()

	// 3. Persist
	if err := h.repo.Update(user); err != nil {
		return nil, err
	}

	return user, nil
}
