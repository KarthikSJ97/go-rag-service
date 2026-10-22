package routes

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"rag-service/src/service"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type QueryRequest struct {
	Query string `json:"query"`
	TopK  uint64 `json:"top_k"`
	DocumentIDs []string `json:"document_ids"`
}

type QueryResponse struct {
	Query          string   `json:"query"`
	Answer         string   `json:"answer"`
	MatchedChunksCount  int `json:"matched_chunks_count"`
	ProcessingTime string   `json:"processing_time"`
}

// RegisterRoutes sets up all API paths on the Fiber app
func RegisterRoutes(app *fiber.App, container *service.Container) {
	// Healthcheck endpoint
	app.Get("/health", func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "healthy"})
	})

	api := app.Group("/api/v1")

	// 1. Ingest Endpoint
	api.Post("/ingest", func(c *fiber.Ctx) error {
		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing 'file' payload"})
		}

		rawText, err := container.Chunker.ParseUploadedFile(file)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
		}

		chunks := container.Chunker.ChunkText(rawText)
		if len(chunks) == 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "File is empty or contains no extractable text"})
		}

		// Concurrent Embeddings via Worker Pool
		embeddings := make([][]float32, len(chunks))
		var wg sync.WaitGroup
		sem := make(chan struct{}, 5)
		var errOccurred error

		for i, chunk := range chunks {
			wg.Add(1)
			go func(idx int, text string) {
				defer wg.Done()
				sem <- struct{}{}
				defer func() { <-sem }()

				emb, err := container.Ollama.GetEmbedding(text)
				if err != nil {
					errOccurred = err
					return
				}
				embeddings[idx] = emb
			}(i, chunk)
		}
		wg.Wait()

		if errOccurred != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Embedding failure: %v", errOccurred)})
		}

		docID := uuid.New().String()
		fileName := file.Filename
		fileSize := file.Size
		if err := container.Store.UpsertChunks(c.Context(), chunks, embeddings, docID, fileName, fileSize); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to store vectors: %v", err)})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":      "File ingested successfully",
			"document_id":  docID,
			"chunks_count": len(chunks),
		})
	})

	// 2. Query RAG Endpoint
	api.Post("/query", func(c *fiber.Ctx) error {
		start := time.Now()
		var req QueryRequest
		if err := c.BodyParser(&req); err != nil || strings.TrimSpace(req.Query) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Invalid request payload or empty query"})
		}

		if req.TopK == 0 {
			req.TopK = 3
		}

		queryVector, err := container.Ollama.GetEmbedding(req.Query)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Query embedding failed: %v", err)})
		}

		fmt.Printf("Query vector length: %d\n", len(queryVector))
		retrievedContexts, err := container.Store.Search(c.Context(), queryVector, req.TopK, req.DocumentIDs)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Vector search failed: %v", err)})
		}

		contextBlock := strings.Join(retrievedContexts, "\n---\n")
		prompt := fmt.Sprintf(
			"Context information is provided below:\n"+
				"---------------------\n%s\n"+
				"---------------------\n"+
				"Given the context information above and not prior knowledge, answer the query cleanly and accurately.\n"+
				"Query: %s",
			contextBlock, req.Query,
		)

		answer, err := container.Ollama.GenerateAnswer(prompt)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("LLM generation failed: %v", err)})
		}

		return c.Status(fiber.StatusOK).JSON(QueryResponse{
			Query:          req.Query,
			Answer:         answer,
			MatchedChunksCount:  len(retrievedContexts),
			ProcessingTime: time.Since(start).String(),
		})
	})

	// 3. List Ingested Documents Endpoint
	api.Get("/documents", func(c *fiber.Ctx) error {
		docs, err := container.Store.ListDocuments(c.Context())
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to list documents: %v", err)})
		}
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"documents": docs,
		})
	})

	// 4. Delete Vectors by Document ID Endpoint
	api.Delete("/documents/:id", func(c *fiber.Ctx) error {
		docID := c.Params("id")
		if strings.TrimSpace(docID) == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing document ID parameter"})
		}

		if err := container.Store.DeleteByDocumentID(c.Context(), docID); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": fmt.Sprintf("Failed to delete document vectors: %v", err)})
		}

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message":     "Document vectors deleted successfully",
			"document_id": docID,
		})
	})
}
