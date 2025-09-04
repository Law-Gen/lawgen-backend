package controller

import (
	"user_management/domain"
	"user_management/infrastructure/auth"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// UserDTO represents the user data transfer object
type UserDTO struct {
	ID        string         `json:"id"`
	FullName  string         `json:"full_name" binding:"required,min=3,max=50"`
	Email     string         `json:"email" binding:"required,email"`
	Role      string         `json:"role"`
	Profile   UserProfileDTO `json:"profile"`
	SubscriptionStatus string  `json:"subscription_status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// UserProfileDTO represents the user profile data transfer object
type UserProfileDTO struct {
	Gender               string `json:"gender" oneof:"male,female"`
	ProfilePictureURL string `json:"profile_picture_url"`
	BirthDate           time.Time `json:"birth_date"`
	LanguagePreference  string `json:"language_preference"`
}


// ConvertToDomain converts UserDTO to domain.User
func (dto *UserDTO) ConvertToUserDomain() *domain.User {
	return &domain.User{
		ID:        dto.ID,
		FullName:  dto.FullName,
		Email:     dto.Email,
		Role:      dto.Role,
		Profile: domain.UserProfile{
			Gender:            dto.Profile.Gender,
			ProfilePictureURL: dto.Profile.ProfilePictureURL,
			BirthDate:        dto.Profile.BirthDate,
			LanguagePreference:  dto.Profile.LanguagePreference,
		},
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}



// ConvertToDTO converts domain.User to UserDTO
func ConvertToUserDTO(u *domain.User) *UserDTO {
	return &UserDTO{
		ID:        u.ID,
		FullName:  u.FullName,
		Email:     u.Email,
		Role:      u.Role,
		Profile: UserProfileDTO{
			Gender:            u.Profile.Gender,
			ProfilePictureURL: u.Profile.ProfilePictureURL,
			BirthDate:        u.Profile.BirthDate,
			LanguagePreference:  u.Profile.LanguagePreference,
		},
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}


// UserCreateRequest represents registration payload (DTO)
type UserCreateRequest struct {
	FullName string `json:"full_name" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

// UserLoginRequest represents login payload (DTO)
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type EmailReq struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp_code" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type PasswordResetRequest struct {
	Token         string `json:"reset_token" binding:"required"`
	NewPassword   string `json:"new_password" binding:"required,min=8"`
}

type AuthController struct {
	authUsecase domain.AuthUsecase
	jwt         *auth.JWT
}

func NewAuthController(uc domain.AuthUsecase, jwt *auth.JWT) *AuthController {
	return &AuthController{authUsecase: uc, jwt: jwt}
}

func (c *AuthController) Register(ctx *gin.Context) {
	var req UserCreateRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := c.authUsecase.Register(ctx, req.Email, req.FullName, req.Password)
	if err != nil {
		ctx.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusCreated, gin.H{"message": "user registered successfully please check your email to activate your account"})
}

func (c *AuthController) Login(ctx *gin.Context) {
	var req UserLoginRequest

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	accessToken, refreshToken, _, user, err := c.authUsecase.Login(ctx, req.Email, req.Password)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"user":          ConvertToUserDTO(user),
	})
}

func (c *AuthController) ActivateUser(ctx *gin.Context) {
	token := ctx.Query("token")
	if token == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "activation token is required"})
		return
	}

	email := ctx.Query("email")
	if email == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "email is required"})
		return
	}

	err := c.authUsecase.ActivateUser(ctx.Request.Context(), token, email)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx.HTML(http.StatusOK, "activation_success.html", gin.H{
		"email": email,
	})

}

func (ac *AuthController) ResendActivationEmail(ctx *gin.Context) {
	var req EmailReq

	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ac.authUsecase.ResendActivationEmail(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "activation email resent successfully"})
}

func (ac *AuthController) ForgotPassword(ctx *gin.Context) {
	var req EmailReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ac.authUsecase.ForgotPassword(ctx.Request.Context(), req.Email)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{"message": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"message": "reset otp is sent to email successfully"})
}

func (ac *AuthController) ResetPassword(c *gin.Context) {
	var request PasswordResetRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := ac.authUsecase.ResetPassword(c.Request.Context(), request.Token, request.NewPassword)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "password has been reset successfully"})
}

func (ac *AuthController) VerifyOTP(c *gin.Context) {
	var request VerifyOTPRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ac.authUsecase.VerifyOTP(c.Request.Context(), request.Email, request.OTP)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "OTP verified successfully", "password_reset_token": token})
}

// Refresh Tokens
func (c *AuthController) RefreshAccessToken(ctx *gin.Context) {
	// Accept refresh token from either header or JSON body
	var req RefreshTokenRequest

	refreshToken := ctx.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
			return
		}
		refreshToken = req.RefreshToken
	}

	accessToken, refreshTokenNew, expiresIn, err := c.authUsecase.RefreshTokens(ctx.Request.Context(), refreshToken)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"access_token":  accessToken,
		"refresh_token": refreshTokenNew,
		"expires_in":    expiresIn,
	})
}

// Logout (single device)
func (c *AuthController) Logout(ctx *gin.Context) {
	var req RefreshTokenRequest

	refreshToken := ctx.GetHeader("X-Refresh-Token")
	if refreshToken == "" {
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "refresh token is required"})
			return
		}
		refreshToken = req.RefreshToken
	}

	if err := c.authUsecase.Logout(ctx.Request.Context(), refreshToken); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "successfully logged out"})
}

// LogoutAll (all devices)
func (c *AuthController) LogoutAll(ctx *gin.Context) {
	userID, exists := ctx.Get("user_id")
	if !exists {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	objID := userID.(string)
	if objID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "user ID is required"})
		return
	}

	if err := c.authUsecase.LogoutAll(ctx.Request.Context(), objID); err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "logout failed"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{"message": "logged out from all devices"})
}
