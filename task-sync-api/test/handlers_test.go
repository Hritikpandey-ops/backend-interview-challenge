package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/config"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/database"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/handlers"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/models"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/services"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestApp() (*gin.Engine, func()) {
	// Create temporary database
	cfg := &config.Config{
		DatabasePath:  ":memory:",
		SyncBatchSize: 10,
		MaxRetries:    3,
	}

	db, err := database.NewSQLiteDB(cfg.DatabasePath)
	if err != nil {
		panic(err)
	}

	syncService := services.NewSyncService(db, cfg)
	taskService := services.NewTaskService(db, syncService)
	taskHandler := handlers.NewTaskHandler(taskService)
	syncHandler := handlers.NewSyncHandler(syncService)

	gin.SetMode(gin.TestMode)
	router := gin.New()

	api := router.Group("/api")
	{
		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.POST("/tasks", taskHandler.CreateTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.POST("/sync/trigger", syncHandler.TriggerSync)
		api.GET("/sync/status", syncHandler.GetSyncStatus)
	}

	cleanup := func() {
		db.Close()
	}

	return router, cleanup
}

func TestCreateTask(t *testing.T) {
	router, cleanup := setupTestApp()
	defer cleanup()

	task := models.CreateTaskRequest{
		Title:       "Test Task",
		Description: stringPtr("Test Description"),
	}

	body, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "task")
	taskData := response["task"].(map[string]interface{})
	assert.Equal(t, "Test Task", taskData["title"])
}

func TestGetTasks(t *testing.T) {
	router, cleanup := setupTestApp()
	defer cleanup()

	// Create a task first
	task := models.CreateTaskRequest{
		Title: "Test Task",
	}

	body, _ := json.Marshal(task)
	req, _ := http.NewRequest("POST", "/api/tasks", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(httptest.NewRecorder(), req)

	// Get all tasks
	req, _ = http.NewRequest("GET", "/api/tasks", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "tasks")
	tasks := response["tasks"].([]interface{})
	assert.Len(t, tasks, 1)
}

func TestSyncStatus(t *testing.T) {
	router, cleanup := setupTestApp()
	defer cleanup()

	req, _ := http.NewRequest("GET", "/api/sync/status", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.Contains(t, response, "sync_status")
}

func TestMain(m *testing.M) {
	code := m.Run()
	os.Exit(code)
}
