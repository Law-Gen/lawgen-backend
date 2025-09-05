package Repositories

import (
	"context"
	"time"

	domain "lawgen/admin-service/Domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// mongoAnalyticsRepository implements the IAnalyticsRepository interface using MongoDB.
type mongoAnalyticsRepository struct {
	collection *mongo.Collection
}

// NewMongoAnalyticsRepository creates a new instance of mongoAnalyticsRepository.
func NewMongoAnalyticsRepository(db *mongo.Database) domain.IAnalyticsRepository {
	return &mongoAnalyticsRepository{collection: db.Collection("analytics_events")}
}

// SaveEvent inserts a single analytics event into the database.
func (r *mongoAnalyticsRepository) SaveEvent(ctx context.Context, event *domain.AnalyticsEvent) error {
	_, err := r.collection.InsertOne(ctx, event)
	return err
}

// GetQueryTrends aggregates top keywords and topics with demographics over a given time range.
func (r *mongoAnalyticsRepository) GetQueryTrends(ctx context.Context, start, end time.Time, limit int) (*domain.QueryTrendsResult, error) {
	startUnix := start.Unix()
	endUnix := end.Unix()

	// Helper stage to create demographic counter fields, reused by both facets.
	addDemoCounters := bson.D{{
		Key: "$addFields", Value: bson.D{
			{"maleCount", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$payload.gender", "male"}}}, 1, 0}}}},
			{"femaleCount", bson.D{{"$cond", bson.A{bson.D{{"$eq", bson.A{"$payload.gender", "female"}}}, 1, 0}}}},
			{"age18_24", bson.D{{"$cond", bson.A{bson.D{{"$and", bson.A{bson.D{{"$gte", bson.A{"$payload.age", 18}}}, bson.D{{"$lte", bson.A{"$payload.age", 24}}}}}}, 1, 0}}}},
			{"age25_34", bson.D{{"$cond", bson.A{bson.D{{"$and", bson.A{bson.D{{"$gte", bson.A{"$payload.age", 25}}}, bson.D{{"$lte", bson.A{"$payload.age", 34}}}}}}, 1, 0}}}},
			{"age35_44", bson.D{{"$cond", bson.A{bson.D{{"$and", bson.A{bson.D{{"$gte", bson.A{"$payload.age", 35}}}, bson.D{{"$lte", bson.A{"$payload.age", 44}}}}}}, 1, 0}}}},
			{"age45_54", bson.D{{"$cond", bson.A{bson.D{{"$and", bson.A{bson.D{{"$gte", bson.A{"$payload.age", 45}}}, bson.D{{"$lte", bson.A{"$payload.age", 54}}}}}}, 1, 0}}}},
			{"age55_plus", bson.D{{"$cond", bson.A{bson.D{{"$gte", bson.A{"$payload.age", 55}}}, 1, 0}}}},
		},
	}}

	// Defines the aggregation pipeline for the "keywords" facet branch.
	keywordsBranch := bson.A{
		bson.D{{"$match", bson.D{{"payload.term", bson.D{{"$ne", ""}}}}}},
		addDemoCounters,
		bson.D{{"$group", bson.D{
			{"_id", "$payload.term"}, {"count", bson.D{{"$sum", 1}}},
			{"male", bson.D{{"$sum", "$maleCount"}}}, {"female", bson.D{{"$sum", "$femaleCount"}}},
			{"age18_24", bson.D{{"$sum", "$age18_24"}}}, {"age25_34", bson.D{{"$sum", "$age25_34"}}},
			{"age35_44", bson.D{{"$sum", "$age35_44"}}}, {"age45_54", bson.D{{"$sum", "$age45_54"}}},
			{"age55_plus", bson.D{{"$sum", "$age55_plus"}}},
		}}},
		bson.D{{"$sort", bson.D{{"count", -1}}}},
		bson.D{{"$limit", limit}},
		bson.D{{"$project", bson.D{
			{"term", "$_id"}, {"count", 1}, {"_id", 0},
			{"demographics", bson.D{
				{"male", "$male"}, {"female", "$female"},
				{"age_ranges", bson.D{
					{"18-24", "$age18_24"}, {"25-34", "$age25_34"}, {"35-44", "$age35_44"},
					{"45-54", "$age45_54"}, {"55+", "$age55_plus"},
				}},
			}},
		}}},
	}

	// Defines the aggregation pipeline for the "topics" facet branch.
	topicsBranch := bson.A{
		bson.D{{"$match", bson.D{{"payload.topic", bson.D{{"$ne", ""}}}}}},
		addDemoCounters,
		bson.D{{"$group", bson.D{
			{"_id", "$payload.topic"}, {"count", bson.D{{"$sum", 1}}},
			{"male", bson.D{{"$sum", "$maleCount"}}}, {"female", bson.D{{"$sum", "$femaleCount"}}},
			{"age18_24", bson.D{{"$sum", "$age18_24"}}}, {"age25_34", bson.D{{"$sum", "$age25_34"}}},
			{"age35_44", bson.D{{"$sum", "$age35_44"}}}, {"age45_54", bson.D{{"$sum", "$age45_54"}}},
			{"age55_plus", bson.D{{"$sum", "$age55_plus"}}},
		}}},
		bson.D{{"$sort", bson.D{{"count", -1}}}},
		bson.D{{"$limit", limit}},
		bson.D{{"$project", bson.D{
			{"name", "$_id"}, {"count", 1}, {"_id", 0},
			{"demographics", bson.D{
				{"male", "$male"}, {"female", "$female"},
				{"age_ranges", bson.D{
					{"25-34", "$age25_34"}, {"35-44", "$age35_44"}, {"18-24", "$age18_24"},
					{"45-54", "$age45_54"}, {"55+", "$age55_plus"},
				}},
			}},
		}}},
	}

	// The main aggregation pipeline.
	pipeline := mongo.Pipeline{
		bson.D{{"$match", bson.D{
			{"event_type", "QUERY"},
			{"timestamp", bson.D{{"$gte", startUnix}, {"$lte", endUnix}}},
		}}},
		bson.D{{"$facet", bson.D{
			{"keywords", keywordsBranch},
			{"topics", topicsBranch},
		}}},
	}

	cur, err := r.collection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	// Define a struct to decode the single result document from the $facet stage.
	type rawResult struct {
		Keywords []domain.KeywordTrend `bson:"keywords"`
		Topics   []domain.TopicTrend   `bson:"topics"`
	}

	var results []rawResult
	if err := cur.All(ctx, &results); err != nil {
		return nil, err
	}

	finalResult := &domain.QueryTrendsResult{
		Keywords: []domain.KeywordTrend{},
		Topics:   []domain.TopicTrend{},
	}

	if len(results) > 0 {
		finalResult.Keywords = results[0].Keywords
		finalResult.Topics = results[0].Topics
	}

	return finalResult, nil
}