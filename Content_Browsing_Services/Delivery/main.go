package main

import (
	"context"
	"lawgen/admin-service/Delivery/controllers"
	"lawgen/admin-service/Delivery/routers"
	"lawgen/admin-service/Repositories"
	usecases "lawgen/admin-service/Usecases"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func mongoConntect(ctx context.Context) (*mongo.Database, error) {
	mongoURI := "mongodb://localhost:27017"
	dbName := "lawgen_admin_db"

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		return nil, err
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	log.Println("Successfully connected to MongoDB.")
	return client.Database(dbName), nil
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := mongoConntect(ctx)
	if err != nil {
		log.Fatalf("Fatal Error: Could not connect to MongoDB: %v", err)
	}

	legalEntityRepo := Repositories.NewMongoEntityRepository(db)
	legalEntityUsecase := usecases.NewLegalEntityUsecase(legalEntityRepo)
	legalEntityController := controllers.NewLegalEntityController(legalEntityUsecase)

	//feedback
	feedbackRepo := Repositories.NewMongoFeedbackRepository(db)
    feedbackUsecase := usecases.NewFeedbackUsecase(feedbackRepo)
    feedbackController := controllers.NewFeedbackController(feedbackUsecase)

	router := routers.NewRouter(legalEntityController, feedbackController)
	log.Println("Starting Gin server on http://localhost:8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Fatal Error: Could not start Gin server: %s\n", err)
	}
}