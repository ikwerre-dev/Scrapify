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
	ms := service.NewMediaService()
	is := service.NewImageService()
	gs, err := service.NewGeminiService(apiKey)
	if err != nil {
		log.Fatalf("Failed to initialize Gemini service: %v", err)
	}
	ss := service.NewStorageService("tasks.json")
	ws := service.NewWorkerService(ms, is, gs, ss, 100)

	// Start worker pool (e.g., 5 concurrent workers)
	ws.Start(5)

	// Initialize handlers
	uh := handler.NewUploadHandler(ss, ws)
	sh := handler.NewStatusHandler(ss)

	// Setup router
	r := gin.Default()

	// Max upload size 100MB
	r.MaxMultipartMemory = 100 << 20

	r.POST("/upload", uh.HandleUpload)
	r.GET("/status/:id", sh.GetStatus)

	// Serve static files for processed images
	r.Static("/processed", "./processed")

	log.Println("Server starting on :8080...")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
