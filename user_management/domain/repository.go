package domain

import (
	"context"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	FindByEmail(ctx context.Context, email string) (*User, error)
	FindByID(ctx context.Context, id string) (*User, error)
	UpdateUserProfile(ctx context.Context, gender string, birthDate time.Time, languagePreference string, imagePath string, Email string) error
	UpdateUserRole(ctx context.Context, role string, Email string) error
	UpdateActiveStatus(ctx context.Context, email string) error
	UpdateUserPassword(ctx context.Context, email string, newPasswordHash string) error
	GetAllUsers(ctx context.Context, page int, limit int) ([]User, int64, error)
	UpdateUserSubscriptionStatus(ctx context.Context, userID string, newStatus string) error
}

type UnactiveUserRepo interface {
	CreateUnactiveUser(ctx context.Context, user *UnactivatedUser) error
	FindByEmailUnactive(ctx context.Context, email string) (*UnactivatedUser, error)
	DeleteUnactiveUser(ctx context.Context, email string) error
	UpdateActiveToken(ctx context.Context, email string, token string, expiry time.Time) error
}

type PasswordResetRepository interface {
	CreateOTP(ctx context.Context, otp *PasswordResetOTP) error
	UpdateOTP(ctx context.Context, otp *PasswordResetOTP) error
	Create(ctx context.Context, token *PasswordResetToken) error
	Update(ctx context.Context, token *PasswordResetToken) error
	GetByToken(ctx context.Context, token string) (*PasswordResetToken, error)
	GetOTPByEmail(ctx context.Context, email string) (*PasswordResetOTP, error)
	Delete(ctx context.Context, email string) error
	DeleteOTP(ctx context.Context, email string) error
}

type TokenRepository interface {
	StoreRefreshToken(ctx context.Context, accessToken *RefreshToken) error
	FindRefreshToken(ctx context.Context, token string) (*RefreshToken, error)
	DeleteRefreshToken(ctx context.Context, token string) (error)
	DeleteAllForUser(ctx context.Context, userID string) error
}

type SubscriptionPlanRepository interface {
	FindAll(ctx context.Context) ([]SubscriptionPlan, error)
	FindByID(ctx context.Context, id string) (*SubscriptionPlan, error)
}

