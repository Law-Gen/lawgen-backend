package repository

import (
	"context"
	"fmt"
	"user_management/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// SubscriptionPlanRepoDTO is the DTO for MongoDB storage.
// It contains the BSON tags required by the mongo-driver.
type SubscriptionPlanRepoDTO struct {
	ID    primitive.ObjectID  `bson:"_id,omitempty"`
	Name  string  `bson:"name"`
	Price int 		`bson:"price"`
}

// ToDomain converts the repository DTO to the domain model.
func (dto *SubscriptionPlanRepoDTO) ToDomain() *domain.SubscriptionPlan {
	return &domain.SubscriptionPlan{
		ID:    dto.ID.Hex(),
		Name:  dto.Name,
		Price: dto.Price,
	}
}

// FromDomain converts a domain model to the repository DTO.
func FromDomain(plan *domain.SubscriptionPlan) *SubscriptionPlanRepoDTO {
	var objectID primitive.ObjectID
	if plan.ID != "" {
		var err error
		objectID, err = primitive.ObjectIDFromHex(plan.ID)
		if err != nil {
			objectID = primitive.NilObjectID
		}
	}
	return &SubscriptionPlanRepoDTO{
		ID:    objectID,
		Name:  plan.Name,
		Price: plan.Price,
	}
}


// mongoSubscriptionPlanRepository is the MongoDB implementation of the repository.
type SubscriptionPlanRepository struct {
	collection *mongo.Collection
}

// NewMongoSubscriptionPlanRepository creates a new instance of the repository.
func NewSubscriptionPlanRepository(db *mongo.Database) domain.SubscriptionPlanRepository {
	coll := db.Collection("subscription_plans")
	return &SubscriptionPlanRepository{collection: coll}
}

// FindAll retrieves all subscription plans from the database.
func (r *SubscriptionPlanRepository) FindAll(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	// We fetch the data into the DTO struct with BSON tags.
	var planDTOs []SubscriptionPlanRepoDTO
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.Background())

	if err = cursor.All(context.Background(), &planDTOs); err != nil {
		return nil, err
	}

	// Convert the slice of DTOs to a slice of domain models before returning.
	var plans []domain.SubscriptionPlan
	for _, dto := range planDTOs {
		plans = append(plans, *dto.ToDomain())
	}

	return plans, nil
}

// FindByID retrieves a single subscription plan by its ID.
func (r *SubscriptionPlanRepository) FindByID(ctx context.Context, id string) (*domain.SubscriptionPlan, error) {
	var planDTO SubscriptionPlanRepoDTO
	
	idObj, _ := primitive.ObjectIDFromHex(id)
	err := r.collection.FindOne(ctx, bson.M{"_id": idObj}).Decode(&planDTO)
	if err != nil {
		fmt.Println("Error finding plan by ID:", id, err)
		return nil, err
	}
	// Convert the DTO to a domain model before returning.
	return planDTO.ToDomain(), nil
}
