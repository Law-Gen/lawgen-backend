package domain

import (
	"errors"
	"time"
)

// Plan represents a subscription plan offered to users.
type Plan struct {
	ID        string
	Name      string
	Price     int64  // Price in the smallest currency unit (e.g., cents)
	Currency  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewPlan is the constructor for creating a valid Plan.
func NewPlan(id, name, currency string, price int64) (*Plan, error) {
	if id == "" || name == "" || currency == "" {
		return nil, errors.New("plan ID, name, and currency cannot be empty")
	}
	if price <= 0 {
		return nil, errors.New("plan price must be positive")
	}

	now := time.Now().UTC()
	return &Plan{
		ID:        id,
		Name:      name,
		Price:     price,
		Currency:  currency,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// PlanRepository defines the persistence contract for Plan entities.
// This interface belongs to the domain layer.
type PlanRepository interface {
	FindAll() ([]*Plan, error)
	FindByID(id string) (*Plan, error)
}