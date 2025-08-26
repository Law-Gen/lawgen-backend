package repository

import (
	"context"
	"errors"
	"time"

	"subscription_service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoPlanRepository struct {
	collection *mongo.Collection
}

// NewMongoPlanRepository creates a new repository for Plans.
func NewMongoPlanRepository(db *MongoDatabase) domain.PlanRepository {
	return &mongoPlanRepository{
		collection: db.Client.Database(db.DBName).Collection("plans"),
	}
}

func (r *mongoPlanRepository) FindByID(id string) (*domain.Plan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var planDTO PlanDTO
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&planDTO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Not found is a valid business case, not an error.
			return nil, nil
		}
		return nil, err
	}

	return planDTO.toDomain(), nil
}

func (r *mongoPlanRepository) FindAll() ([]*domain.Plan, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var plans []*domain.Plan
	for cursor.Next(ctx) {
		var planDTO PlanDTO
		if err := cursor.Decode(&planDTO); err != nil {
			return nil, err
		}
		plans = append(plans, planDTO.toDomain())
	}

	return plans, nil
}