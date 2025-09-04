package Repositories

import (
	"context"
	"log"

	domain "lawgen/admin-service/Domain"

	"go.mongodb.org/mongo-driver/bson"
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

// === Enterprise Query Trends Aggregation ===
func (r *mongoAnalyticsRepository) GetQueryTrends(ctx context.Context, startDate, endDate string, limit int) (*domain.QueryTrendsResult, error) {
	// NOTE: This is a placeholder aggregation pipeline.
	// You’ll need to adapt it based on how you store query search events.

	pipeline := mongo.Pipeline{
		{{"$match", bson.D{
			{"event_type", "QUERY"},
			{"timestamp", bson.D{
				{"$gte", startDate},
				{"$lte", endDate},
			}},
		}}},
		{{"$group", bson.D{
			{"_id", "$payload.query"},
			{"count", bson.D{{"$sum", 1}}},
		}}},
		{{"$sort", bson.D{{"count", -1}}}},
		{{"$limit", limit}},
	}

	cursor, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	result := &domain.QueryTrendsResult{
		Keywords: []domain.KeywordTrend{},
		Topics:   []domain.TopicTrend{},
	}

	// For now, log the raw docs for debugging.
	for cursor.Next(ctx) {
		var doc bson.M
		if err := cursor.Decode(&doc); err != nil {
			log.Printf("Decode error: %v", err)
			continue
		}
		// TODO: map doc into KeywordTrend or TopicTrend
	}

	return result, nil
}
