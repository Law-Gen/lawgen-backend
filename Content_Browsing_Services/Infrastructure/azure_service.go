package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	domain "lawgen/admin-service/Domain"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
)

type azureBlobStorage struct {
	client        *azblob.Client
	accountName   string
	containerName string
}

func NewAzureBlobStorage() (domain.IContentStorage, error) {
	connectionString := os.Getenv("AZURE_STORAGE_CONNECTION_STRING")
	if connectionString == "" {
		connectionString = "DefaultEndpointsProtocol=https;AccountName=lawgenstorage;AccountKey=aVzZlzugZBbdznZTH2E/Ia3syRGDq+AmAhp5ZfYpWXkTrH4wNP5XsYKEGzW9QFsZGGxR/Hs8GOhQ+AStZ6R56Q==;EndpointSuffix=core.windows.net"
	}
	accountName := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	if accountName == "" {
		accountName = "lawgenstorage"
	}
	containerName := "lawgen-pdfs"

	if connectionString == "" || accountName == "" {
		return nil, errors.New("AZURE_STORAGE_CONNECTION_STRING and AZURE_STORAGE_ACCOUNT_NAME must be set")
	}

	// Create a client from connection string
	client, err := azblob.NewClientFromConnectionString(connectionString, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create Azure Blob client: %w", err)
	}

	log.Println("Successfully connected to Azure Blob Storage.")

	return &azureBlobStorage{
		client:        client,
		accountName:   accountName,
		containerName: containerName,
	}, nil
}

func (s *azureBlobStorage) Upload(ctx context.Context, fileKey string, file io.Reader) (string, error) {
	_, err := s.client.UploadStream(ctx, s.containerName, fileKey, file, nil)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to Azure Blob Storage: %w", err)
	}

	url := fmt.Sprintf("https://%s.blob.core.windows.net/%s/%s", s.accountName, s.containerName, fileKey)
	log.Printf("Successfully uploaded file to %s", url)
	return url, nil
}

func (s *azureBlobStorage) Delete(ctx context.Context, fileKey string) error {
	_, err := s.client.DeleteBlob(ctx, s.containerName, fileKey, nil)
	if err != nil {
		return fmt.Errorf("failed to delete blob '%s': %w", fileKey, err)
	}
	log.Printf("Successfully deleted blob %s", fileKey)
	return nil
}
