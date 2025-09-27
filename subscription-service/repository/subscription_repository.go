package repository

import (
	"context"
	"errors"
	"log"
	"time"

	"subscription_service/domain"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoSubscriptionRepository struct {
	collection *mongo.Collection
}

// NewMongoSubscriptionRepository creates a new repository for Subscriptions.
func NewMongoSubscriptionRepository(db *MongoDatabase) domain.SubscriptionRepository {
	return &mongoSubscriptionRepository{
		collection: db.Client.Database(db.DBName).Collection("subscriptions"),
	}
}

func (r *mongoSubscriptionRepository) FindByUserID(userID string) (*domain.Subscription, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var subDTO SubscriptionDTO
	err := r.collection.FindOne(ctx, bson.M{"user_id": userID}).Decode(&subDTO)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // Not found
		}
		return nil, err
	}

	return subDTO.toDomain(), nil
}

func (r *mongoSubscriptionRepository) Create(sub *domain.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subDTO := fromSubscriptionDomain(sub)
	_, err := r.collection.InsertOne(ctx, subDTO)
	if err != nil {
		log.Printf("Error creating subscription for user %s: %v", sub.UserID, err)
	}
	return err
}

func (r *mongoSubscriptionRepository) Update(sub *domain.Subscription) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	subDTO := fromSubscriptionDomain(sub)
	filter := bson.M{"_id": subDTO.ID}
	update := bson.M{"$set": subDTO}

	opts := options.Update().SetUpsert(false) // Do not create a doc if it doesn't exist
	_, err := r.collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		log.Printf("Error updating subscription %s: %v", sub.ID, err)
	}
	return err
}