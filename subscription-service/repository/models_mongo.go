package repository

import (
	"time"

	"subscription_service/domain"

	"go.mongodb.org/mongo-driver/mongo"
)

// MongoDatabase holds the client and database name.
type MongoDatabase struct {
	Client *mongo.Client
	DBName string
}

// --- Plan DTO and Converters ---

// PlanDTO is the data transfer object for the Plan entity, specific to MongoDB persistence.
type PlanDTO struct {
	ID        string    `bson:"_id"`
	Name      string    `bson:"name"`
	Price     int64     `bson:"price"`
	Currency  string    `bson:"currency"`
	CreatedAt time.Time `bson:"created_at"`
	UpdatedAt time.Time `bson:"updated_at"`
}

// fromPlanDomain converts a domain.Plan to a PlanDTO.
func fromPlanDomain(p *domain.Plan) *PlanDTO {
	return &PlanDTO{
		ID:        p.ID,
		Name:      p.Name,
		Price:     p.Price,
		Currency:  p.Currency,
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

// toDomain converts a PlanDTO to a domain.Plan.
func (dto *PlanDTO) toDomain() *domain.Plan {
	return &domain.Plan{
		ID:        dto.ID,
		Name:      dto.Name,
		Price:     dto.Price,
		Currency:  dto.Currency,
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

// --- Subscription DTO and Converters ---

// SubscriptionDTO is the data transfer object for the Subscription entity.
type SubscriptionDTO struct {
	ID                 string     `bson:"_id"`
	UserID             string     `bson:"user_id"`
	PlanID             string     `bson:"plan_id"`
	Status             string     `bson:"status"`
	CurrentPeriodStart time.Time  `bson:"current_period_start"`
	CurrentPeriodEnd   time.Time  `bson:"current_period_end"`
	CanceledAt         *time.Time `bson:"canceled_at,omitempty"` // omitempty makes it nullable in MongoDB
	CreatedAt          time.Time  `bson:"created_at"`
	UpdatedAt          time.Time  `bson:"updated_at"`
}

// fromSubscriptionDomain converts a domain.Subscription to a SubscriptionDTO.
func fromSubscriptionDomain(s *domain.Subscription) *SubscriptionDTO {
	return &SubscriptionDTO{
		ID:                 s.ID,
		UserID:             s.UserID,
		PlanID:             s.PlanID,
		Status:             string(s.Status),
		CurrentPeriodStart: s.CurrentPeriodStart,
		CurrentPeriodEnd:   s.CurrentPeriodEnd,
		CanceledAt:         s.CanceledAt,
		CreatedAt:          s.CreatedAt,
		UpdatedAt:          s.UpdatedAt,
	}
}

// toDomain converts a SubscriptionDTO to a domain.Subscription.
func (dto *SubscriptionDTO) toDomain() *domain.Subscription {
	return &domain.Subscription{
		ID:                 dto.ID,
		UserID:             dto.UserID,
		PlanID:             dto.PlanID,
		Status:             domain.SubscriptionStatus(dto.Status),
		CurrentPeriodStart: dto.CurrentPeriodStart,
		CurrentPeriodEnd:   dto.CurrentPeriodEnd,
		CanceledAt:         dto.CanceledAt,
		CreatedAt:          dto.CreatedAt,
		UpdatedAt:          dto.UpdatedAt,
	}
}