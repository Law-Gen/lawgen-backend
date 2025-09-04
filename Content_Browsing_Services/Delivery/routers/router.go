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

	// --- Public API ---
	apiV1 := router.Group("/api/v1")
	{
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

		contentsAPI := apiV1.Group("/contents")
		{
			contentsAPI.GET("", contentController.GetAllContent)
			contentsAPI.GET("/:id/view", analyticsController.ViewContentAndRedirect)
			contentsAPI.GET("/group/:groupID", contentController.GetContentsByGroupID)
		}

	
		enterpriseAPI := apiV1.Group("/enterprise/analytics")
		{
				enterpriseAPI.GET("/query-trends", analyticsController.GetQueryTrends)
		}

	}


	adminV1 := router.Group("/api/v1/admin")
	{
		adminContentAPI := adminV1.Group("/contents")
		{
			adminContentAPI.POST("", contentController.CreateContent)
			adminContentAPI.DELETE("/:id", contentController.DeleteContent)
		}
	}

	return router
}
