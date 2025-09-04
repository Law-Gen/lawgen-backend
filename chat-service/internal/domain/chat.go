package domain

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// --- Subscription Tiers and Parameters ---

type SubscriptionTier string

const (
	TierGuest      SubscriptionTier = "visitor"
	TierFree       SubscriptionTier = "free"
	TierBasic      SubscriptionTier = "basic"
	TierPro        SubscriptionTier = "pro"
	TierEnterprise SubscriptionTier = "enterprise"
)

type UserParams struct {
	MaxAnswerWords int
	MaxReferences  int
	ContextWindow  int
	SaveHistory    bool
}

// GetUserParamsFromPlanID returns subscription-specific parameters based on the provided plan ID.
func GetUserParamsFromPlanID(planID string) UserParams {
	switch SubscriptionTier(planID) {
	case TierFree:
		return UserParams{MaxAnswerWords: 150, MaxReferences: 7, ContextWindow: 2, SaveHistory: true}
	case TierBasic:
		return UserParams{MaxAnswerWords: 250, MaxReferences: 10, ContextWindow: 3, SaveHistory: true}
	case TierPro:
		return UserParams{MaxAnswerWords: 500, MaxReferences: 10, ContextWindow: 5, SaveHistory: true}
	case TierEnterprise:
		return UserParams{MaxAnswerWords: 500, MaxReferences: 15, ContextWindow: 5, SaveHistory: true}
	case TierGuest: // Visitor
		return UserParams{MaxAnswerWords: 80, MaxReferences: 5, ContextWindow: 1, SaveHistory: false}
	default:
		return UserParams{MaxAnswerWords: 80, MaxReferences: 5, ContextWindow: 1, SaveHistory: false} // Default guest-like
	}
}

// --- Session Model ---

type Session struct {
	ID           string             `bson:"-" json:"id"`
	MongoID      primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	UserID       string             `bson:"userId,omitempty" json:"user_id,omitempty"`
	Language     string             `bson:"language" json:"language"`
	CreatedAt    time.Time          `bson:"createdAt" json:"created_at"`
	LastActiveAt time.Time          `bson:"lastActiveAt" json:"last_active_at"`
	IsGuest      bool               `bson:"isGuest" json:"is_guest"`
	Title        string             `bson:"title" json:"title"`
}

// --- Chat Entry Model ---

type ChatMessageType string

const (
	MessageTypeUser    ChatMessageType = "user_query"
	MessageTypeLLM     ChatMessageType = "llm_response"
	MessageTypeSummary ChatMessageType = "summary"
)

type ChatEntry struct {
	ID         string             `bson:"-" json:"id"`
	MongoID    primitive.ObjectID `bson:"_id,omitempty" json:"-"`
	SessionID  string             `bson:"sessionId" json:"session_id"`
	Type       ChatMessageType    `bson:"type" json:"type"`
	Content    string             `bson:"content" json:"content"`
	Sources    []RAGSource        `bson:"sources,omitempty" json:"sources,omitempty"`
	CreatedAt  time.Time          `bson:"createdAt" json:"created_at"`
	SyncedToDB bool               `bson:"syncedToDB,omitempty" json:"-"`
}

type RAGSource struct {
	Content       string   `json:"content"`
	Source        string   `json:"source"`
	ArticleNumber string   `json:"article_number"`
	Topics        []string `json:"topics,omitempty"`
}

type RAGResult struct {
	Results    []RAGSource `json:"results"`
	Message    string      `json:"message"`
	References []string    `json:"references,omitempty"`
}

// --- Repository Interfaces ---

type SessionRepository interface {
	GetSessionByID(ctx context.Context, id string) (*Session, error)
	CreateSession(ctx context.Context, session *Session) error
	UpdateSession(ctx context.Context, session *Session) error
	GetSessionsByUserID(ctx context.Context, userID string, page, limit int) ([]*Session, int, error)
	SetSessionTTL(ctx context.Context, sessionID string, ttl time.Duration) error
	// For sync job to find active user sessions
	GetUserSessionIDs(ctx context.Context) ([]string, error)
	AddUserSessionID(ctx context.Context, sessionID string) error
	RemoveUserSessionID(ctx context.Context, sessionID string) error
}

type ChatRepository interface {
	GetChatHistory(ctx context.Context, sessionID string, limit int) ([]ChatEntry, error)
	SaveChatEntry(ctx context.Context, entry *ChatEntry) error
	BulkSaveChatEntries(ctx context.Context, entries []ChatEntry) error
	GetUnsyncedChatEntries(ctx context.Context, sessionID string) ([]ChatEntry, error)
	MarkChatEntriesAsSynced(ctx context.Context, sessionID string, entryMongoIDs []string) error
}

// --- External Service Interfaces ---

type LLMStreamResponse struct {
	Chunk string
	Error error
	Done  bool
}

type LLMService interface {
	StreamGenerate(ctx context.Context, prompt string, history []ChatEntry, maxWords int) (<-chan LLMStreamResponse, error)
	Generate(ctx context.Context, prompt string, history []ChatEntry) (string, error)
	Translate(ctx context.Context, text, targetLang string) (string, error)
	Close() error
}

type RAGService interface {
	Retrieve(ctx context.Context, query string, k int) (*RAGResult, error)
	Close() error
}
