package db

import (
	"context"
	"fmt"
	"net/url"
	"tuno_backend/internal/config"
	"tuno_backend/pkg/logger"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func NewPostgres(cfg config.DatabaseConfig) (*pgxpool.Pool, error) {
	// Construct connection string properly using url.URL to handle special characters
	u := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(cfg.User, cfg.Password),
		Host:   fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
		Path:   cfg.Name,
	}

	// Add query parameters
	q := u.Query()
	q.Set("sslmode", cfg.SSLMode)
	u.RawQuery = q.Encode()

	dsn := u.String()

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database config: %w", err)
	}

	// Optional: Configure pool settings (max conns, etc.)
	poolConfig.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	logger.Info("Connected to PostgreSQL", zap.String("host", cfg.Host), zap.String("db", cfg.Name))
	return pool, nil
}
