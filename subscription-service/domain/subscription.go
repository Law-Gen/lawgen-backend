package domain

import (
	"errors"
	"fmt"
	"time"
)

// SubscriptionStatus defines the possible states of a user's subscription.
type SubscriptionStatus string

const (
	SubscriptionStatusActive   SubscriptionStatus = "active"
	SubscriptionStatusCanceled SubscriptionStatus = "canceled"
	SubscriptionStatusPastDue  SubscriptionStatus = "past_due"
	SubscriptionStatusInactive SubscriptionStatus = "inactive"
)

// IsValid checks if the status is one of the predefined valid types.
func (s SubscriptionStatus) IsValid() error {
	switch s {
	case SubscriptionStatusActive, SubscriptionStatusCanceled, SubscriptionStatusPastDue, SubscriptionStatusInactive:
		return nil
	}
	return fmt.Errorf("invalid subscription status: %s", s)
}

// Subscription represents a user's enrollment in a subscription plan.
type Subscription struct {
	ID                 string
	UserID             string
	PlanID             string
	Status             SubscriptionStatus
	CurrentPeriodStart time.Time
	CurrentPeriodEnd   time.Time
	CanceledAt         *time.Time // Nullable
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// NewSubscription is the constructor for creating a valid Subscription.
func NewSubscription(id, userID, planID string, status SubscriptionStatus, periodStart, periodEnd time.Time) (*Subscription, error) {
	if id == "" || userID == "" || planID == "" {
		return nil, errors.New("subscription ID, userID, and planID cannot be empty")
	}
	if err := status.IsValid(); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	return &Subscription{
		ID:                 id,
		UserID:             userID,
		PlanID:             planID,
		Status:             status,
		CurrentPeriodStart: periodStart,
		CurrentPeriodEnd:   periodEnd,
		CanceledAt:         nil,
		CreatedAt:          now,
		UpdatedAt:          now,
	}, nil
}

// Cancel marks a subscription as canceled. This is a core domain logic method.
func (s *Subscription) Cancel() {
	now := time.Now().UTC()
	s.Status = SubscriptionStatusCanceled
	s.CanceledAt = &now
	s.UpdatedAt = now
}


// --- REPOSITORY PORT ---

// SubscriptionRepository defines the persistence contract for Subscription entities.
type SubscriptionRepository interface {
	FindByUserID(userID string) (*Subscription, error)
	Create(sub *Subscription) error
	Update(sub *Subscription) error
}


// --- USECASE / APPLICATION PORT ---

// SubscriptionUsecase defines the application's business rules for managing subscriptions.
// This interface acts as a port, dictating how the delivery layer (e.g., Gin handlers)
// interacts with the application's core logic.
type SubscriptionUsecase interface {
	// ListPlans retrieves all available subscription plans.
	ListPlans() ([]*Plan, error)

	// InitiateSubscriptionCheckout creates a payment gateway session for a user to subscribe to a plan.
	InitiateSubscriptionCheckout(userID, planID string) (checkoutURL string, err error)

	// GetUserSubscription fetches the active subscription details for a specific user.
	GetUserSubscription(userID string) (*Subscription, error)

	// CancelSubscription cancels the current user's active subscription.
	CancelSubscription(userID string) error

	// HandlePaymentWebhook processes incoming events from the payment gateway.
	HandlePaymentWebhook(payload []byte, signature string) error
}