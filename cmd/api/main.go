package main

import (
	"log"
	"os"
	"scrapify/internal/handler"
	"scrapify/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		log.Fatal("GEMINI_API_KEY is required")
	}

	// Initialize services
	mediaService := service.NewMediaService()
	imageService := service.NewImageService()
	geminiService, err := service.NewGeminiService(apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini service: %v", err)
	}

	// Initialize handlers
	uploadHandler := handler.NewUploadHandler(mediaService, imageService, geminiService)

	// Setup router
	r := gin.Default()

	// Max upload size 100MB
	r.MaxMultipartMemory = 100 << 20

	r.POST("/upload", uploadHandler.HandleUpload)

	// Serve static files for processed images
	r.Static("/processed", "./processed")

	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
