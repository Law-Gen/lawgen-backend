package routers

import (
	"lawgen/admin-service/Delivery/controllers"

	"github.com/gin-gonic/gin"
)

func NewRouter(LegalEntityController *controllers.LegalEntityController) *gin.Engine {
	router := gin.Default()

	apiV1 := router.Group("/api/v1") 
		{
			legalEntityAPI := apiV1.Group("/legal-entities")
				{
					legalEntityAPI.POST("", LegalEntityController.CreateLegalEntity)
					legalEntityAPI.GET("/:id", LegalEntityController.FetchLegalEntityById)
					legalEntityAPI.GET("", LegalEntityController.FetchAllLegalEntities)
					legalEntityAPI.PUT("/:id", LegalEntityController.UpdateLegalEntity)
					legalEntityAPI.DELETE("/:id", LegalEntityController.DeleteLegalEntity)
				}
		}
	
	return router
}