package transport

import (
	"context"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/domain"
)

type ContentPort interface {
	Create(ctx context.Context, c domain.Content) (string, error)
	Update(ctx context.Context, id string, c domain.Content) error
	GetByID(ctx context.Context, id string) (*domain.Content, error)
	Search(ctx context.Context, query, language string, limit, offset int) ([]domain.Content, int, error)
}

type FeedbackPort interface {
	Submit(ctx context.Context, fb domain.Feedback) (string, error)
}

type LegalEntityPort interface {
	Create(ctx context.Context, e domain.LegalEntity) (string, error)
	Update(ctx context.Context, id string, e domain.LegalEntity) error
	GetByID(ctx context.Context, id string) (*domain.LegalEntity, error)
	Search(ctx context.Context, query, country string, limit, offset int) ([]domain.LegalEntity, int, error)
}

type AnalyticsPort interface {
	GetQueryTrends(ctx context.Context, timeWindow string, limit int) ([]domain.AnalyticsTrend, error)
}