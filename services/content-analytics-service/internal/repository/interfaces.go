package repository

import (
	"context"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/domain"
)

type ContentRepository interface {
	Create(ctx context.Context, c domain.Content) (string, error)
	Update(ctx context.Context, id string, c domain.Content) error
	GetByID(ctx context.Context, id string) (*domain.Content, error)
	Search(ctx context.Context, query string, language string, limit, offset int) ([]domain.Content, int, error)
}

type FeedbackRepository interface {
	Create(ctx context.Context, fb domain.Feedback) (string, error)
}

// Teammate-owned repositories
type LegalEntityRepository interface {
	Create(ctx context.Context, e domain.LegalEntity) (string, error)
	Update(ctx context.Context, id string, e domain.LegalEntity) error
	GetByID(ctx context.Context, id string) (*domain.LegalEntity, error)
	Search(ctx context.Context, query string, country string, limit, offset int) ([]domain.LegalEntity, int, error)
}

type AnalyticsRepository interface {
	// Placeholder: teammate will define exact methods (e.g., record events, compute trends)
	QueryTrends(ctx context.Context, timeWindow string, limit int) ([]domain.AnalyticsTrend, error)
}