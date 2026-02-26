package repository

import (
	"context"
	"fmt"
	"tuno_backend/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &PostgresUserRepository{pool: pool}
}

func (r *PostgresUserRepository) Create(user *domain.User) error {
	query := `
		INSERT INTO users (id, phone_number, name, photo_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`
	_, err := r.pool.Exec(context.Background(), query,
		user.ID,
		user.PhoneNumber,
		user.Name,
		user.PhotoURL,
		user.CreatedAt,
		user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) Update(user *domain.User) error {
	query := `
		UPDATE users
		SET name = $1, photo_url = $2, updated_at = $3
		WHERE id = $4
	`
	_, err := r.pool.Exec(context.Background(), query,
		user.Name,
		user.PhotoURL,
		user.UpdatedAt,
		user.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) FindByPhoneNumber(phoneNumber string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, name, photo_url, created_at, updated_at
		FROM users
		WHERE phone_number = $1
	`
	row := r.pool.QueryRow(context.Background(), query, phoneNumber)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.PhotoURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find user by phone number: %w", err)
	}
	return &user, nil
}

func (r *PostgresUserRepository) FindByID(id string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, name, photo_url, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(context.Background(), query, id)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.PhotoURL,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &user, nil
}
