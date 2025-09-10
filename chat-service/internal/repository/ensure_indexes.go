package repository

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// EnsureIndexes creates necessary indexes for optimal performance.
func EnsureIndexes(db *mongo.Database) error {
	ctx := context.Background()
	_, err := db.Collection("quiz_categories").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "name", Value: 1}}})
	if err != nil {
		return err
	}

	_, err = db.Collection("quizzes").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "category_id", Value: 1}}})
	if err != nil {
		return err
	}

	_, err = db.Collection("quizzes").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "questions._id", Value: 1}}})
	if err != nil {
		return err
	}

	_, err = db.Collection("sessions").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "userId", Value: 1}}})
	if err != nil {
		return err
	}

	_, err = db.Collection("sessions").Indexes().CreateOne(ctx,
		mongo.IndexModel{Keys: bson.D{{Key: "lastActiveAt", Value: -1}}})
	return err
}
