package usecases

import (
	"context"
	"errors"
	domain "lawgen/admin-service/Domain"
)

type LegalEntityUsecase struct {
	repo domain.LegalEntityRepository
}

func NewLegalEntityUsecase (repo domain.LegalEntityRepository) *LegalEntityUsecase {
	return &LegalEntityUsecase{repo: repo}
}

func (uc *LegalEntityUsecase) CreateLegalEntity(ctx context.Context, entity *domain.LegalEntity) (*domain.LegalEntity, error) {
	if err := entity.IsValid(); err != nil {
		return nil, errors.New("validation failed")
	}

	return uc.repo.Save(ctx, entity)
}

func (uc *LegalEntityUsecase) FetchLegalEntityById(ctx context.Context, id string) (*domain.LegalEntity, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *LegalEntityUsecase) FetchAllLegalEntities(ctx context.Context, page, limit int, search string) (*domain.PaginatedLegalEntityRespose, error) {
	return uc.repo.GetAll(ctx, page, limit,search)
}

func (uc *LegalEntityUsecase) UpdateLegalEntity(ctx context.Context, id string, entity *domain.LegalEntity) error {
	if err := entity.IsValid(); err != nil {
		return err
	}

	return uc.repo.Update(ctx, id, entity)
}

func (uc *LegalEntityUsecase) DeleteLegalEntity(ctx context.Context, id string) error {
	return uc.repo.Delete(ctx, id)
}