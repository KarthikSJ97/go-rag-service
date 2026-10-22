package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                 int
	MaxPayloadSizeInMB   int
	QdrantHost           string
	QdrantPort           int
	QdrantCollectionName string
	OllamaHost           string
	OllamaEmbModel       string
	OllamaLlmModel       string
	ChunkSize            int
	Overlap              int
	VectorSize           int
}

func LoadConfig() *Config {
	// Attempt to load .env file (ignores error if running inside Docker where env vars are injected directly)
	if err := godotenv.Load(); err != nil {
		log.Println("[Config] No .env file found or failed to load, relying on system environment variables")
	}

	return &Config{
		Port:                 getEnvInt("PORT", 8080),
		MaxPayloadSizeInMB:   getEnvInt("MAX_PAYLOAD_SIZE_IN_MB", 20),
		QdrantHost:           getEnvString("QDRANT_HOST", "localhost"),
		QdrantPort:           getEnvInt("QDRANT_PORT", 6334),
		QdrantCollectionName: getEnvString("QDRANT_COLLECTION_NAME", "rag_documents"),
		OllamaHost:           getEnvString("OLLAMA_HOST", "http://localhost:11434"),
		OllamaEmbModel:       getEnvString("OLLAMA_EMB_MODEL", "bge-small-en-v1.5"),
		OllamaLlmModel:       getEnvString("OLLAMA_LLM_MODEL", "llama3"),
		ChunkSize:            getEnvInt("CHUNK_SIZE", 100),
		Overlap:              getEnvInt("OVERLAP", 15),
		VectorSize:           getEnvInt("VECTOR_SIZE", 384),
	}
}

func getEnvString(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}
