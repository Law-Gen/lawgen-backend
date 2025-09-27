package controller

import (
	"net/http"
	"user_management/domain"

	"github.com/gin-gonic/gin"
)

// SubscriptionPlanControllerDTO is the DTO for JSON responses.
// It contains the 'json' tags for serialization.
type SubscriptionPlanControllerDTO struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Price int     `json:"price"`
}

// fromDomain converts a single domain model to the controller DTO for JSON response.
func fromDomain(plan domain.SubscriptionPlan) SubscriptionPlanControllerDTO {
	return SubscriptionPlanControllerDTO{
		ID:    plan.ID,
		Name:  plan.Name,
		Price: plan.Price,
	}
}

// fromDomainSlice converts a slice of domain models to a slice of controller DTOs.
func fromDomainSlice(plans []domain.SubscriptionPlan) []SubscriptionPlanControllerDTO {
	planDTOs := make([]SubscriptionPlanControllerDTO, len(plans))
	for i, p := range plans {
		planDTOs[i] = fromDomain(p)
	}
	return planDTOs
}


// SubscriptionController holds the usecase and handles Gin routes.
type SubscriptionController struct {
	subscriptionUsecase domain.SubscriptionUsecase
}

// NewSubscriptionController creates a new instance of the controller.
func NewSubscriptionController(subscriptionUsecase domain.SubscriptionUsecase) *SubscriptionController {
	return &SubscriptionController{subscriptionUsecase: subscriptionUsecase}
}

// GetAllPlans handles the GET request to return all subscription plans.
func (c *SubscriptionController) GetAllPlans(ctx *gin.Context) {
	// The usecase layer returns pure domain models.
	plans, err := c.subscriptionUsecase.GetAllPlans(ctx)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve subscription plans"})
		return
	}

	// Convert domain models to controller DTOs for the JSON response.
	planDTOs := fromDomainSlice(plans)
	ctx.JSON(http.StatusOK, planDTOs)
}

// CreateSubscription handles the POST request to subscribe a user to a plan.
func (c *SubscriptionController) CreateSubscription(ctx *gin.Context) {
	// User ID is assumed to be set in the context by an auth middleware.
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var requestBody struct {
		PlanID string `json:"plan_id"`
	}

	if err := ctx.ShouldBindJSON(&requestBody); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body, 'plan_id' is required"})
		return
	}

	err := c.subscriptionUsecase.CreateSubscription(ctx, userID.(string), requestBody.PlanID)
	if err != nil {
		// This could be a 404 if plan not found, or 500 for other DB errors.
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Subscription created successfully"})
}

// CancelSubscription handles the POST request to cancel a user's subscription.
func (c *SubscriptionController) CancelSubscription(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	if err := c.subscriptionUsecase.CancelSubscription(ctx, userID.(string)); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel subscription"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "Subscription canceled successfully"})
}