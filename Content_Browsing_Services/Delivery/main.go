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

// mongoConnect is a helper function to establish and verify a MongoDB connection.
func mongoConnect(ctx context.Context) (*mongo.Database, error) {
	// For production, load these from environment variables or a config file.
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb://localhost:27017"
	}
	dbName := "lawgen_admin_db"

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	// Ping the primary to verify that the connection is alive.
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	
	log.Println("Successfully connected to MongoDB!")
	return client.Database(dbName), nil
}

func main() {
	// --- 1. INITIALIZATION & CONFIGURATION ---

	// Set up a context with a timeout for initialization steps.
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Connect to the MongoDB database.
	// If this fails, the application cannot start.
	db, err := mongoConnect(ctx)
	if err != nil {
		log.Fatalf("Fatal Error: Could not connect to MongoDB: %v", err)
	}

	// --- 2. REPOSITORY LAYER (The "How" - Data & External Services) ---
	// This layer contains the concrete implementations for talking to databases and cloud services.
	
	// Initialize the repository for Legal Entities (uses MongoDB).
	legalEntityRepo := Repositories.NewMongoEntityRepository(db)
	
	// Initialize the repository for file storage (uses Azure Blob Storage).
	// If this fails (e.g., missing credentials), the application cannot start.
	contentStorage, err := infrastructure.NewAzureBlobStorage()
	if err != nil { 
		log.Fatalf("Fatal Error: Could not create Azure Blob storage: %v", err) 
	}
	
	// Initialize the repository for content metadata (uses MongoDB).
	contentMetadataRepo := Repositories.NewMongoContentRepository(db)

	// Initialize the repository for analytics events (uses MongoDB).
	analyticsRepo := Repositories.NewMongoAnalyticsRepository(db)

	// --- 3. USECASE LAYER (The "What" - Business Logic) ---
	// This layer orchestrates the logic, depending on the repository interfaces.

	legalEntityUsecase := usecases.NewLegalEntityUsecase(legalEntityRepo)
	contentUsecase := usecases.NewContentUsecase(contentStorage, contentMetadataRepo)
	analyticsUsecase := usecases.NewAnalyticsUsecase(analyticsRepo)

	// --- 4. CONTROLLER LAYER (The "Entrypoint" - HTTP Handling) ---
	// This layer handles HTTP requests and calls the usecases.

	legalEntityController := controllers.NewLegalEntityController(legalEntityUsecase)
	contentController := controllers.NewContentController(contentUsecase)
	analyticsController := controllers.NewAnalyticsController(analyticsUsecase, contentUsecase)

	// --- 5. ROUTER & SERVER ---
	// The router takes the controllers and maps their methods to API endpoints.

	// Ensure Gin runs in "release" mode in a production environment.
	// gin.SetMode(gin.ReleaseMode)
	router := routers.NewRouter(
		legalEntityController, 
		contentController, 
		analyticsController,
	)

	// Start the Gin HTTP server.
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting Gin server on http://localhost:%s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Fatal Error: Could not start Gin server: %s\n", err)
	}
}