package domain

import "errors"

var ErrNotFound = errors.New("entity not found")

type ContentViewPayload struct {
	ContentID string `json:"content_id" bson:"content_id"`
}

type AnalyticsEvent struct {
	ID        string      `json:"id" bson:"_id,omitempty"`
	EventType string      `json:"event_type" bson:"event_type"`
	UserID    string      `json:"user_id" bson:"user_id"`
	Payload   interface{} `json:"payload" bson:"payload"`
	Timestamp int64       `json:"timestamp" bson:"timestamp"`
}