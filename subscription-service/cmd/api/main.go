package main

import (
	"subscription_service/config"
	"subscription_service/infrastructure"
	// "subscription_service/delivery/controller"
	"subscription_service/delivery/route"
	// "subscription_service/infrastructure"
	// "subscription_service/infrastructure/ai"
	// "subscription_service/infrastructure/auth"
	// "subscription_service/infrastructure/cache"
	// "subscription_service/infrastructure/database"
	// "subscription_service/infrastructure/email"
	// "subscription_service/infrastructure/image"
	// "subscription_service/repository"
	// "subscription_service/usecase"
	// "time"
	// "github.com/didip/tollbooth/v7"
	// "github.com/didip/tollbooth/v7/limiter"
)

func main() {
	// Initialize configuration
	infrastructure.InitLogger()
	config.LoadConfig()
	// dbName := config.AppConfig.DbName

	// Initialize MongoDB connection
	// db := database.InitMongoDB().Database(dbName)
	// blogCollection := db.Collection("blogs")

	// Initialize repository, usecase, controller for blogs
	// repoCacheService := cache.NewInMemoryCache(5*time.Minute, 10*time.Minute)
	// blogRepo := repository.NewBlogRepository(blogCollection, repoCacheService)
	// blogUsecase := usecase.NewBlogUsecase(blogRepo, authRepo)
	// blogController := controller.NewBlogController(blogUsecase)



    // Initialize router
    r := route.NewRouter()

	// Start the server on port 8080
	if err := r.Run("localhost:8080"); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
