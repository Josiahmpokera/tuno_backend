package main

import (
	"context"
	"fmt"
	"os"

	"tuno_backend/internal/config"
	"tuno_backend/internal/db"
	"tuno_backend/internal/handler"
	"tuno_backend/internal/repository"
	"tuno_backend/internal/service"
	"tuno_backend/internal/websocket"
	"tuno_backend/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// 2. Initialize Logger
	logger.Init(cfg.Server.Env)
	defer logger.Log.Sync() // Flushes buffer, if any

	logger.Info("Starting Tuno Backend", zap.String("env", cfg.Server.Env))

	// 3. Connect to Database (PostgreSQL)
	pgPool, err := db.NewPostgres(cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to PostgreSQL", zap.Error(err))
	}
	defer pgPool.Close()

	// 4. Connect to Redis
	redisClient, err := db.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 5. Initialize Services & Handlers
	// Repositories
	userRepo := repository.NewPostgresUserRepository(pgPool)

	// Services
	otpService := service.NewOtpService(redisClient)

	// Command Bus
	commandBus := service.NewCommandBus()

	// Register Command Handlers
	sendOtpHandler := service.NewSendOtpHandler(otpService)
	commandBus.Register("SendOtpCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return sendOtpHandler.Handle(ctx, cmd)
	})

	registerUserHandler := service.NewRegisterUserHandler(userRepo, otpService)
	commandBus.Register("RegisterUserCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return registerUserHandler.Handle(ctx, cmd)
	})

	loginUserHandler := service.NewLoginUserHandler(userRepo, otpService)
	commandBus.Register("LoginUserCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return loginUserHandler.Handle(ctx, cmd)
	})

	// API Handlers
	authHandler := handler.NewAuthHandler(commandBus)

	// 6. Setup Gin Router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize WebSocket Hub
	hub := websocket.NewHub()
	go hub.Run()

	r := gin.New()
	r.Use(gin.Recovery())
	// Use a custom logger middleware if needed, for now standard is okay
	r.Use(gin.Logger())

	// Routes
	api := r.Group("/api/v1")
	{
		api.GET("/", handler.Home)
		api.GET("/health", handler.HealthCheck)
		api.GET("/ws", func(c *gin.Context) {
			websocket.ServeWs(hub, c)
		})

		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/otp", authHandler.SendOtp)
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
		}
	}

	// 7. Start Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server listening", zap.String("port", cfg.Server.Port))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
