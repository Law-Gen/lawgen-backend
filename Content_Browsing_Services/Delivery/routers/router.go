package routers

import (
	"lawgen/admin-service/Delivery/controllers"

	"github.com/gin-gonic/gin"
)

func NewRouter(
	legalEntityController *controllers.LegalEntityController,
	contentController *controllers.ContentController,
	analyticsController *controllers.AnalyticsController,
	FeedbackController *controllers.FeedbackController,
) *gin.Engine {
	
	router := gin.Default()

	// --- Public API Group ---
	apiV1 := router.Group("/api/v1")
	{
		// Legal Entity Routes
		legalEntityAPI := apiV1.Group("/legal-entities")
		{
			legalEntityAPI.POST("", legalEntityController.CreateLegalEntity)
			legalEntityAPI.GET("/:id", legalEntityController.FetchLegalEntityById)
			legalEntityAPI.GET("", legalEntityController.FetchAllLegalEntities)
			legalEntityAPI.PUT("/:id", legalEntityController.UpdateLegalEntity)
			legalEntityAPI.DELETE("/:id", legalEntityController.DeleteLegalEntity)
		}

		feedbackAPI := apiV1.Group("/feedback")
        {
            feedbackAPI.POST("", FeedbackController.CreateFeedback)
            feedbackAPI.GET("/:id", FeedbackController.GetFeedbackByID)
            feedbackAPI.GET("", FeedbackController.ListFeedbacks)
        }

		// Public Content & Analytics Routes
		contentAPI := apiV1.Group("/contents")
		{
			contentAPI.GET("", contentController.GetAllContent)
			contentAPI.GET("/:id/view", analyticsController.ViewContentAndRedirect)
		}
	}

	// --- Admin API Group (for protected routes) ---
	adminV1 := router.Group("/api/v1/admin")
	// This is where you would apply your admin-only JWT middleware
	// adminV1.Use(Infrastructure.GinAdminAuthMiddleware())
	{
		// Admin Content Management Routes
		adminContentAPI := adminV1.Group("/contents")
		{
			adminContentAPI.POST("", contentController.CreateContent)
			// You would add admin-only GET, PUT, DELETE for content here
		}
	}

	return router
}

