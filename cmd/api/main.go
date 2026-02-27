package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"tuno_backend/internal/auth"
	"tuno_backend/internal/config"
	"tuno_backend/internal/db"
	"tuno_backend/internal/domain"
	"tuno_backend/internal/handler"
	"tuno_backend/internal/middleware"
	"tuno_backend/internal/notification"
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

	// Initialize Database Schema
	if err := db.InitSchema(pgPool); err != nil {
		logger.Fatal("Failed to initialize database schema", zap.Error(err))
	}

	// 4. Connect to Redis
	redisClient, err := db.NewRedis(cfg.Redis)
	if err != nil {
		logger.Fatal("Failed to connect to Redis", zap.Error(err))
	}
	defer redisClient.Close()

	// 5. Initialize Services & Handlers
	// Repositories
	userRepo := repository.NewPostgresUserRepository(pgPool)
	groupRepo := repository.NewPostgresGroupRepository(pgPool)
	messageRepo := repository.NewPostgresMessageRepository(pgPool)
	conversationRepo := repository.NewPostgresConversationRepository(pgPool)
	dmRepo := repository.NewPostgresDirectMessageRepository(pgPool)

	// Services
	otpService := service.NewOtpService(redisClient)
	jwtService := auth.NewJWTService(cfg.JWT)

	// Event Bus
	eventBus := service.NewEventBus()

	// Event Handlers
	groupCreatedHandler := service.NewGroupCreatedEventHandler(groupRepo)
	eventBus.Subscribe("GroupCreatedEvent", func(ctx context.Context, event domain.Event) error {
		return groupCreatedHandler.Handle(ctx, event)
	})

	conversationStartedHandler := service.NewConversationStartedEventHandler(conversationRepo)
	eventBus.Subscribe("ConversationStartedEvent", func(ctx context.Context, event domain.Event) error {
		return conversationStartedHandler.Handle(ctx, event)
	})

	dmSentHandler := service.NewDirectMessageSentEventHandler(dmRepo)
	eventBus.Subscribe("DirectMessageSentEvent", func(ctx context.Context, event domain.Event) error {
		return dmSentHandler.Handle(ctx, event)
	})

	// Command Bus
	commandBus := service.NewCommandBus()

	// Register Command Handlers
	sendOtpHandler := service.NewSendOtpHandler(otpService)
	commandBus.Register("SendOtpCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return sendOtpHandler.Handle(ctx, cmd)
	})

	verifyOTPHandler := service.NewVerifyOTPHandler(userRepo, otpService, jwtService)
	commandBus.Register("VerifyOTPCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return verifyOTPHandler.Handle(ctx, cmd)
	})

	updateProfileHandler := service.NewUpdateProfileHandler(userRepo)
	commandBus.Register("UpdateProfileCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return updateProfileHandler.Handle(ctx, cmd)
	})

	createGroupHandler := service.NewCreateGroupHandler(eventBus, userRepo, redisClient)
	commandBus.Register("CreateGroupCommand", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		return createGroupHandler.Handle(ctx, cmd)
	})

	startConversationHandler := service.NewStartConversationHandler(conversationRepo, groupRepo, eventBus, redisClient)
	commandBus.Register("StartConversation", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		// Since our handler expects service.Command (which is interface{}), and cmd is interface{}, we need type assertion in handler
		// But wait, previous handlers took specific types in Handle?
		// No, previously I implemented Handle taking `cmd Command` (interface) in `conversation_service.go`.
		// However, `commandBus.Register` expects a function `func(context.Context, interface{}) (interface{}, error)`.
		// So here we pass `cmd` directly.
		// Wait, `StartConversationHandler.Handle` takes `service.Command`.
		// `service.Command` is likely `interface{ CommandName() string }`.
		// If `cmd` passed here is just the struct, it implements it.
		// I need to cast `cmd` to `service.Command`.
		if c, ok := cmd.(service.Command); ok {
			return startConversationHandler.Handle(ctx, c)
		}
		return nil, fmt.Errorf("invalid command type")
	})

	sendDMHandler := service.NewSendDirectMessageHandler(conversationRepo, dmRepo, groupRepo, eventBus, redisClient)
	commandBus.Register("SendDirectMessage", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		if c, ok := cmd.(service.Command); ok {
			return sendDMHandler.Handle(ctx, c)
		}
		return nil, fmt.Errorf("invalid command type")
	})

	markReadHandler := service.NewMarkMessagesReadHandler(dmRepo, conversationRepo, eventBus, redisClient)
	commandBus.Register("MarkMessagesRead", func(ctx context.Context, cmd interface{}) (interface{}, error) {
		if c, ok := cmd.(service.Command); ok {
			return markReadHandler.Handle(ctx, c)
		}
		return nil, fmt.Errorf("invalid command type")
	})

	messageService := service.NewMessageService(groupRepo, messageRepo)
	commandBus.Register("SendMessageCommand", messageService.SendMessage)

	// API Handlers
	authHandler := handler.NewAuthHandler(commandBus)
	userHandler := handler.NewUserHandler(commandBus)
	groupHandler := handler.NewGroupHandler(commandBus, groupRepo, messageRepo)
	conversationHandler := handler.NewConversationHandler(commandBus, conversationRepo, dmRepo)

	// 6. Setup Gin Router
	if cfg.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize WebSocket Hub
	hub := websocket.NewHub(userRepo)
	go hub.Run()

	// Notification Service (WebSocket Glue)
	notificationService := notification.NewNotificationService(hub)
	eventBus.Subscribe("DirectMessageSentEvent", func(ctx context.Context, event domain.Event) error {
		return notificationService.HandleDirectMessageSent(ctx, event)
	})
	eventBus.Subscribe("MessagesReadEvent", func(ctx context.Context, event domain.Event) error {
		return notificationService.HandleMessagesRead(ctx, event)
	})

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
			// Authenticate
			tokenStr := c.Query("token")
			if tokenStr == "" {
				// Fallback to Authorization header
				authHeader := c.GetHeader("Authorization")
				if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
					tokenStr = authHeader[7:]
				}
			}

			if tokenStr == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
				return
			}

			// Validate token
			claims, err := jwtService.ValidateToken(tokenStr)
			if err != nil {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
				return
			}

			userID := claims.UserID
			websocket.ServeWs(hub, c.Writer, c.Request, userID, commandBus)
		})

		// Auth Routes
		auth := api.Group("/auth")
		{
			auth.POST("/otp", authHandler.SendOtp)
			auth.POST("/verify", authHandler.VerifyOTP)
		}

		// Protected Auth Routes
		authProtected := api.Group("/auth")
		authProtected.Use(middleware.AuthMiddleware(jwtService))
		{
			authProtected.POST("/register", authHandler.Register)
		}

		// User Routes
		users := api.Group("/users")
		{
			users.PUT("/profile", userHandler.UpdateProfile)
		}

		// Group Routes (Protected)
		groups := api.Group("/groups")
		groups.Use(middleware.AuthMiddleware(jwtService))
		{
			groups.POST("/", groupHandler.CreateGroup)
			groups.GET("/", groupHandler.GetUserGroups)
			groups.GET("/:id", groupHandler.GetGroupDetails)
			groups.GET("/:id/members", groupHandler.GetGroupMembers)
			groups.GET("/:id/messages", groupHandler.GetGroupMessages)
			groups.POST("/:id/messages", groupHandler.SendMessage)
		}

		// Conversation Routes (Protected)
		conversations := api.Group("/conversations")
		conversations.Use(middleware.AuthMiddleware(jwtService))
		{
			conversations.POST("/", conversationHandler.StartConversation)
			conversations.GET("/", conversationHandler.GetConversations)
			conversations.POST("/:id/messages", conversationHandler.SendMessage)
			conversations.GET("/:id/messages", conversationHandler.GetMessages)
			conversations.POST("/:id/read", conversationHandler.MarkMessagesRead)
		}
	}

	// 7. Start Server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.Info("Server listening", zap.String("port", cfg.Server.Port))

	if err := r.Run(addr); err != nil {
		logger.Fatal("Failed to start server", zap.Error(err))
	}
}
