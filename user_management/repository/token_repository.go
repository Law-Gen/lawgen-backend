package repository

import (
	"context"
	"user_management/domain"
	"user_management/infrastructure"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// RefreshTokenDTO is a Data Transfer Object for Refresh Tokens
type RefreshTokenDTO struct {
	UserID    primitive.ObjectID `bson:"user_id"`
	Token     string             `bson:"token"`
	ExpiresAt time.Time          `bson:"expires_at"`
}

// ConvertToDomain converts RefreshTokenDTO to domain.RefreshToken
func (dto *RefreshTokenDTO) ConvertToDomain() *domain.RefreshToken {
	return &domain.RefreshToken{
		UserID:    dto.UserID.Hex(),
		Token:     dto.Token,
		ExpiresAt: dto.ExpiresAt,
	}
}

// ConvertToDTO converts domain.RefreshToken to RefreshTokenDTO
func ConvertToDTO(token *domain.RefreshToken) *RefreshTokenDTO {
	userID, _ := primitive.ObjectIDFromHex(token.UserID)
	return &RefreshTokenDTO{
		UserID:    userID,
		Token:     token.Token,
		ExpiresAt: token.ExpiresAt,
	}
}

type UserID struct {
	ID primitive.ObjectID `bson:"user_id"`
}

type TokenRepository struct {
	collection *mongo.Collection
}

func NewTokenRepository(db *mongo.Database) *TokenRepository {
	coll := db.Collection("refresh_tokens")
	index := mongo.IndexModel{
		Keys:    bson.M{"activation_token_expiry": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	if _, err := coll.Indexes().CreateOne(context.Background(), index); err != nil {
		infrastructure.Log.Fatalf("Failed to create TTL index: %v", err)
	}
	return &TokenRepository{
		collection: coll,
	}
}

func (r *TokenRepository) StoreRefreshToken(ctx context.Context, refreshToken *domain.RefreshToken) error {
	_, err := r.collection.InsertOne(ctx, ConvertToDTO(refreshToken))
	return err
}

// For single device logout
func (r *TokenRepository) FindRefreshToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	var result RefreshTokenDTO
	err := r.collection.FindOne(ctx, bson.M{"token": token}).Decode(&result)
	return result.ConvertToDomain(), err
}

func (r *TokenRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"token": token})
	return err
}

// For multiple device logout
func (r *TokenRepository) DeleteAllForUser(ctx context.Context, userID string) error {
	userIDObj, _ := primitive.ObjectIDFromHex(userID)
	_, err := r.collection.DeleteMany(ctx, bson.M{"user_id": userIDObj})
	return err
}
