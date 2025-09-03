package main

import (
	"log"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/config"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/database"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/handlers"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/services"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.NewSQLiteDB(cfg.DatabasePath)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}
	defer db.Close()

	// Initialize services
	syncService := services.NewSyncService(db, cfg)
	taskService := services.NewTaskService(db, syncService)

	// Initialize handlers
	taskHandler := handlers.NewTaskHandler(taskService)
	syncHandler := handlers.NewSyncHandler(syncService)

	// Setup router
	router := gin.Default()

	// Add logging middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Task routes
	api := router.Group("/api")
	{
		api.GET("/tasks", taskHandler.GetTasks)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.POST("/tasks", taskHandler.CreateTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)

		// Sync routes
		api.GET("/sync/queue", syncHandler.GetSyncQueue)
		api.POST("/sync/trigger", syncHandler.TriggerSync)
		api.GET("/sync/status", syncHandler.GetSyncStatus)
		api.POST("/sync/batch", syncHandler.BatchSync)
	}

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	log.Printf("Server starting on port %s", cfg.Port)
	log.Fatal(router.Run(":" + cfg.Port))
}
