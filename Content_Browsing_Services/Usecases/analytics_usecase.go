package usecases

import (
	"context"
	"time"

	domain "lawgen/admin-service/Domain"
)

type AnalyticsUsecase struct {
	repo domain.IAnalyticsRepository
}

func NewAnalyticsUsecase(repo domain.IAnalyticsRepository) *AnalyticsUsecase {
	return &AnalyticsUsecase{repo: repo}
}

// Log content views (already wired in your controller)
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

// OPTIONAL but recommended: Log query searches (so trends have data)
func (uc *AnalyticsUsecase) LogQuerySearch(ctx context.Context, userID, term, topic string, age int, gender string) error {
	event := &domain.AnalyticsEvent{
		EventType: "QUERY",
		UserID:    userID,
		Payload: domain.QuerySearchAnalytic{
			Term:   term,
			Topic:  topic,
			Age:    age,
			Gender: gender,
		},
		Timestamp: time.Now().Unix(),
	}
	return uc.repo.SaveEvent(ctx, event)
}

// Parse YYYY-MM-DD (inclusive) range safely.
// If end < start or either empty/invalid, return an error.
func parseDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	const layout = "2006-01-02"
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	start, err := time.Parse(layout, startDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	// end of day inclusive: add 1 day, minus 1 second
	end, err := time.Parse(layout, endDate)
	if err != nil {
		return time.Time{}, time.Time{}, err
	}
	if end.Before(start) {
		return time.Time{}, time.Time{}, ErrInvalidDateRange
	}
	// end inclusive: set to 23:59:59 local of parsed date. Safer: add 24h and subtract 1s.
	end = end.Add(24*time.Hour - time.Second)
	return start, end, nil
}

var ErrInvalidDateRange = domain.ErrNotFound // reuse or define your own package error if you prefer

// Query trends
func (uc *AnalyticsUsecase) GetQueryTrends(ctx context.Context, startDate, endDate string, limit int) (*domain.QueryTrendsResult, error) {
	if limit <= 0 {
		limit = 10
	}
	start, end, err := parseDateRange(startDate, endDate)
	if err != nil {
		return nil, err
	}
	return uc.repo.GetQueryTrends(ctx, start, end, limit)
}
