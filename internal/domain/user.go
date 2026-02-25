package domain

import (
	"time"
)

type User struct {
	ID          string    `json:"id" db:"id"`
	PhoneNumber string    `json:"phone_number" db:"phone_number"`
	Name        string    `json:"name" db:"name"`
	PhotoURL    string    `json:"photo_url" db:"photo_url"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

type UserRepository interface {
	Create(user *User) error
	FindByPhoneNumber(phoneNumber string) (*User, error)
	FindByID(id string) (*User, error)
}
