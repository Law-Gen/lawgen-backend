package repository

import (
	"context"
	"errors"
	"time"
	"user_management/domain"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type PasswordResetTokenResponseDTO struct {
	Token      string    `bson:"token,omitempty"`
	Email    string    `bson:"email,omitempty"`
	ExpiresAt time.Time `bson:"expires_at,omitempty"`
}

type OTPResponseDTO struct {
	OTP      string    `bson:"otp,omitempty"`
	Email    string    `bson:"email,omitempty"`
	ExpiresAt time.Time `bson:"expires_at,omitempty"`
}

type PasswordReset struct {
	collection  *mongo.Collection
	otpCollection *mongo.Collection
}

func NewPasswordReset(db *mongo.Database) *PasswordReset {
	return &PasswordReset{
		collection:  db.Collection("password_reset"),
		otpCollection: db.Collection("password_reset_otp"),
	}
}

func (pr *PasswordReset) Create(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := pr.collection.InsertOne(ctx, fromDomainPass(token))
	return err
}

func (pr *PasswordReset) CreateOTP(ctx context.Context, otp *domain.PasswordResetOTP) error {
	_, err := pr.otpCollection.InsertOne(ctx, fromDomainOTP(otp))
	return err
}

func (pr *PasswordReset) Update(ctx context.Context, token *domain.PasswordResetToken) error {
	_, err := pr.collection.UpdateOne(ctx, bson.M{"email": token.Email}, bson.M{"$set": fromDomainPass(token)})
	return err
}

func (pr *PasswordReset) UpdateOTP(ctx context.Context, otp *domain.PasswordResetOTP) error {
	_, err := pr.otpCollection.UpdateOne(ctx, bson.M{"email": otp.Email}, bson.M{"$set": fromDomainOTP(otp)})
	return err
}

func (pr *PasswordReset) GetByToken(ctx context.Context, token string) (*domain.PasswordResetToken, error) {
	var dto PasswordResetTokenResponseDTO
	filter := bson.M{"token": token}
	if err := pr.collection.FindOne(ctx, filter).Decode(&dto); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("token not found")
		}
		return nil, err
	}
	return toDomainPass(&dto), nil
}

func (pr *PasswordReset) GetOTPByEmail(ctx context.Context, email string) (*domain.PasswordResetOTP, error) {
	var dto OTPResponseDTO
	filter := bson.M{"email": email}
	if err := pr.otpCollection.FindOne(ctx, filter).Decode(&dto); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("email not found")
		}
		return nil, err
	}
	return toDomainOTP(&dto), nil
}

func (pr *PasswordReset) Delete(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	_, err := pr.collection.DeleteOne(ctx, filter)
	return err
}

func (pr *PasswordReset) DeleteOTP(ctx context.Context, email string) error {
	filter := bson.M{"email": email}
	_, err := pr.otpCollection.DeleteOne(ctx, filter)
	return err
}

func fromDomainPass(token *domain.PasswordResetToken) *PasswordResetTokenResponseDTO {
	return &PasswordResetTokenResponseDTO{
		Token:      token.Token,
		Email:     token.Email,
		ExpiresAt: token.ExpiresAt,
	}
}

func toDomainPass(dto *PasswordResetTokenResponseDTO) *domain.PasswordResetToken {
	return &domain.PasswordResetToken{
		Token:     dto.Token,
		Email:    dto.Email,
		ExpiresAt: dto.ExpiresAt,
	}
}

func fromDomainOTP(otp *domain.PasswordResetOTP) *OTPResponseDTO {
	return &OTPResponseDTO{
		OTP:      otp.OTP,
		Email:   otp.Email,
		ExpiresAt: otp.ExpiresAt,
	}
}

func toDomainOTP(dto *OTPResponseDTO) *domain.PasswordResetOTP {
	return &domain.PasswordResetOTP{
		OTP:      dto.OTP,
		Email:   dto.Email,
		ExpiresAt: dto.ExpiresAt,
	}
}
