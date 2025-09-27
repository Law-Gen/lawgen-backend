package repository

import (
	"context"
	"errors"
	"user_management/domain"
	"user_management/infrastructure"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// UnactivatedUserDTO represents a user who has not yet activated their account
type UnactivatedUserDTO struct {
	ID                    primitive.ObjectID `bson:"_id,omitempty"`
	FullName              string             `bson:"full_name"`
	Email                 string             `bson:"email"`
	Password              string             `bson:"password"`
	Activated             bool               `bson:"activated"`
	ActivationToken       string             `bson:"activation_token,omitempty"`
	ActivationTokenExpiry *time.Time         `bson:"activation_token_expiry,omitempty"`
	CreatedAt             time.Time          `bson:"created_at"`
	UpdatedAt             time.Time          `bson:"updated_at"`
}

// ConvertToDomain converts UnactivatedUserDTO to domain.UnactivatedUser
func (dto *UnactivatedUserDTO) ConvertToUnactivatedUserDomain() *domain.UnactivatedUser {
	return &domain.UnactivatedUser{
		ID:                    dto.ID.Hex(),
		FullName:              dto.FullName,
		Email:                 dto.Email,
		Password:              dto.Password,
		Activated:             dto.Activated,
		ActivationToken:       dto.ActivationToken,
		ActivationTokenExpiry: dto.ActivationTokenExpiry,
		CreatedAt:             dto.CreatedAt,
		UpdatedAt:             dto.UpdatedAt,
	}
}

// ConvertToDTO converts domain.UnactivatedUser to UnactivatedUserDTO
func ConvertToUnactivatedUserDTO(u *domain.UnactivatedUser) *UnactivatedUserDTO {
	userID, _ := primitive.ObjectIDFromHex(u.ID)
	return &UnactivatedUserDTO{
		ID:                    userID,
		FullName:              u.FullName,
		Email:                 u.Email,
		Password:              u.Password,
		Activated:             u.Activated,
		ActivationToken:       u.ActivationToken,
		ActivationTokenExpiry: u.ActivationTokenExpiry,
		CreatedAt:             u.CreatedAt,
		UpdatedAt:             u.UpdatedAt,
	}
}

type UnactiveUserRepo struct {
	collection *mongo.Collection
}

func NewUnactiveUserRepo(db *mongo.Database) domain.UnactiveUserRepo {
	coll := db.Collection("unactivated_users")

	index := mongo.IndexModel{
		Keys:    bson.M{"activation_token_expiry": 1},
		Options: options.Index().SetExpireAfterSeconds(0),
	}

	if _, err := coll.Indexes().CreateOne(context.Background(), index); err != nil {
		infrastructure.Log.Fatalf("Failed to create TTL index: %v", err)
	}

	return &UnactiveUserRepo{
		collection: coll,
	}
}

func (at *UnactiveUserRepo) CreateUnactiveUser(ctx context.Context, user *domain.UnactivatedUser) error {
	_, err := at.collection.InsertOne(ctx, ConvertToUnactivatedUserDTO(user))
	return err
}

func (at *UnactiveUserRepo) FindByEmailUnactive(ctx context.Context, email string) (*domain.UnactivatedUser, error) {
	var user UnactivatedUserDTO
	filter := bson.M{"email": email}
	err := at.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user.ConvertToUnactivatedUserDomain(), nil
}

func (at *UnactiveUserRepo) DeleteUnactiveUser(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	_, err := at.collection.DeleteOne(ctx, filter)
	return err
}

func (at *UnactiveUserRepo) UpdateActiveToken(ctx context.Context, email, token string, expiry time.Time) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"activation_token":         token,
			"activation_token_expiry":  expiry,
			"updated_at":               time.Now(),
		},
	}
	_, err := at.collection.UpdateOne(ctx, filter, update)
	return err
}
