package routers

import (
	"lawgen/admin-service/Delivery/controllers"
	infrastructure "lawgen/admin-service/Infrastructure"
	"lawgen/admin-service/Infrastructure/middleware"
	"time"

	"github.com/gin-gonic/gin"
 	"github.com/gin-contrib/cors"
)

// NewRouter sets up the Gin router with all routes and middleware
func NewRouter(
	legalEntityController *controllers.LegalEntityController,
	contentController *controllers.ContentController,
	analyticsController *controllers.AnalyticsController,
	feedbackController *controllers.FeedbackController,
	jwtHandler *infrastructure.JWT,
) *gin.Engine {

	router := gin.Default()
	config := cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Client-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    router.Use(cors.New(config))


	// --- PUBLIC API (no authentication required) ---
	apiV1 := router.Group("/api/v1")
	{
		// Legal Entity routes (publicly accessible)
		legalEntityAPI := apiV1.Group("/legal-entities")
		{
			legalEntityAPI.POST("", legalEntityController.CreateLegalEntity)
			legalEntityAPI.GET("/:id", legalEntityController.FetchLegalEntityById)
			legalEntityAPI.GET("", legalEntityController.FetchAllLegalEntities)
			legalEntityAPI.PUT("/:id", legalEntityController.UpdateLegalEntity)
			legalEntityAPI.DELETE("/:id", legalEntityController.DeleteLegalEntity)
		}

		// Feedback routes (publicly accessible)
		feedbackAPI := apiV1.Group("/feedback")
		{
			feedbackAPI.POST("", feedbackController.CreateFeedback)
			feedbackAPI.GET("/:id", feedbackController.GetFeedbackByID)
			feedbackAPI.GET("", feedbackController.ListFeedbacks)
		}

		// Content routes (some require authentication for analytics)
		contentsAPI := apiV1.Group("/contents")
		{
			contentsAPI.GET("", contentController.GetAllContent)
			contentsAPI.GET("/:id/view", middleware.AuthMiddleware(jwtHandler), analyticsController.ViewContentAndRedirect)
			contentsAPI.GET("/group/:groupID", contentController.GetContentsByGroupID)
		}

		// Enterprise analytics (requires enterprise plan)
		enterpriseAPI := apiV1.Group("/enterprise/analytics")
		{
			enterpriseAPI.Use(middleware.AuthMiddleware(jwtHandler))     // must be logged in
			enterpriseAPI.Use(middleware.EnterprisePlanMiddleware())   // must have enterprise plan
			enterpriseAPI.GET("/query-trends", analyticsController.GetQueryTrends)
		}
	}

	// --- ADMIN API (requires admin role) ---
	adminV1 := router.Group("/api/v1/admin")
	{
		adminV1.Use(middleware.AuthMiddleware(jwtHandler))   // must be logged in
		adminV1.Use(middleware.RoleMiddleware("admin"))      // must have admin role

		// Admin content management
		adminContentAPI := adminV1.Group("/contents")
		{
			adminContentAPI.POST("", contentController.CreateContent)
			adminContentAPI.DELETE("/:id", contentController.DeleteContent)
		}

		// You can add more admin-only routes here
	}

	return router
}
