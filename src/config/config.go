package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 string
	QdrantHost           string
	QdrantPort           string
	QdrantCollectionName string
	OllamaHost           string
	OllamaEmbModel       string
	OllamaLlmModel       string
}

func LoadConfig() *Config {
	// Attempt to load .env file (ignores error if running inside Docker where env vars are injected directly)
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or failed to load, relying on system environment variables")
	}

	return &Config{
		Port:                 getEnv("PORT", "8080"),
		QdrantHost:           getEnv("QDRANT_HOST", "localhost"),
		QdrantPort:           getEnv("QDRANT_PORT", "6334"),
		QdrantCollectionName: getEnv("QDRANT_COLLECTION_NAME", "rag_documents"),
		OllamaHost:           getEnv("OLLAMA_HOST", "http://localhost:11434"),
		OllamaEmbModel:       getEnv("OLLAMA_EMB_MODEL", "bge-small-en-v1.5"),
		OllamaLlmModel:       getEnv("OLLAMA_LLM_MODEL", "llama3"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
