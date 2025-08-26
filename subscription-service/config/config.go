package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DbName             string
	MongoURI          string
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

	AppConfig = &Config{
		DbName:   dbName,
		MongoURI: mongoURI,
	}
}

