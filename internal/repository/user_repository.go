package repository

import (
	"context"
	"fmt"
	"time"
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

func (r *PostgresUserRepository) UpdatePresence(userID string, isOnline bool) error {
	query := `
		UPDATE users
		SET is_online = $1, last_seen_at = NOW()
		WHERE id = $2
	`
	_, err := r.pool.Exec(context.Background(), query, isOnline, userID)
	if err != nil {
		return fmt.Errorf("failed to update user presence: %w", err)
	}
	return nil
}

func (r *PostgresUserRepository) FindByPhoneNumber(phoneNumber string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, name, photo_url, is_online, last_seen_at, created_at, updated_at
		FROM users
		WHERE phone_number = $1
	`
	row := r.pool.QueryRow(context.Background(), query, phoneNumber)

	var user domain.User
	var isOnline *bool
	var lastSeenAt *time.Time

	err := row.Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.PhotoURL,
		&isOnline,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find user by phone number: %w", err)
	}

	if isOnline != nil {
		user.IsOnline = *isOnline
	}
	if lastSeenAt != nil {
		user.LastSeenAt = *lastSeenAt
	}

	return &user, nil
}

func (r *PostgresUserRepository) FindByPhoneNumbers(phoneNumbers []string) ([]domain.User, error) {
	if len(phoneNumbers) == 0 {
		return []domain.User{}, nil
	}

	query := `
		SELECT id, phone_number, name, photo_url, is_online, last_seen_at, created_at, updated_at
		FROM users
		WHERE phone_number = ANY($1)
	`
	rows, err := r.pool.Query(context.Background(), query, phoneNumbers)
	if err != nil {
		return nil, fmt.Errorf("failed to find users by phone numbers: %w", err)
	}
	defer rows.Close()

	users := make([]domain.User, 0)
	for rows.Next() {
		var user domain.User
		var isOnline *bool
		var lastSeenAt *time.Time
		if err := rows.Scan(
			&user.ID,
			&user.PhoneNumber,
			&user.Name,
			&user.PhotoURL,
			&isOnline,
			&lastSeenAt,
			&user.CreatedAt,
			&user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}

		if isOnline != nil {
			user.IsOnline = *isOnline
		}
		if lastSeenAt != nil {
			user.LastSeenAt = *lastSeenAt
		}

		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	return users, nil
}

func (r *PostgresUserRepository) FindByID(id string) (*domain.User, error) {
	query := `
		SELECT id, phone_number, name, photo_url, is_online, last_seen_at, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	row := r.pool.QueryRow(context.Background(), query, id)

	var user domain.User
	var isOnline *bool
	var lastSeenAt *time.Time

	err := row.Scan(
		&user.ID,
		&user.PhoneNumber,
		&user.Name,
		&user.PhotoURL,
		&isOnline,
		&lastSeenAt,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Not found
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	if isOnline != nil {
		user.IsOnline = *isOnline
	}
	if lastSeenAt != nil {
		user.LastSeenAt = *lastSeenAt
	}

	return &user, nil
}
