package usecases

import (
	"context"
	domain "lawgen/admin-service/Domain"
	"time"
)

type AnalyticsUsecase struct {
	repo domain.IAnalyticsRepository
}

func NewAnalyticsUsecase(repo domain.IAnalyticsRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo}
}

// Log content views 
func (uc *AnalyticsUsecase) LogContentView(ctx context.Context, userID, contentID, keyword string, age int, gender string) error {
    event := &domain.AnalyticsEvent{
        EventType: "CONTENT_VIEW",
        UserID:    userID,
        Payload: domain.ContentViewAnalytic{
            ContentID: contentID,
            Keyword:   keyword,
            Age:       age,
            Gender:    gender,
        },
        Timestamp: time.Now().Unix(),
    }

    return uc.repo.SaveEvent(ctx, event)
}


// === New: Query trends ===
func (uc *AnalyticsUsecase) GetQueryTrends(ctx context.Context, startDate, endDate string, limit int) (*domain.QueryTrendsResult, error) {
	if limit <= 0 {
		limit = 10
	}
	return uc.repo.GetQueryTrends(ctx, startDate, endDate, limit)
}

