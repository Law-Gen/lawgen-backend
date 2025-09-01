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

func (uc *AnalyticsUsecase) LogContentView(ctx context.Context, userID string, contentID string) error {
	event := &domain.AnalyticsEvent{
		EventType: "CONTENT_VIEW",
		UserID: userID,
		Payload: domain.ContentViewPayload{ContentID: contentID},
		Timestamp: time.Now().Unix(),
	}

	return uc.repo.SaveEvent(ctx, event)
}