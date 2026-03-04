package domain

import (
	"time"
)

type User struct {
	ID          string    `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Name        string    `json:"name" db:"name"`
	PhotoURL    string    `json:"photo_url" db:"photo_url"`
	IsOnline    bool      `json:"is_online" db:"is_online"`
	LastSeenAt  time.Time `json:"last_seen_at" db:"last_seen_at"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UserRepository interface {
	Create(user *User) error
	Update(user *User) error
	UpdatePresence(userID string, isOnline bool) error
	FindByPhoneNumber(phoneNumber string) (*User, error)
	FindByPhoneNumbers(phoneNumbers []string) ([]User, error)
	FindByID(id string) (*User, error)
}
