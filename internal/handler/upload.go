package handler

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"scrapify/internal/service"
	"scrapify/pkg/models"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	mediaService  *service.MediaService
	imageService  *service.ImageService
	geminiService *service.GeminiService
}

func NewUploadHandler(ms *service.MediaService, is *service.ImageService, gs *service.GeminiService) *UploadHandler {
	return &UploadHandler{
		mediaService:  ms,
		imageService:  is,
		geminiService: gs,
	}
}

func (h *UploadHandler) HandleUpload(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video file is required"})
		return
	}

	uploadPath := filepath.Join("uploads", file.Filename)
	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save video"})
		return
	}

	// 1. Generate Snapshots
	snapshotDir := filepath.Join("snapshots", file.Filename)
	snapshots, err := h.mediaService.GenerateSnapshots(uploadPath, snapshotDir)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate snapshots: %v", err)})
		return
	}

	// 2. Extract Audio
	audioPath := filepath.Join("processed", file.Filename+".mp3")
	if err := h.mediaService.ExtractAudio(uploadPath, audioPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to extract audio: %v", err)})
		return
	}

	// 3. Create Image Grid
	gridPath := filepath.Join("processed", file.Filename+"_grid.jpg")
	if err := h.imageService.CreateGrid(snapshots, gridPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create image grid: %v", err)})
		return
	}

	// 4. Transcribe Audio
	transcription, err := h.geminiService.TranscribeAudio(audioPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to transcribe audio: %v", err)})
		return
	}

	// 5. Generate Study Guide
	guide, err := h.geminiService.GenerateStudyGuide(transcription, gridPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to generate study guide: %v", err)})
		return
	}

	// 6. Save final JSON
	result := models.AnalysisResult{
		Transcription: transcription,
		StudyGuide:    *guide,
		ImagePath:     gridPath,
	}

	outputPath := filepath.Join("output", file.Filename+".json")
	outputFile, err := os.Create(outputPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save output JSON"})
		return
	}
	defer outputFile.Close()

	// Return result
	c.JSON(http.StatusOK, result)
}
