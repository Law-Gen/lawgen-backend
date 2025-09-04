package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	DbName             string
	MongoURI          string
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	GoogleClientID     string
	GoogleClientSecret string
	OauthStateString    string
}

// AppConfig is the global config instance
var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system env variables")
	}

	dbName := os.Getenv("MONGODB_DB")
	mongoURI := os.Getenv("MONGODB_URI")

	accessSecret := os.Getenv("ACCESS_TOKEN_SECRET")
	refreshSecret := os.Getenv("REFRESH_TOKEN_SECRET")

	accessExpiry := parseDuration(os.Getenv("ACCESS_TOKEN_EXPIRY"))
	refreshExpiry := parseDuration(os.Getenv("REFRESH_TOKEN_EXPIRY"))

	googleClientID := os.Getenv("GOOGLE_OAUTH_CLIENT_ID")
	googleClientSecret := os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET")
	oauthStateString := os.Getenv("OAUTH_STATE_STRING")

	AppConfig = &Config{
		DbName 			:   dbName,
		MongoURI		:mongoURI,
		AccessTokenSecret:  accessSecret,
		RefreshTokenSecret: refreshSecret,
		AccessTokenExpiry:  accessExpiry,
		RefreshTokenExpiry: refreshExpiry,
		GoogleClientID:     googleClientID,
		GoogleClientSecret: googleClientSecret,
		OauthStateString:    oauthStateString,
	}
}

func parseDuration(value string) time.Duration {
	duration, err := time.ParseDuration(value)
	if err != nil {
		log.Fatalf("Invalid duration(ACCESS_TOKEN_EXPIRY/REFRESH_TOKEN_EXPIRY) value: %s", value)
		return 0
	}
	return duration
}
