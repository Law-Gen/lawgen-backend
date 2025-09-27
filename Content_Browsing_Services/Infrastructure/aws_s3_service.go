package infrastructure

import (
	"context"
	"errors"
	"fmt"
	"io"
	domain "lawgen/admin-service/Domain"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type awsS3Storage struct {
    client     *s3.Client
    bucketName string
    region     string
}

// NewAwsS3Storage creates a new AWS S3 storage service client.
// It reads credentials and configuration from environment variables.
func NewAwsS3Storage() (domain.IContentStorage, error) {
    bucketName := os.Getenv("AWS_S3_BUCKET")
    if bucketName == "" {
        return nil, errors.New("AWS_S3_BUCKET environment variable must be set")
    }

    // Load the default AWS configuration (reads credentials from env vars, etc.)
    cfg, err := config.LoadDefaultConfig(context.TODO())
    if err != nil {
        return nil, fmt.Errorf("failed to load AWS config: %w", err)
    }

    if cfg.Region == "" {
        return nil, errors.New("AWS_REGION environment variable must be set")
    }

    client := s3.NewFromConfig(cfg)

    log.Println("Successfully connected to AWS S3.")

    return &awsS3Storage{
        client:     client,
        bucketName: bucketName,
        region:     cfg.Region,
    }, nil
}

func (s *awsS3Storage) Upload(ctx context.Context, fileKey string, file io.Reader) (string, error) {
    _, err := s.client.PutObject(ctx, &s3.PutObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(fileKey),
        Body:   file,
    })
    if err != nil {
        return "", fmt.Errorf("failed to upload file to AWS S3: %w", err)
    }

    // Construct the object URL
    url := fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucketName, s.region, fileKey)
    log.Printf("Successfully uploaded file to %s", url)
    return url, nil
}

func (s *awsS3Storage) Delete(ctx context.Context, fileKey string) error {
    _, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
        Bucket: aws.String(s.bucketName),
        Key:    aws.String(fileKey),
    })
    if err != nil {
        return fmt.Errorf("failed to delete object '%s' from S3: %w", fileKey, err)
    }
    log.Printf("Successfully deleted object %s from S3", fileKey)
    return nil
}