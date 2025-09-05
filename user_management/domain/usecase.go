package domain

import (
	"context"
	"errors"
	"io"
	"time"

	"golang.org/x/oauth2"
)


type UserUsecase interface {
	Promote(ctx context.Context,userId, email string) error
	Demote(ctx context.Context, userId, email string) error
	ProfileUpdate(ctx context.Context, userid string, gender string, birthDate time.Time, languagePreference string, file io.Reader) error
	GetAllUsers(ctx context.Context, page int, limit int) ([]User, int64, error)
	GetUserByID(ctx context.Context, id string) (*User, error)
}

type AuthUsecase interface {
	Register(ctx context.Context, email, fullName, password string) error
	Login(ctx context.Context, email, password string) (string, string, int, *User, error)
	RefreshTokens(ctx context.Context, refreshToken string) (string, string, int, error)
	Logout(ctx context.Context, refreshToken string) error
	LogoutAll(ctx context.Context, userID string) error
	ActivateUser(ctx context.Context, token, email string) error
	ResendActivationEmail(ctx context.Context, email string) error
	ForgotPassword(ctx context.Context, email string) error
	ResetPassword(ctx context.Context, passwordResetToken, newPassword string) error
	VerifyOTP(c context.Context, email, otp string) (string, error)
}

type OAuthUsecase interface {
	OAuthLogin(c context.Context, googleOauthConfig oauth2.Config, code string, CodeVerifier string) (string, string, int, *User, error)
}

var ErrUnauthorized = errors.New("unauthorized action")
var ErrInvalidpreftype = errors.New("invalid preference type")

type AIUseCase interface {
	GenerateIntialSuggestion(ctx context.Context, title string) (string, error)
	GenerateBasedOnTags(ctx context.Context, content string, tags []string) (string, error)
}

type SubscriptionUsecase interface {
	GetAllPlans(ctx context.Context) ([]SubscriptionPlan, error)
	CreateSubscription(ctx context.Context, userID, planID string) error
	CancelSubscription(ctx context.Context, userID string) error
}