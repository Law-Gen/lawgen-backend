package main

import (
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/repository"
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase"

	// "github.com/LAWGEN/lawgen-backend/chat-service/internal/usecase/client"
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/LAWGEN/lawgen-backend/chat-service/internal/app"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
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

	// Initialize repositories
	// chatRepo := repository.NewChatRepository(db)
	quizRepo := repository.NewQuizRepository(db)

	// Initialize clients
	// llmClient, err := client.NewLLMClient(cfg)
	// if err != nil {
	// 	log.Fatalf("Failed to create LLM client: %v", err)
	// }
	// Use the mock RAG client instead of the gRPC one
	// ragClient, err := client.NewMockRAGClient()
	// if err != nil {
	// 	log.Fatalf("Failed to create Mock RAG client: %v", err)
	// }
	// defer ragClient.Close()

	// Initialize use cases
	// chatUseCase := usecase.NewChatUseCase(chatRepo)
	quizUseCase := usecase.NewQuizUseCase(quizRepo)

	// Initialize controllers
	// chatController := app.NewChatController(chatUseCase, ragClient, llmClient)
	quizController := app.NewQuizController(quizUseCase)

	// Setup router
	router := gin.Default()
	router.Use(UserContextMiddleware())

	// Register routes
	app.RegisterRoutes(router, quizController, AdminAuthMiddleware())

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
