package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port           string
	MongoURI       string
	MongoDBName    string
	// RAGServiceAddr string
	// GoogleAPIKey   string
	// GeminiModel    string
}

// New loads configuration from environment variables.
func New() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	port := getEnv("PORT", "8080")
	mongoURI := getEnv("DB_ENDPOINT", "")
	mongoDBName := getEnv("DB_NAME", "chat_db")
	// ragServiceAddr := getEnv("RAG_SERVICE_ADDR", "localhost:50051")
	// googleAPIKey := getEnv("GOOGLE_API_KEY", "")
	// geminiModel := getEnv("GEMINI_MODEL", "gemini-pro")

	if mongoURI == "" {
		return nil, fmt.Errorf("DB_ENDPOINT environment variable is required")
	}
	// if googleAPIKey == "" {
	// 	return nil, fmt.Errorf("GOOGLE_API_KEY environment variable is required")
	// }

	return &Config{
		Port:           port,
		MongoURI:       mongoURI,
		MongoDBName:    mongoDBName,
		// RAGServiceAddr: ragServiceAddr,
		// GoogleAPIKey:   googleAPIKey,
		// GeminiModel:    geminiModel,
	}, nil
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
