package domain

import "context"

type LegalEntityRepository interface {
	Save(ctx context.Context, entity *LegalEntity) (*LegalEntity, error)
	GetByID(ctx context.Context, id string) (*LegalEntity, error)
	GetAll(ctx context.Context, page, limit int, search string) (*PaginatedLegalEntityRespose, error)
	Update(ctx context.Context, id string, entity *LegalEntity) error  
 	Delete(ctx context.Context, id string) error
}

type PaginatedLegalEntityRespose struct {
	Items    []LegalEntity  `json:"items"`
	TotalItems int          `json:"total_items"`
	TotalPages int          `json:"total_pages"`
	CurrentPage int         `json:"current_page"`
	PageSize    int         `json:"page_size"`
}