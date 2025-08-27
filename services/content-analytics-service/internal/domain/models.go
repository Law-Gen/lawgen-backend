package domain

import "time"

type Content struct {
	ID         string    `json:"id" bson:"id"`
	Title      string    `json:"title" bson:"title"`
	Summary    string    `json:"summary,omitempty" bson:"summary,omitempty"`
	Language   string    `json:"language,omitempty" bson:"language,omitempty"`
	Tags       []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	SourceURL  string    `json:"source_url,omitempty" bson:"sourceUrl,omitempty"`
	Deleted    bool      `json:"deleted,omitempty" bson:"deleted,omitempty"`
	CreatedAt  time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time `json:"updated_at" bson:"updated_at"`
}

type Feedback struct {
	ID          string    `json:"id" bson:"id"`
	UserID      string    `json:"user_id" bson:"user_id"`
	Type        string    `json:"type" bson:"type"`
	Description string    `json:"description" bson:"description"`
	Severity    string    `json:"severity,omitempty" bson:"severity,omitempty"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at"`
}

// Teammate-owned domain
type LegalEntity struct {
	ID        string    `json:"id" bson:"id"`
	Name      string    `json:"name" bson:"name"`
	Type      string    `json:"type" bson:"type"` // e.g., "company", "individual"
	Country   string    `json:"country" bson:"country"`
	Tags      []string  `json:"tags,omitempty" bson:"tags,omitempty"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at"`
}

type AnalyticsTrend struct {
	Topic      string `json:"topic"`
	Count      int64  `json:"count"`
	Language   string `json:"language,omitempty"`
	TimeWindow string `json:"time_window,omitempty"`
}

// Error codes
const (
	ErrInvalidInput       = "INVALID_INPUT"
	ErrMissingField       = "MISSING_FIELD"
	ErrUnauthorized       = "UNAUTHORIZED"
	ErrAccessDenied       = "ACCESS_DENIED"
	ErrNotFound           = "NOT_FOUND"
	ErrDuplicateResource  = "DUPLICATE_RESOURCE"
	ErrRateLimitExceeded  = "RATE_LIMIT_EXCEEDED"
	ErrServerError        = "SERVER_ERROR"
	ErrServiceUnavailable = "SERVICE_UNAVAILABLE"
)