package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application.
type Config struct {
	Port                    string
	RedisAddr               string
	MongoURI                string
	MongoDBName             string
	RAGServiceAddr          string
	GoogleAPIKey            string
	GeminiModel             string
	LLMPromptRefine         string
	LLMPromptAnswer         string
	LLMPromptNoResult       string
	LLMPromptConverter      string
	SessionTTLSeconds       int
	ChatHistorySyncInterval time.Duration
	STTApiBase              string // e.g. http://127.0.0.1:8000/speech-to-text/
	TranslateApiUrl         string // e.g. http://127.0.0.1:8000/translate
	TTSApiUrl               string // e.g. http://127.0.0.1:8000/text-to-speech
}

// New loads configuration from environment variables.
func New() (*Config, error) {
	// Load .env file if it exists
	_ = godotenv.Load()

	return &Config{
		Port:                    getEnv("PORT", "8080"),
		RedisAddr:               getEnv("REDIS_ADDR", "localhost:6379"),
		MongoURI:                getEnv("DB_ENDPOINT", "mongodb://localhost:27017"),
		MongoDBName:             getEnv("DB_NAME", "chatdb"),
		GoogleAPIKey:            getEnv("GOOGLE_API_KEY", ""),
		GeminiModel:             getEnv("GEMINI_MODEL", "gemini-pro"),
		RAGServiceAddr:          getEnv("RAG_SERVICE_ADDR", "localhost:50051"), // gRPC address
		LLMPromptRefine:         getEnv("LLM_PROMPT_REFINE", "Refine the following query for a RAG system, making it concise and clear: {{.Query}}"),
		LLMPromptAnswer:         getEnv("LLM_PROMPT_ANSWER", "Based on the provided context and conversation history, answer the user's question. Adhere strictly to word limits and sources. Do not hallucinate. Context: {{.RAGResults}} History: {{.ChatHistory}} Question: {{.Query}} MaxWords: {{.MaxWords}} MaxRefs: {{.MaxRefs}}"),
		LLMPromptNoResult:       getEnv("LLM_PROMPT_NO_RESULT", "I couldn't find information related to your question. Here are some refined questions you might try, separated by newlines: {{.Query}}"),
		LLMPromptConverter:      getEnv("LLM_PROMPT_CONVERTER", "Translate the following text to English, maintaining its original meaning and context. Text: {{.Text}}"),
		SessionTTLSeconds:       getEnvAsInt("SESSION_TTL_SECONDS", 7200),                                            // 2 hours
		ChatHistorySyncInterval: time.Second * time.Duration(getEnvAsInt("CHAT_HISTORY_SYNC_INTERVAL_SECONDS", 300)), // 5 minutes
		STTApiBase:              getEnv("STT_API_BASE", "http://127.0.0.1:8000/speech-to-text/"),
		TranslateApiUrl:         getEnv("TRANSLATE_API_URL", "http://127.0.0.1:8000/translate"),
		TTSApiUrl:               getEnv("TTS_API_URL", "http://127.0.0.1:8000/text-to-speech"),
	}, nil

}

func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// getEnvAsInt returns the value of the environment variable as an int, or fallback if not set / invalid.
func getEnvAsInt(key string, fallback int) int {
	if valueStr, exists := os.LookupEnv(key); exists {
		if value, err := strconv.Atoi(valueStr); err == nil {
			return value
		}
	}
	return fallback
}
