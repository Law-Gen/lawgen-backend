package controllers

import (
	"context"

	"net/http"
	"strconv"

	"time"

	domain "lawgen/admin-service/Domain"
	usecases "lawgen/admin-service/Usecases"

	"github.com/gin-gonic/gin"
)

type AnalyticsController struct {
	analyticsUsecase *usecases.AnalyticsUsecase
	contentUsecase   *usecases.ContentUsecase
}

func NewAnalyticsController(analyticsUC *usecases.AnalyticsUsecase, contentUC *usecases.ContentUsecase) *AnalyticsController {
	return &AnalyticsController{
		analyticsUsecase: analyticsUC,
		contentUsecase:   contentUC,
	}
}


func (c *AnalyticsController) ViewContentAndRedirect(ctx *gin.Context) {
    contentID := ctx.Param("id")

    // Get user info from context (populated by middleware)
    userID, _ := ctx.Get("userID")
    age, _ := ctx.Get("age")
    gender, _ := ctx.Get("gender")

    // Fetch content metadata
    content, err := c.contentUsecase.FetchContentByID(ctx.Request.Context(), contentID)
    if err != nil {
        if err == domain.ErrNotFound {
            ctx.JSON(http.StatusNotFound, gin.H{"code": "NOT_FOUND", "message": "Content not found."})
        } else {
            ctx.JSON(http.StatusInternalServerError, gin.H{"code": "SERVER_ERROR", "message": err.Error()})
        }
        return
    }

    // Log analytics asynchronously
    go func() {
        bgCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = c.analyticsUsecase.LogContentView(bgCtx, userID.(string), content.ID, content.Name, age.(int), gender.(string))
    }()

    // Return metadata
    ctx.JSON(http.StatusOK, content)
}


// Enterprise Query Trends
func (c *AnalyticsController) GetQueryTrends(ctx *gin.Context) {
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")
	limitStr := ctx.DefaultQuery("limit", "10")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"code": "INVALID_INPUT", "message": "limit must be a number"})
		return
	}

	// AuthZ check (role must be enterprise)
	// role, exists := ctx.Get("role")
	// if !exists {
	// 	ctx.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "User not logged in"})
	// 	return
	// }

	// if role != "enterprise" {
	// 	ctx.JSON(http.StatusForbidden, gin.H{"code": "ACCESS_DENIED", "message": "Not an enterprise user"})
	// 	return
	// }

	result, err := c.analyticsUsecase.GetQueryTrends(ctx.Request.Context(), startDate, endDate, limit)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"code": "SERVER_ERROR", "message": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, result)
}


