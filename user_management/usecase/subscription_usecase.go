package usecase

import (
	"context"
	"user_management/domain"
)

// subscriptionUsecase implements the SubscriptionUsecase interface.
type subscriptionUsecase struct {
	planRepo domain.SubscriptionPlanRepository
	userRepo domain.UserRepository
}

// NewSubscriptionUsecase creates a new instance of the usecase.
func NewSubscriptionUsecase(planRepo domain.SubscriptionPlanRepository, userRepo domain.UserRepository) domain.SubscriptionUsecase {
	return &subscriptionUsecase{planRepo: planRepo, userRepo: userRepo}
}

// GetAllPlans retrieves all available subscription plans.
func (uc *subscriptionUsecase) GetAllPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	return uc.planRepo.FindAll(ctx)
}

// CreateSubscription handles the logic for a user subscribing to a plan.
func (uc *subscriptionUsecase) CreateSubscription(ctx context.Context, userID, planID string) error {
	// Find the plan to ensure it's valid.
	plan, err := uc.planRepo.FindByID(ctx, planID)
	if err != nil {
		return err // Plan not found or database error
	}
	// Update the user's status with the name of the retrieved plan.
	return uc.userRepo.UpdateUserSubscriptionStatus(ctx, userID, plan.Name)
}

// CancelSubscription reverts a user's subscription status to 'free'.
func (uc *subscriptionUsecase) CancelSubscription(ctx context.Context, userID string) error {
	return uc.userRepo.UpdateUserSubscriptionStatus(ctx, userID, "free")
}