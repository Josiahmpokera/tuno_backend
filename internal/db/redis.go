package db

import (
	"context"
	"fmt"
	"tuno_backend/internal/config"
	"tuno_backend/pkg/logger"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func NewRedis(cfg config.RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("unable to connect to redis: %w", err)
	}

	logger.Info("Connected to Redis", zap.String("addr", cfg.Addr))
	return client, nil
}
