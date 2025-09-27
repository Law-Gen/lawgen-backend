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

// UserDTO represents the user data transfer object
type UserDTO struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	FullName  string             `bson:"full_name"`
	Email     string             `bson:"email"`
	Password  string             `bson:"password"`
	Role      string             `bson:"role"`
	Activated bool               `bson:"activated"`
	Profile   UserProfileDTO     `bson:"profile"`
	SubscriptionStatus string     `bson:"subscription_status"`
	CreatedAt time.Time          `bson:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"`
}

// UserProfileDTO represents the user profile data transfer object
type UserProfileDTO struct {
	Gender               string `bson:"gender,omitempty"`
	ProfilePictureURL    string `bson:"profile_picture_url,omitempty"`
	LanguagePreference   string `bson:"language_preference,omitempty"`
	BirthDate           time.Time `bson:"birth_date,omitempty"`
}

// ConvertToDomain converts UserDTO to domain.User
func (dto *UserDTO) ConvertToUserDomain() *domain.User {
	return &domain.User{
		ID:        dto.ID.Hex(),
		FullName:  dto.FullName,
		Email:     dto.Email,
		Password:  dto.Password,
		Role:      dto.Role,
		Activated: dto.Activated,
		SubscriptionStatus: dto.SubscriptionStatus,
		Profile: domain.UserProfile{
			Gender:               dto.Profile.Gender,
			ProfilePictureURL: dto.Profile.ProfilePictureURL,
			BirthDate:           dto.Profile.BirthDate,
			LanguagePreference:  dto.Profile.LanguagePreference,
		},
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

// ConvertToDTO converts domain.User to UserDTO
func ConvertToUserDTO(u *domain.User) *UserDTO {
	id, _ := primitive.ObjectIDFromHex(u.ID)

	return &UserDTO{
		ID:        id,
		FullName:  u.FullName,
		Email:     u.Email,
		Password:  u.Password,
		Role:      u.Role,
		Activated: u.Activated,
		SubscriptionStatus: u.SubscriptionStatus,
		Profile: UserProfileDTO{
			Gender:               u.Profile.Gender,
			ProfilePictureURL:    u.Profile.ProfilePictureURL,
			LanguagePreference:   u.Profile.LanguagePreference,
			BirthDate:           u.Profile.BirthDate,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func ConvertDTOSlicetoDomian(users []UserDTO) []domain.User {
	domainUsers := make([]domain.User, len(users))
	for i, dto := range users {
		domainUsers[i] = *dto.ConvertToUserDomain()
	}
	return domainUsers
}

type UserRepository struct {
	collection *mongo.Collection
}

func NewUserRepository(db *mongo.Database) domain.UserRepository {
	coll := db.Collection("users")
	index := mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	}

	if _, err := coll.Indexes().CreateOne(context.Background(), index); err != nil {
		infrastructure.Log.Fatalf("Failed to create index: %v", err)
	}

	return &UserRepository{
		collection: coll,
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	_, err := r.collection.InsertOne(ctx, ConvertToUserDTO(user))
	return err
}

func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	var user UserDTO
	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, errors.New("user not found")
	}
	return user.ConvertToUserDomain(), err
}

func (r *UserRepository) FindByID(ctx context.Context, id string) (*domain.User, error) {
	var user UserDTO
	idObj, _ := primitive.ObjectIDFromHex(id)
	err := r.collection.FindOne(ctx, bson.M{"_id": idObj}).Decode(&user)
	return user.ConvertToUserDomain(), err
}

func (mr *UserRepository) UpdateUserProfile(ctx context.Context, gender string, birthDate time.Time, languagePreference string, imagePath string, Email string) error {
	filter := bson.M{"email": Email}
	update := bson.M{
		"$set": bson.M{
			"profile": bson.M{
				"gender":              gender,
				"birth_date":          birthDate,
				"language_preference": languagePreference,
				"profile_picture_url": imagePath,
			},
		},
	}

	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (mr *UserRepository) UpdateUserRole(ctx context.Context, role string, Email string) error {
	filter := bson.M{"email": Email}
	update := bson.M{
		"$set": bson.M{
			"role": role,
		},
	}

	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (mr *UserRepository) UpdateActiveStatus(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"activated": true,
		},
	}
	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (mr *UserRepository) DeactivateUser(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"activated": false,
		},
	}
	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (mr *UserRepository) UpdateUserPassword(ctx context.Context, email string, newPasswordHash string) error {
	filter := bson.M{"email": email}
	update := bson.M{
		"$set": bson.M{
			"password": newPasswordHash,
		},
	}

	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (mr *UserRepository) UpdateUserSubscriptionStatus(ctx context.Context, userID string, newStatus string) error {
	idObj, _ := primitive.ObjectIDFromHex(userID)
	filter := bson.M{"_id": idObj}
	update := bson.M{
		"$set": bson.M{
			"subscription_status": newStatus,
		},
	}

	if res, err := mr.collection.UpdateOne(ctx, filter, update); err != nil {
		return err
	} else {
		if res.MatchedCount == 0 {
			return errors.New("user not found")
		}
		return nil
	}
}

func (ur *UserRepository) GetAllUsers(ctx context.Context, page, limit int) ([]domain.User, int64, error) {
	setskip := int64((page - 1) * limit)
	setlimit := int64(limit)

	opts := options.Find().SetSkip(setskip).SetLimit(setlimit)

	cursor, err := ur.collection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var users []UserDTO
	if err := cursor.All(ctx, &users); err != nil {
		return nil, 0, err
	}

	total, _ := ur.collection.CountDocuments(ctx, bson.M{})

	return ConvertDTOSlicetoDomian(users), total, nil
}
