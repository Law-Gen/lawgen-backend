package domain

import (
	"context"
	"io"
)

type ILegalEntityRepository interface {
	Save(ctx context.Context, entity *LegalEntity) (*LegalEntity, error)
	GetByID(ctx context.Context, id string) (*LegalEntity, error)
	GetAll(ctx context.Context, page, limit int, search string) (*PaginatedLegalEntityResponse, error)
	Update(ctx context.Context, id string, entity *LegalEntity) error
	Delete(ctx context.Context, id string) error
}

type PaginatedLegalEntityResponse struct {
	Items       []LegalEntity `json:"items"`
	TotalItems  int           `json:"total_items"`
	TotalPages  int           `json:"total_pages"`
	CurrentPage int           `json:"current_page"`
	PageSize    int           `json:"page_size"`
}

type IAnalyticsRepository interface {
	SaveEvent(ctx context.Context, event *AnalyticsEvent) error
}

type IContentStorage interface {
	Upload(ctx context.Context, fileKey string, file io.Reader) (url string, err error)
	Delete(ctx context.Context, fileKey string) error
}

type IContentRepository interface {
	Save(ctx context.Context, content *Content) (string, error)
	GetByID(ctx context.Context, id string) (*Content, error)
	GetAll(ctx context.Context, page, limit int, search string) (*PaginatedContentResponse, error)
}

type FeedbackRepository interface {
    Create(feedback *Feedback) error
    GetByID(id string) (*Feedback, error)
    List() ([]Feedback, error)
}

type FeedbackUsecase interface {
    CreateFeedback(feedback *Feedback) error
    GetFeedbackByID(id string) (*Feedback, error)
    ListFeedbacks() ([]Feedback, error)
}
