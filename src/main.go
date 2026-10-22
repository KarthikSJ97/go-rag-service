package main

import (
	"log"
	"strconv"

	"rag-service/src/config"
	"rag-service/src/routes"
	"rag-service/src/service"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	// 1. Load configuration
	cfg := config.LoadConfig()

	// 2. Initialize Dependency Container (Clients & Databases)
	container, err := service.NewContainer(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize service container: %v", err)
	}

	// 3. Initialize Fiber HTTP Server
	app := fiber.New(fiber.Config{
		BodyLimit: cfg.MaxPayloadSizeInMB * 1024 * 1024,
	})
	app.Use(logger.New())

	// 4. Register Routes
	routes.RegisterRoutes(app, container)

	// 5. Start Server
	log.Printf("RAG Web Service listening on port %s...", cfg.Port)
	log.Fatal(app.Listen(":" + strconv.Itoa(cfg.Port)))
}
