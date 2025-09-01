package Repositories

import (
	"context"
	domain "lawgen/admin-service/Domain"

	"go.mongodb.org/mongo-driver/mongo"
)

type mongoAnalyticsRepository struct {
	collection *mongo.Collection
}

func NewMongoAnalyticsRepository(db *mongo.Database) domain.IAnalyticsRepository {
	return &mongoAnalyticsRepository{collection: db.Collection("analytics_events")}
}

func (r *mongoAnalyticsRepository) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}
