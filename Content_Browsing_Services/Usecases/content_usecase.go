package usecases

import (
	"context"
	"fmt"
	"io"
	domain "lawgen/admin-service/Domain"

	"path/filepath"

	"github.com/google/uuid"
)

type ContentUsecase struct {
	storage      domain.IContentStorage
	metadataRepo domain.IContentRepository
}

func NewContentUsecase(storage domain.IContentStorage, repo domain.IContentRepository) *ContentUsecase {
	return &ContentUsecase{storage: storage, metadataRepo: repo}
}

func (uc *ContentUsecase) CreateContent(ctx context.Context, file io.Reader, originalFilename, groupName, name, description, language string) (*domain.Content, error) {
	fileExtension := filepath.Ext(originalFilename)
	uniqueFileName := fmt.Sprintf("%s%s", uuid.New().String(), fileExtension)

	fileURL, err := uc.storage.Upload(ctx, uniqueFileName, file)
	if err != nil { return nil, fmt.Errorf("failed to upload file: %w", err) }

	newContent := &domain.Content{
		GroupName:   groupName,
		Name:        name,
		Description: description,
		URL:         fileURL,
		Language:    language,
	}

	id, err := uc.metadataRepo.Save(ctx, newContent)
	if err != nil { return nil, fmt.Errorf("failed to save content metadata: %w", err) }
	
	newContent.ID = id
	return newContent, nil
}

func (uc *ContentUsecase) FetchAllContent(ctx context.Context, page, limit int, search string) (*domain.PaginatedContentResponse, error) {
	return uc.metadataRepo.GetAll(ctx, page, limit, search)
}

func (uc *ContentUsecase) FetchContentByID(ctx context.Context, id string) (*domain.Content, error) {
	return uc.metadataRepo.GetByID(ctx, id)
}