package main

import (
	"context"
	"log"
	"os"
	"time"

	"lawgen/admin-service/Delivery/controllers"
	"lawgen/admin-service/Delivery/routers"
	infrastructure "lawgen/admin-service/Infrastructure"
	"lawgen/admin-service/Repositories"
	usecases "lawgen/admin-service/Usecases"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// mongoConnect establishes and verifies a MongoDB connection.
func mongoConnect(ctx context.Context) (*mongo.Database, error) {
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := "lawgen_admin_db"

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB!")
	return client.Database(dbName), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	db, err := mongoConnect(ctx)
	if err != nil {
		log.Fatalf("Fatal Error: Could not connect to MongoDB: %v", err)
	}

	// --- Initialize Repositories ---
	legalEntityRepo := Repositories.NewMongoEntityRepository(db)
	contentMetadataRepo := Repositories.NewMongoContentRepository(db)
	analyticsRepo := Repositories.NewMongoAnalyticsRepository(db)
	feedbackRepo := Repositories.NewMongoFeedbackRepository(db)

	// --- Initialize Usecases ---
	legalEntityUsecase := usecases.NewLegalEntityUsecase(legalEntityRepo)
	contentUsecase := usecases.NewContentUsecase(nil, contentMetadataRepo) // storage optional for now
	analyticsUsecase := usecases.NewAnalyticsUsecase(analyticsRepo)
	feedbackUsecase := usecases.NewFeedbackUsecase(feedbackRepo)

	// --- Initialize Controllers ---
	legalEntityController := controllers.NewLegalEntityController(legalEntityUsecase)
	contentController := controllers.NewContentController(contentUsecase)
	analyticsController := controllers.NewAnalyticsController(analyticsUsecase, contentUsecase)
	feedbackController := controllers.NewFeedbackController(feedbackUsecase)

	// --- JWT Handler (only validate access token from auth service) ---
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	if accessSecret == "" {
		accessSecret = "your_access_token_secret"
	}
	jwtHandler := &infrastructure.JWT{
		AccessSecret: accessSecret,
	}

	// --- Initialize Router ---
	router := routers.NewRouter(
		legalEntityController,
		contentController,
		analyticsController,
		feedbackController,
		jwtHandler,
	)

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Gin server on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Fatal Error: Could not start Gin server: %v", err)
	}
}
