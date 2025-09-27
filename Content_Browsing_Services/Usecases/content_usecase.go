package usecases

import (
	"context"
	"fmt"
	"io"
	domain "lawgen/admin-service/Domain"

	"path/filepath"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
    if err != nil {
        return nil, fmt.Errorf("failed to upload file: %w", err)
    }

    var groupID primitive.ObjectID
    existingContent, err := uc.metadataRepo.FindOneByGroupName(ctx, groupName)
    if err != nil && err != domain.ErrNotFound {
        return nil, fmt.Errorf("failed to check for existing group: %w", err)
    }

    if existingContent != nil {
        groupID = existingContent.GroupID
    } else {
        groupID = primitive.NewObjectID()
    }

    newContent := &domain.Content{
        GroupID:     groupID,
        GroupName:   groupName,
        Name:        name,
        Description: description,
        URL:         fileURL,
        Language:    language,
    }

    id, err := uc.metadataRepo.Save(ctx, newContent)
    if err != nil {
        return nil, fmt.Errorf("failed to save content metadata: %w", err)
    }

    newContent.ID = id
    return newContent, nil
}



func (uc *ContentUsecase) FetchAllGroups(ctx context.Context, page, limit int) (*domain.PaginatedGroupResponse, error) {
	return uc.metadataRepo.GetAll(ctx, page, limit)
}

func (uc *ContentUsecase) FetchContentByID(ctx context.Context, id string) (*domain.Content, error) {
	return uc.metadataRepo.GetByID(ctx, id)
}

func (uc *ContentUsecase) GetContentsByGroupID(ctx context.Context, groupID string, page, limit int) (*domain.PaginatedContentResponse, error) {
	return uc.metadataRepo.GetContentByGroup(ctx, groupID, page, limit)
}


func (uc *ContentUsecase) DeleteContent(ctx context.Context, id string) error {
    // 1. Get content (so we know the file URL)
    content, err := uc.metadataRepo.GetByID(ctx, id)
    if err != nil {
        return err
    }

    // 2. Delete from MongoDB
    if err := uc.metadataRepo.Delete(ctx, id); err != nil {
        return err
    }

    // 3. Delete from Azure Blob Storage
    if err := uc.storage.Delete(ctx, content.URL); err != nil {
        return err
    }

    return nil
}

