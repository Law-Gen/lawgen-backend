package app

import (
	"github.com/LAWGEN/lawgen-backend/chat-service/internal/config"
	"github.com/gin-gonic/gin"
)

func RegisterQuizRoutes(router *gin.Engine, quizController *QuizController, authMiddleware gin.HandlerFunc) {
	// Public routes
	public := router.Group("/api/v1/quizzes")
	{
		public.GET("/categories", quizController.GetCategories)
		public.GET("/categories/:categoryId", quizController.GetQuizzesByCategory)
		public.GET("/:quizId", quizController.GetQuiz)
		public.GET("/:quizId/questions", quizController.GetQuestionsByQuiz)
		// public.POST("/:quizId/submit", quizController.SubmitQuiz)
	}

	// Admin routes
	admin := router.Group("/api/v1/admin/quizzes")
	admin.Use(authMiddleware)
	{
		// Category management
		admin.POST("/categories", quizController.CreateCategory)
		admin.PUT("/categories/:categoryId", quizController.UpdateCategory)
		admin.DELETE("/categories/:categoryId", quizController.DeleteCategory)

		// Quiz management
		admin.POST("/", quizController.CreateQuiz)
		admin.PUT("/:quizId", quizController.UpdateQuiz)
		admin.DELETE("/:quizId", quizController.DeleteQuiz)

		// Question management
		admin.POST("/:quizId/questions", quizController.AddQuestion)
		admin.PUT("/:quizId/questions/:questionId", quizController.UpdateQuestion)
		admin.DELETE("/:quizId/questions/:questionId", quizController.DeleteQuestion)
	}
}

func RegisterChatRoutes(router *gin.Engine, chatController *ChatController, cfg *config.Config) {
	public := router.Group("/api/v1/chats")
	{
		public.POST("/query", chatController.postQuery)
		public.GET("/sessions", chatController.listSessions)
		public.GET("/sessions/:sessionId/messages", chatController.getMessages)
		public.POST("/voice-query", VoiceChatHandlerWithConfig(chatController.chatService, cfg))
	}
}
