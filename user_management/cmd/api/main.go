package main

import (
	"user_management/config"
	"user_management/delivery/controller"
	"user_management/delivery/route"
	"user_management/infrastructure"
	"user_management/infrastructure/auth"
	"user_management/infrastructure/database"
	"user_management/infrastructure/email"
	"user_management/infrastructure/image"
	"user_management/repository"
	"user_management/usecase"
	"time"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
)

func main() {
	// Initialize configuration
	infrastructure.InitLogger()
	config.LoadConfig()
	dbName := config.AppConfig.DbName
	accessSecret := config.AppConfig.AccessTokenSecret
	refreshSecret := config.AppConfig.RefreshTokenSecret
	accessExpiry := config.AppConfig.AccessTokenExpiry
	refreshExpiry := config.AppConfig.RefreshTokenExpiry
	url := config.AppConfig.URL
	env := config.AppConfig.ENV

	// Initialize MongoDB connection
	db := database.InitMongoDB().Database(dbName)
	// Initialize repository, usecase, controller for authentication
	authRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	unActiveUserRepo := repository.NewUnactiveUserRepo(db)
	passwordResetRepo := repository.NewPasswordReset(db)
	emailService := email.NewEmailService()
	jwt := auth.NewJWT(accessSecret, refreshSecret, accessExpiry, refreshExpiry)
	authUsecase := usecase.NewAuthUsecase(authRepo, tokenRepo, jwt, unActiveUserRepo, emailService, passwordResetRepo, url, env)
	authController := controller.NewAuthController(authUsecase, jwt)


	subRepo := repository.NewSubscriptionPlanRepository(db)
	subUsecase := usecase.NewSubscriptionUsecase(subRepo, authRepo)
	subscriptionController := controller.NewSubscriptionController(subUsecase)


	// Initialize OAuth usecase and controller
	oauthUsecase := usecase.NewOAuthUsecase(authRepo, tokenRepo, jwt)
	oauthController := controller.NewOAuthController(oauthUsecase)

	// Initialize repository, usecase, controller for user management
	imageUpload := image.NewCloudinaryService()
	userRepo := repository.NewUserRepository(db)
	userUsecase := usecase.NewUserUsecase(userRepo, imageUpload)
	userController := controller.NewUserController(userUsecase)


    // Initialize router
    r := route.NewRouter()
	contentCreationLimiter := tollbooth.NewLimiter(0.5, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Hour})
	contentReadLimiter := tollbooth.NewLimiter(1, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Second})

	// Register authentication routes
	authLimiter := tollbooth.NewLimiter(0.16, &limiter.ExpirableOptions{DefaultExpirationTTL: time.Minute})
	route.AuthRouter(r, authController, jwt, authLimiter)

	// Register OAuth routes
	route.OAuthRouter(r, oauthController, authLimiter)

	// user management routes
	route.UserRouter(r, userController, jwt, contentCreationLimiter, contentReadLimiter)

	route.SubscriptionRouter(r, subscriptionController, jwt, contentCreationLimiter, contentReadLimiter)


	// Start the server on the configured port
	if err := r.Run(":" + config.AppConfig.PORT); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
