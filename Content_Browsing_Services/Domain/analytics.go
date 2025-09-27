package domain

import (
	"errors"
)

var ErrNotFound = errors.New("entity not found")



// ContentViewAnalytic is the payload for a "CONTENT_VIEW" event.
type ContentViewAnalytic struct {
	ContentID string `bson:"content_id" json:"content_id"`
	Keyword   string `bson:"keyword" json:"keyword"`
	Age       int    `bson:"age" json:"age"`
	Gender    string `bson:"gender" json:"gender"`
}

// QuerySearchAnalytic is the payload for a "QUERY" event.
type QuerySearchAnalytic struct {
	Term   string `json:"term" bson:"term"`
	Topic  string `json:"topic" bson:"topic"`
	Age    int    `json:"age" bson:"age"`
	Gender string `json:"gender" bson:"gender"`
}

// AnalyticsEvent is the generic container for any analytic event stored in the database.
type AnalyticsEvent struct {
	ID        string      `json:"id" bson:"_id,omitempty"`
	EventType string      `json:"event_type" bson:"event_type"` // e.g., "CONTENT_VIEW", "QUERY"
	UserID    string      `json:"user_id" bson:"user_id"`
	Payload   interface{} `json:"payload" bson:"payload"` // Holds structs like ContentViewAnalytic or QuerySearchAnalytic
	Timestamp int64       `json:"timestamp" bson:"timestamp"` // Unix seconds
}

// === Query Trends API Shapes ===

// Demographics represents the breakdown of an audience by gender and age.
type Demographics struct {
	Male      int            `json:"male" bson:"male"`
	Female    int            `json:"female" bson:"female"`
	AgeRanges map[string]int `json:"age_ranges" bson:"age_ranges"`
}

// KeywordTrend represents the aggregated trend data for a single search term.
type KeywordTrend struct {
	Term         string       `json:"term" bson:"term"`
	Count        int          `json:"count" bson:"count"`
	Demographics Demographics `json:"demographics" bson:"demographics"`
}

// TopicTrend represents the aggregated trend data for a single topic.
type TopicTrend struct {
	Name         string       `json:"name" bson:"name"`
	Count        int          `json:"count" bson:"count"`
	Demographics Demographics `json:"demographics" bson:"demographics"`
}

// QueryTrendsResult is the top-level object returned by the GetQueryTrends use case.
type QueryTrendsResult struct {
	Keywords []KeywordTrend `json:"keywords" bson:"keywords"`
	Topics   []TopicTrend   `json:"topics" bson:"topics"`
}


