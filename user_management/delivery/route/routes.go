package route

import (
	"time"
	"user_management/delivery/controller"
	"user_management/infrastructure/auth"
	"user_management/infrastructure/middleware"

	"github.com/didip/tollbooth/v7/limiter"
	// "github.com/didip/tollbooth_gin"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func AuthRouter(r *gin.Engine, authController *controller.AuthController, jwt *auth.JWT, authLimiter *limiter.Limiter) {
    authGroup := r.Group("/auth")
    // authGroup.Use(tollbooth_gin.LimitHandler(authLimiter)) // Apply rate limiting middleware
    {
        authGroup.POST("/register", authController.Register)
        authGroup.POST("/login", authController.Login)
        authGroup.GET("/activate", authController.ActivateUser)
        authGroup.POST("/resend-activation", authController.ResendActivationEmail)
        authGroup.POST("/forgot-password", authController.ForgotPassword)
        authGroup.POST("/reset-password", authController.ResetPassword)
        authGroup.POST("/refresh", authController.RefreshAccessToken)
        authGroup.POST("/verify-otp", authController.VerifyOTP)
        
        authGroup.Use(middleware.AuthMiddleware(jwt)) // Apply auth middleware
        {
            authGroup.POST("/logout", authController.Logout)      // Single device
            authGroup.POST("/logout-all", authController.LogoutAll) // All devices
        }
    }
}

func OAuthRouter(r *gin.Engine, oauthController *controller.OAuthController, authLimiter *limiter.Limiter) {
    oauthGroup := r.Group("/auth/google")
    // oauthGroup.Use(tollbooth_gin.LimitHandler(authLimiter)) // Apply rate limiting middleware
    {
        oauthGroup.POST("/", oauthController.HandleGoogleLogin)
    }
}

func UserRouter(r *gin.Engine, userController *controller.UserController, jwt *auth.JWT, contentCreationLimiter *limiter.Limiter, contentReadLimiter *limiter.Limiter) {
    userGroup := r.Group("/")
    {
        userGroup.Use(middleware.AuthMiddleware(jwt)) // Apply auth middleware
        {
            userGroup.PUT("/users/me", userController.HandleUpdateUser)
            userGroup.GET("/users/me", userController.HandleGetUserByID)
            userGroup.PUT("/users/me/change-password", userController.HandleChangePassword)
            // Admin routes
            userGroup.POST("/admin/promote", middleware.RoleMiddleware(), userController.HandlePromote)
            userGroup.POST("/admin/demote", middleware.RoleMiddleware(), userController.HandleDemote)
			userGroup.GET("/admin/users", middleware.RoleMiddleware(), userController.HandleGetAllUsers)
            userGroup.POST("/admin/deactivate", middleware.RoleMiddleware(), userController.HandleDeactivateUser)
            userGroup.POST("/admin/activate", middleware.RoleMiddleware(), userController.HandleActivateUser)
		}
	}
}

func SubscriptionRouter(r *gin.Engine, subscriptionController *controller.SubscriptionController, jwt *auth.JWT, contentCreationLimiter *limiter.Limiter, contentReadLimiter *limiter.Limiter) {
    subscriptionGroup := r.Group("/subscriptions")
    {
        {
            subscriptionGroup.GET("/plans", subscriptionController.GetAllPlans)
            subscriptionGroup.POST("/subscribe", middleware.AuthMiddleware(jwt),  subscriptionController.CreateSubscription)
            subscriptionGroup.POST("/cancel", middleware.AuthMiddleware(jwt),  subscriptionController.CancelSubscription)
        }
    }
}



// HealthRouter registers a health check endpoint
func HealthRouter(r *gin.Engine) {
    r.GET("/health", func(ctx *gin.Context) {
        ctx.JSON(200, gin.H{"status": "ok"})
    })
}


// NewRouter initializes the Gin engine and registers all routes
func NewRouter() *gin.Engine {
	r := gin.Default()
    config := cors.Config{
        AllowOrigins: []string{
            "http://localhost:3000",
            "https://lawgen-frontend-wine.vercel.app",
        },
        AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization", "X-Client-Type"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }
    r.Use(cors.New(config))
	r.LoadHTMLGlob("utils/*.html")
    HealthRouter(r) // Register health check endpoint
	return r
}
