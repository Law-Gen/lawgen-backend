package domain

import "context"

type LegalEntityRepository interface {
	Save(ctx context.Context, entity *LegalEntity) (*LegalEntity, error)
	GetByID(ctx context.Context, id string) (*LegalEntity, error)
	GetAll(ctx context.Context, page, limit int, search string) (*PaginatedLegalEntityRespose, error)
	Update(ctx context.Context, entity *LegalEntity) error
	Delete(ctx context.Context, id string) error
}

type PaginatedLegalEntityRespose