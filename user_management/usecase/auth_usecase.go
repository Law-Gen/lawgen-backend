package usecase

import (
	"context"
	"errors"
	"fmt"
	"user_management/domain"
	"user_management/infrastructure/auth"
	"user_management/infrastructure/email"
	"user_management/utils"
	"time"
)

type AuthUsecase struct {
	userRepo     domain.UserRepository
	tokenRepo    domain.TokenRepository
	hasher       *auth.PasswordHasher
	jwt          *auth.JWT
	unactiveRepo domain.UnactiveUserRepo
	emailService *email.EmailService
	passRepo     domain.PasswordResetRepository
}

func NewAuthUsecase(
	ur domain.UserRepository,
	tr domain.TokenRepository,
	jwt *auth.JWT,
	ar domain.UnactiveUserRepo,
	es *email.EmailService,
	passRepo domain.PasswordResetRepository,
) *AuthUsecase {
	return &AuthUsecase{
		userRepo:     ur,
		tokenRepo:    tr,
		hasher:       &auth.PasswordHasher{},
		jwt:          jwt,
		unactiveRepo: ar,
		emailService: es,
		passRepo:     passRepo,
	}
}

func (uc *AuthUsecase) Register(ctx context.Context, email, fullName, password string) error {
	if _, err := uc.userRepo.FindByEmail(ctx, email); err == nil {
		return errors.New("user already exists")
	}

	if _, err := uc.unactiveRepo.FindByEmailUnactive(ctx, email); err == nil {
		return errors.New("user already exists please activate your account")
	}

	hashedPassword, err := uc.hasher.HashPassword(password)
	if err != nil {
		return err
	}

	token, expiry, err := utils.GenerateRandomToken()
	if err != nil {
		return err
	}

	user := &domain.UnactivatedUser{
		FullName:              fullName,
		Email:                 email,
		Password:              hashedPassword,
		Activated:             false,
		ActivationToken:       token,
		ActivationTokenExpiry: expiry,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := uc.unactiveRepo.CreateUnactiveUser(ctx, user); err != nil {
		return err
	}

	activationLink := "http://localhost:8080/auth/activate?token=" + user.ActivationToken + "&email=" + user.Email
	go func() {
		err := uc.emailService.SendActivationEmail(user.Email, activationLink)
		if err != nil {
			fmt.Printf("Failed to send activation email: %v\n", err)
		}
	}()

	return nil
}

func (uc *AuthUsecase) Login(ctx context.Context, email, password string) (string, string, int, *domain.User, error) {
	if _, err := uc.unactiveRepo.FindByEmailUnactive(ctx, email); err == nil {
		return "", "", 0, nil, errors.New("user not activated, please check your email for activation link")
	}

	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return "", "", 0, nil, errors.New("invalid credentials")
	}

	if !uc.hasher.CompareHashAndPassword(user.Password, password) {
		return "", "", 0, nil, errors.New("invalid credentials")
	}

	accessToken, err := uc.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", 0, nil, errors.New("failed to generate token")
	}

	refreshToken, err := uc.jwt.GenerateRefreshToken()
	if err != nil {
		return "", "", 0, nil, errors.New("failed to generate refresh token")
	}

	// Store the refresh token
	expiry := time.Now().Add(uc.jwt.RefreshExpiry)
	newRefreshTokenModel := &domain.RefreshToken{
		UserID:    user.ID,
		Token:     refreshToken,
		ExpiresAt: expiry,
	}

	if err := uc.tokenRepo.StoreRefreshToken(ctx, newRefreshTokenModel); err != nil {
		return "", "", 0, nil, errors.New("failed to store refresh token")
	}

	return accessToken, refreshToken, int(uc.jwt.AccessExpiry.Seconds()), user, nil
}

