package controller

import (
	"fmt"
	"user_management/config"
	"user_management/domain"
	"net/http"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// OAuthController handles Google OAuth2 authentication requests.
type OAuthController struct {
	usecase domain.OAuthUsecase 
}

// NewOAuthController creates a new instance of OAuthController.
func NewOAuthController(uc domain.OAuthUsecase) *OAuthController {
	return &OAuthController{usecase: uc}
}

// getGoogleOauthConfig initializes and returns the Google OAuth2 configuration.
func getGoogleOauthConfig() *oauth2.Config {
	return &oauth2.Config{
		RedirectURL:  "http://localhost:8080/auth/google/callback",
		ClientID:     config.AppConfig.GoogleClientID,
		ClientSecret: config.AppConfig.GoogleClientSecret,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

type OAuthRequest struct {
	Code         string `json:"code" binding:"required"`
	CodeVerifier string `json:"code_verifier" binding:"required"` // PKCE verifier
}

// HandleGoogleCallback processes the callback from Google after user authentication.
func (oc *OAuthController) HandleGoogleLogin(c *gin.Context) {

	var req OAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: missing code or code_verifier"})
		return
	}

	googleOauthConfig := getGoogleOauthConfig()

	
	// Call the OAuthLogin usecase to handle token exchange, user info retrieval,
	accessToken, refreshToken, accessExpirySeconds, user, err := oc.usecase.OAuthLogin(c.Request.Context(), *googleOauthConfig, req.Code, req.CodeVerifier)
	if err != nil {
		// Log the error for debugging purposes (optional, but recommended)
		fmt.Printf("Error during OAuthLogin: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// On successful login, return the tokens and user information
	c.JSON(http.StatusOK, gin.H{
		"refresh_token": refreshToken,
		"access_token":  accessToken,
		"expiry_in":     accessExpirySeconds, 
		"user":          user,
	})
}