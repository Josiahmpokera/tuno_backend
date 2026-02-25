package service

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type OtpService struct {
	redisClient *redis.Client
}

func NewOtpService(redisClient *redis.Client) *OtpService {
	return &OtpService{redisClient: redisClient}
}

// GenerateOTP generates a 6-digit OTP and stores it in Redis with a 5-minute TTL.
func (s *OtpService) GenerateOTP(ctx context.Context, phoneNumber string) (string, error) {
	// Generate 6-digit OTP
	rand.Seed(time.Now().UnixNano())
	otp := fmt.Sprintf("%06d", rand.Intn(1000000))

	// Store in Redis
	key := fmt.Sprintf("otp:%s", phoneNumber)
	err := s.redisClient.Set(ctx, key, otp, 5*time.Minute).Err()
	if err != nil {
		return "", fmt.Errorf("failed to store OTP: %w", err)
	}

	return otp, nil
}

// VerifyOTP checks if the provided OTP matches the one stored in Redis.
func (s *OtpService) VerifyOTP(ctx context.Context, phoneNumber, otp string) error {
	key := fmt.Sprintf("otp:%s", phoneNumber)
	storedOTP, err := s.redisClient.Get(ctx, key).Result()

	if err == redis.Nil {
		return fmt.Errorf("OTP expired or not found")
	} else if err != nil {
		return fmt.Errorf("failed to retrieve OTP: %w", err)
	}

	if storedOTP != otp {
		return fmt.Errorf("invalid OTP")
	}

	// Delete OTP after successful verification to prevent reuse
	s.redisClient.Del(ctx, key)

	return nil
}