func (uc *AuthUsecase) ActivateUser(ctx context.Context, token, email string) error {
	unActiveUser, err := uc.unactiveRepo.FindByEmailUnactive(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	if unActiveUser.ActivationToken != token {
		return errors.New("invalid activation token")
	}

	if unActiveUser.ActivationTokenExpiry.Before(time.Now()) {
		return errors.New("activation token has expired")
	}

	user := &domain.User{
		FullName:  unActiveUser.FullName,
		Email:     unActiveUser.Email,
		Password:  unActiveUser.Password,
		Role:      "user",
		Activated: true,
		SubscriptionStatus: "free",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.userRepo.Create(ctx, user); err != nil {
		return errors.New("failed to create user")
	}

	if err := uc.unactiveRepo.DeleteUnactiveUser(ctx, email); err != nil {
		return errors.New("failed to delete unactivated user")
	}
	return nil
}

func (uc *AuthUsecase) ResendActivationEmail(ctx context.Context, email string) error {
	unActiveUser, err := uc.unactiveRepo.FindByEmailUnactive(ctx, email)
	if err != nil {
		return errors.New("user not found")
	}

	// Enforce a minimum time gap (e.g., 30 seconds) between resend requests
	minGap := 30 * time.Second
	if time.Since(unActiveUser.UpdatedAt) < minGap {
		return errors.New("please wait before requesting another activation email")
	}

	if unActiveUser.ActivationTokenExpiry.Before(time.Now()) {
		return errors.New("activation token has expired")
	}

	token, expiry, err := utils.GenerateRandomToken()
	if err != nil {
		return err
	}

	err = uc.unactiveRepo.UpdateActiveToken(ctx, email, token, *expiry)
	if err != nil {
		return err
	}

	activationLink := "http://localhost:8080/auth/activate?token=" + token + "&email=" + unActiveUser.Email
	go func() {
		err := uc.emailService.SendActivationEmail(unActiveUser.Email, activationLink)
		if err != nil {
			fmt.Printf("Failed to send activation email: %v\n", err)
		}
	}()

	return nil
}

func (uc *AuthUsecase) ForgotPassword(ctx context.Context, email string) error {
	user, err := uc.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return errors.New("if an account exists for this email, an otp will be sent")
	}

	otp := utils.GenerateOTP()
	expiresAt := time.Now().Add(15 * time.Minute)

	resetOTP, err := uc.passRepo.GetOTPByEmail(ctx, email)
	if err == nil && resetOTP != nil {
		// If a reset token already exists, update it
		resetOTP.OTP = otp
		resetOTP.ExpiresAt = expiresAt
		if err := uc.passRepo.UpdateOTP(ctx, resetOTP); err != nil {
			return errors.New("failed to update password reset otp")
		}
	} else {
		// If no reset token exists, create a new one
		password_otp := &domain.PasswordResetOTP{
			Email:    user.Email,
			OTP:      otp,
		ExpiresAt: expiresAt,
	}
	uc.passRepo.CreateOTP(ctx, password_otp)
	}


	go func() {
		err := uc.emailService.SendPasswordResetEmail(user.Email, otp)
		if err != nil {
			fmt.Printf("Failed to send password reset otp email: %v\n", err)
		}
	}()

	return nil
}

func (uc *AuthUsecase) ResetPassword(c context.Context, passwordResetToken, newPassword string) error {
	resetToken, err := uc.passRepo.GetByToken(c, passwordResetToken)
	if err != nil {
		return errors.New("token not found")
	}

	if time.Now().After(resetToken.ExpiresAt) {
		uc.passRepo.Delete(c, resetToken.Email) // Clean up expired token.
		return errors.New("token has expired")
	}

	user, err := uc.userRepo.FindByEmail(c, resetToken.Email)
	if err != nil {
		return errors.New("user not found")
	}

	hashedPassword, err := uc.hasher.HashPassword(newPassword)
	if err != nil {
		return err
	}

	if err := uc.userRepo.UpdateUserPassword(c, user.Email, hashedPassword); err != nil {
		return err
	}

	uc.passRepo.Delete(c, resetToken.Email) // Clean up used token.

	return nil
}

func (uc *AuthUsecase) VerifyOTP(c context.Context, email, otp string) (string, error) {
	resetOTP, err := uc.passRepo.GetOTPByEmail(c, email)
	if err != nil {
		return "", errors.New("invalid or expired otp")
	}

	if resetOTP.OTP != otp {
		return "", errors.New("invalid otp")
	}

	if time.Now().After(resetOTP.ExpiresAt) {
		uc.passRepo.DeleteOTP(c, email) // Clean up expired otp.
		return "", errors.New("otp has expired")
	}

	password_token, _ := utils.CreateResetToken(email, 1*time.Hour)
	if err := uc.passRepo.Create(c, password_token); err != nil {
		return "", err
	}

	uc.passRepo.DeleteOTP(c, email) // OTP is single-use; delete after verification.
	return password_token.Token, nil
}

func (uc *AuthUsecase) RefreshTokens(ctx context.Context, refreshToken string) (string, string, int, error) {
	// 1. Validate and delete the old refresh token
	refreshTokenModel, err := uc.tokenRepo.FindRefreshToken(ctx, refreshToken)
	if err != nil {
		return "", "", 0, errors.New("invalid refresh token")
	}

	if refreshTokenModel.ExpiresAt.Before(time.Now()) {
		return "", "", 0, errors.New("refresh token has expired")
	}

	// 2. Get user details
	user, err := uc.userRepo.FindByID(ctx, refreshTokenModel.UserID)
	if err != nil {
		return "", "", 0, errors.New("user not found")
	}

	accessTokenNew, err := uc.jwt.GenerateAccessToken(user.ID, user.Role)
	if err != nil {
		return "", "", 0, errors.New("failed to generate access token")
	}

	return refreshToken, accessTokenNew, int(uc.jwt.AccessExpiry.Seconds()), nil
}

// Logout (single device)
func (uc *AuthUsecase) Logout(ctx context.Context, refreshToken string) error {
	err := uc.tokenRepo.DeleteRefreshToken(ctx, refreshToken)
	return err
}

// LogoutAll (all devices)
func (uc *AuthUsecase) LogoutAll(ctx context.Context, userID string) error {
	return uc.tokenRepo.DeleteAllForUser(ctx, userID)
}
