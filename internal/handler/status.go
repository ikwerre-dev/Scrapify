package handler

import (
	"net/http"
	"scrapify/internal/service"

	"github.com/gin-gonic/gin"
)

type StatusHandler struct {
	storage *service.StorageService
}

func NewStatusHandler(ss *service.StorageService) *StatusHandler {
	return &StatusHandler{storage: ss}
}

func (h *StatusHandler) GetStatus(c *gin.Context) {
	id := c.Param("id")
	task, err := h.storage.GetTask(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to retrieve task"})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	c.JSON(http.StatusOK, task)
}
