package controllers

import (
	"context"
	domain "lawgen/admin-service/Domain"
	usecases "lawgen/admin-service/Usecases"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	analyticsUsecase *usecases.AnalyticsUsecase
	contentUsecase   *usecases.ContentUsecase // Needs content usecase to get the URL
}

func NewAnalyticsController(analyticsUC *usecases.AnalyticsUsecase, contentUC *usecases.ContentUsecase) *AnalyticsController {
	return &AnalyticsController{
		analyticsUsecase: analyticsUC,
		contentUsecase:   contentUC,
	}
}

// ViewContentAndRedirect logs an event and redirects the user to the PDF.
// GET /api/v1/contents/:id/view
func (c *AnalyticsController) ViewContentAndRedirect(ctx *gin.Context) {
	contentID := ctx.Param("id")

	userIDVal, exists := ctx.Get("userID")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"message": "User ID not found in context"})
		return
	}

	userID := userIDVal.(string)

	// 1. Fetch the content metadata to get the S3 URL.
	content, err := c.contentUsecase.FetchContentByID(ctx.Request.Context(), contentID)
	if err != nil {
		if err == domain.ErrNotFound {
			ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Content not found."})
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"code": "SERVER_ERROR", "message": err.Error()})
		}
		return
	}

	// 2. Log the analytics event in the background (as a goroutine)
	// so it doesn't slow down the redirect for the user.
	go func() {
		err := c.analyticsUsecase.LogContentView(context.Background(), userID, contentID)
		if err != nil {
			// This should be logged to your monitoring system (e.g., Sentry, Datadog)
			log.Printf("ERROR: Failed to log content view analytic: %v", err)
		}
	}()

	// 3. Redirect the user's browser to the actual file.
	ctx.Redirect(http.StatusFound, content.URL)
}