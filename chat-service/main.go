package main

import (
	client "github.com/LAWGEN/lawgen-backend/chat-service/internal/client"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/domain"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/repository"
	mongoRepo "github.com/LAWGEN/lawgen-backend/chat-service/internal/repository/mongo"
	redisRepo "github.com/LAWGEN/lawgen-backend/chat-service/internal/repository/redis"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase"

	// "github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase/client"
	"context"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
	"github.com/gin-contrib/cors"
)

// UserContextMiddleware sets user information from headers for simulation.
func UserContextMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetHeader("X-User-ID")
		planID := c.GetHeader("X-Plan-ID")
		role := c.GetHeader("X-User-Role")

		if userID != "" {
			c.Set("userID", userID)
		}
		// Default to 'free' for logged-in users if no plan is specified
		if planID != "" {
			c.Set("plan_id", planID)
		} else if userID != "" {
			c.Set("plan_id", "free")
		}
		if role != "" {
			c.Set("userRole", role)
		}
		c.Next()
	}
}

// AdminAuthMiddleware checks for admin role.
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("userRole")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized: Admin access required"})
			return
		}
		c.Next()
	}
}

func main() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := repository.DBConn(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize Redis
	u, _ := url.Parse(cfg.RedisAddr)

	password, _ := u.User.Password()
	rdb := redis.NewClient(&redis.Options{
		Addr:     u.Host,
		Password: password,
		DB:       0,
	})
	defer func() {
		if err := rdb.Close(); err != nil {
			log.Printf("Error closing Redis client: %v", err)
		}
	}()
	_, err = rdb.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("Failed to ping Redis: %v", err)
	}
	log.Println("Connected to Redis")

	// Initialize repositories
	mongoSessionRepo := mongoRepo.NewSessionRepository(db)
	mongoChatRepo := mongoRepo.NewChatRepository(db)

	redisSessionRepo := redisRepo.NewRedisSessionRepository(rdb, cfg)
	redisChatRepo := redisRepo.NewRedisChatRepository(rdb, cfg)

	quizRepo := repository.NewQuizRepository(db)

	// Initialize clients
	llmClient, err := client.NewLLMClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create LLM client: %v", err)
	}

	var ragClient domain.RAGService
	ragClient, err = client.NewRAGClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize RAG client: %v", err)
	}
	defer ragClient.Close()

	// Initialize use cases
	chatUseCase := usecase.NewChatService(cfg, redisSessionRepo, redisChatRepo, mongoSessionRepo, mongoChatRepo, llmClient, ragClient)
	quizUseCase := usecase.NewQuizUseCase(quizRepo)

	// Initialize controllers
	quizController := app.NewQuizController(quizUseCase)
	chatController := app.NewChatController(chatUseCase)

	// setup middleware
	jwt := NewJWT(cfg.AccessSecret)
	
	// Setup router
	router := gin.Default()
	config := cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
			"https://lawgen.vercel.app"
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Client-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    router.Use(cors.New(config))

	router.StaticFile("/", "./index.html")
	router.Use(AuthMiddleware(*jwt))


	// Register routes
	app.RegisterQuizRoutes(router, quizController, RoleMiddleware())
	app.RegisterChatRoutes(router, chatController, cfg)

	// Start server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: router,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
