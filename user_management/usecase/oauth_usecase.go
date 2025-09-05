package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"user_management/domain"
	"user_management/infrastructure/auth"
	"io"
	"time"
	"golang.org/x/oauth2"
)

// GoogleUserInfo represents the structure of the user information returned by Google's OAuth2 API.
type GoogleUserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	VerifiedEmail bool   `json:"verified_email"`
	Name          string `json:"name"`
	GivenName     string `json:"given_name"`
	FamilyName    string `json:"family_name"`
	Picture       string `json:"picture"` 
	Locale        string `json:"locale"`
}

// OAuthUsecase implements the business logic for OAuth authentication.
type OAuthUsecase struct {
	userRepo    domain.UserRepository
	tokenRepo   domain.TokenRepository
	jwtService  *auth.JWT
}

// NewOAuthUsecase creates a new instance of OAuthUsecase.
func NewOAuthUsecase(
	userRepo domain.UserRepository,
	tokenRepo domain.TokenRepository,
	jwtService *auth.JWT,
) domain.OAuthUsecase {
	return &OAuthUsecase{
		userRepo:    userRepo,
		tokenRepo:   tokenRepo,
		jwtService:  jwtService,
	}
}

func (uc *OAuthUsecase) OAuthLogin(
	ctx context.Context,
	googleOauthConfig oauth2.Config,
	code string,
	CodeVerifier string,
) (string, string, int, *domain.User, error) {
	// Use the PKCE code_verifier in the token exchange.
	pkceOption := oauth2.SetAuthURLParam("code_verifier", CodeVerifier)

	// Exchange the authorization code for tokens, including the PKCE verifier.
	token, err := googleOauthConfig.Exchange(ctx, code, pkceOption)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to exchange authorization code for token: %w", err)
	}

	// Use the obtained token to create an HTTP client
	client := googleOauthConfig.Client(ctx, token)

	// Fetch user information from Google's userinfo endpoint
	response, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to get user info from Google: %w", err)
	}
	defer response.Body.Close()

	// Read the response body
	data, err := io.ReadAll(response.Body)
	if err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to read Google user info response body: %w", err)
	}

	// Unmarshal the JSON response into GoogleUserInfo struct
	var googleUserInfo GoogleUserInfo
	if err := json.Unmarshal(data, &googleUserInfo); err != nil {
		return "", "", 0, nil, fmt.Errorf("failed to unmarshal Google user info: %w", err)
	}

	// --- User Lookup or Creation ---
	existingUser, _ := uc.userRepo.FindByEmail(ctx, googleUserInfo.Email)
	
	if existingUser != nil {
		age := int(time.Since(existingUser.Profile.BirthDate).Hours() / 24 / 365)
		// User exists, generate tokens for them
		accessToken, err := uc.jwtService.GenerateAccessToken(existingUser.ID, existingUser.Role, existingUser.SubscriptionStatus, existingUser.Profile.Gender, age)
		if err != nil {
			return "", "", 0, nil, fmt.Errorf("failed to generate access token: %w", err)
		}
		refreshToken, err := uc.jwtService.GenerateRefreshToken()
		if err != nil {
			return "", "", 0, nil, fmt.Errorf("failed to generate refresh token: %w", err)
		}

		// Store the refresh token in the database
		expiry := time.Now().Add(uc.jwtService.RefreshExpiry)
		newRefreshToken := &domain.RefreshToken{
			UserID:    existingUser.ID,
			Token:     refreshToken,
			ExpiresAt: expiry,
		}

		if err := uc.tokenRepo.StoreRefreshToken(ctx, newRefreshToken); err != nil {
			return "", "", 0, nil, fmt.Errorf("failed to store refresh token: %w", err)
		}
		return accessToken, refreshToken, int(uc.jwtService.AccessExpiry.Seconds()), existingUser, nil
	} else {
		// User does not exist, create a new user
		user := &domain.User{
			ID:        googleUserInfo.ID,
			FullName:  googleUserInfo.Name,
			Email:     googleUserInfo.Email,
			Password:  "", 
			Role:      "user",
			Activated: true,
			SubscriptionStatus: "free",
			Profile: domain.UserProfile{
				ProfilePictureURL: googleUserInfo.Picture,
				Gender:            "", 
				BirthDate:        time.Time{}, 
				LanguagePreference:  "", 
			},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := uc.userRepo.Create(ctx, user); err != nil {
			return "", "", 0, nil, errors.New("failed to create new user")
		}
		

		// --- Generate Application-Specific Tokens ---
		// Generate access token
		age := int(time.Since(user.Profile.BirthDate).Hours() / 24 / 365)
		accessToken, err := uc.jwtService.GenerateAccessToken(user.ID, user.Role, user.SubscriptionStatus, user.Profile.Gender, age)
		if err != nil {
			return "", "", 0, nil, errors.New("failed to generate access token")
		}

		// Generate refresh token
		refreshToken, err := uc.jwtService.GenerateRefreshToken()
		if err != nil {
			return "", "", 0, nil, errors.New("failed to generate refresh token")
		}

		// Store the refresh token in the database
		expiry := time.Now().Add(uc.jwtService.RefreshExpiry)
		newRefreshToken := &domain.RefreshToken{
			UserID:    user.ID,
			Token:     refreshToken,
			ExpiresAt: expiry,
		}

		if err := uc.tokenRepo.StoreRefreshToken(ctx, newRefreshToken); err != nil {
			return "", "", 0, nil, errors.New("failed to store refresh token")
		}

		return accessToken, refreshToken, int(uc.jwtService.AccessExpiry.Seconds()), user, nil
	}
}