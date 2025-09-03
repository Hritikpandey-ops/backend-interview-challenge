package handlers

import (
	"net/http"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/services"

	"github.com/gin-gonic/gin"
)

type SyncHandler struct {
	syncService *services.SyncService
}

func NewSyncHandler(syncService *services.SyncService) *SyncHandler {
	return &SyncHandler{syncService: syncService}
}

func (h *SyncHandler) TriggerSync(c *gin.Context) {
	err := h.syncService.ProcessSyncQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "sync completed successfully"})
}

func (h *SyncHandler) GetSyncStatus(c *gin.Context) {
	status, err := h.syncService.GetSyncStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sync_status": status})
}

func (h *SyncHandler) BatchSync(c *gin.Context) {
	// Process sync queue
	err := h.syncService.ProcessSyncQueue()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Resolve any conflicts
	err = h.syncService.ResolveConflicts()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Get updated status
	status, err := h.syncService.GetSyncStatus()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "batch sync completed",
		"sync_status": status,
	})
}

func (h *SyncHandler) GetSyncQueue(c *gin.Context) {
	items, err := h.syncService.GetSyncQueueContents()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sync_queue": items})
}
