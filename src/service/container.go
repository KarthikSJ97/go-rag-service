package service

import (
	"context"
	"log"
	"time"

	"rag-service/src/config"
	"rag-service/src/document"
	"rag-service/src/llm"
	"rag-service/src/vectorstore"
)

type Container struct {
	Cfg     *config.Config
	Chunker *document.DocumentChunker
	Ollama  *llm.OllamaClient
	Store   *vectorstore.VectorStore
}

// NewContainer initializes all dependencies and waits for external services to be ready
func NewContainer(cfg *config.Config) (*Container, error) {
	chunker := document.NewDocumentChunker(cfg.ChunkSize, cfg.Overlap)
	ollama := llm.NewOllamaClient(cfg.OllamaHost, cfg.OllamaEmbModel, cfg.OllamaLlmModel)

	var store *vectorstore.VectorStore
	var err error

	// Retry loop for Qdrant cold-starts
	for i := 0; i < 10; i++ {
		store, err = vectorstore.NewVectorStore(cfg.QdrantHost, cfg.QdrantPort, cfg.QdrantCollectionName)
		if err == nil {
			break
		}
		log.Printf("Waiting for Qdrant connection... attempt %d/10", i+1)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return nil, err
	}

	// Ensure Qdrant collection exists (bge-small-en-v1.5 vector size = 384)
	ctx := context.Background()
	if err := store.EnsureCollection(ctx, uint64(cfg.VectorSize)); err != nil {
		return nil, err
	}

	return &Container{
		Cfg:     cfg,
		Chunker: chunker,
		Ollama:  ollama,
		Store:   store,
	}, nil
}
