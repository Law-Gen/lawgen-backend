package domain

import (
	"errors"
)

var ErrNotFound = errors.New("entity not found")

// === Base Analytics Event ===
type ContentViewAnalytic struct {
    ContentID string `bson:"content_id" json:"content_id"`
    Keyword   string `bson:"keyword" json:"keyword"`
    Age       int    `bson:"age" json:"age"`
    Gender    string `bson:"gender" json:"gender"`
}



type AnalyticsEvent struct {
	ID        string      `json:"id" bson:"_id,omitempty"`
	EventType string      `json:"event_type" bson:"event_type"`
	UserID    string      `json:"user_id" bson:"user_id"`
	Payload   interface{} `json:"payload" bson:"payload"`
	Timestamp int64       `json:"timestamp" bson:"timestamp"`
}

// === Enterprise Query Trends ===
type Demographics struct {
	Male      int            `json:"male" bson:"male"`
	Female    int            `json:"female" bson:"female"`
	AgeRanges map[string]int `json:"age_ranges" bson:"age_ranges"`
}

type KeywordTrend struct {
	Term        string       `json:"term" bson:"term"`
	Count       int          `json:"count" bson:"count"`
	Demographics Demographics `json:"demographics" bson:"demographics"`
}

type TopicTrend struct {
	Name        string       `json:"name" bson:"name"`
	Count       int          `json:"count" bson:"count"`
	Demographics Demographics `json:"demographics" bson:"demographics"`
}

type QueryTrendsResult struct {
	Keywords []KeywordTrend `json:"keywords"`
	Topics   []TopicTrend   `json:"topics"`
}

type QueryPayload struct {
    Term        string       `json:"term" bson:"term"`
    Topic       string       `json:"topic" bson:"topic"`
    Demographics Demographics `json:"demographics" bson:"demographics"`
}


