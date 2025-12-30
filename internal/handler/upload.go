package handler

import (
	"net/http"
	"path/filepath"
	"scrapify/internal/service"
	"scrapify/pkg/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UploadHandler struct {
	storage       *service.StorageService
	workerService *service.WorkerService
}

func NewUploadHandler(ss *service.StorageService, ws *service.WorkerService) *UploadHandler {
	return &UploadHandler{
		storage:       ss,
		workerService: ws,
	}
}

func (h *UploadHandler) HandleUpload(c *gin.Context) {
	file, err := c.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "video file is required"})
		return
	}

	taskID := uuid.New().String()
	ext := filepath.Ext(file.Filename)
	storedName := taskID + ext
	uploadPath := filepath.Join("uploads", storedName)

	if err := c.SaveUploadedFile(file, uploadPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save video"})
		return
	}

	task := &models.Task{
		ID:           taskID,
		OriginalName: file.Filename,
		StoredName:   storedName,
		Status:       models.StatusPending,
		Stages:       []*models.Stage{},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := h.storage.SaveTask(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to initialize task"})
		return
	}

	// Dispatch to background worker
	h.workerService.Enqueue(task)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Upload successful, processing started.",
		"task_id": taskID,
	})
}
