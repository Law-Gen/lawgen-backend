package main

import (
	"context"
	"log"

	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/delivery/httpserver"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/infrastructure/config"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/infrastructure/db"
	mongorepos "github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/infrastructure/repository/mongo"
	"github.com/Law-Gen/lawgen-backend/services/content-analytics-service/internal/usecase"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config load: %v", err)
	}

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// MongoDB
	client, database, err := db.ConnectMongo(context.Background(), cfg.Mongo.URI, cfg.Mongo.DBName)
	if err != nil {
		logger.Fatal("mongo connect failed", zap.Error(err))
	}
	defer func() { _ = client.Disconnect(context.Background()) }()

	// Repositories
	contentRepo := mongorepos.NewContentRepository(database, logger)
	feedbackRepo := mongorepos.NewFeedbackRepository(database, logger)
	legalRepo := mongorepos.NewLegalEntityRepository(database, logger)   // teammate to implement details
	analyticsRepo := mongorepos.NewAnalyticsRepository(database, logger) // teammate to implement details

	// Use cases
	contentUC := usecase.NewContentUsecase(contentRepo, logger)
	feedbackUC := usecase.NewFeedbackUsecase(feedbackRepo, logger)
	legalUC := usecase.NewLegalEntityUsecase(legalRepo, logger)     // teammate area
	analyticsUC := usecase.NewAnalyticsUsecase(analyticsRepo, logger) // teammate area

	// HTTP server
	srv := httpserver.NewServer(cfg, logger, contentUC, feedbackUC, legalUC, analyticsUC)
	if err := srv.Start(); err != nil {
		logger.Fatal("http server stopped", zap.Error(err))
	}
}