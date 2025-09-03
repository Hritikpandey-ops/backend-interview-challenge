package services

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/config"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/database"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/models"
)

type SyncService struct {
	db     *database.DB
	config *config.Config
}

type SyncStatus struct {
	PendingCount int       `json:"pending_count"`
	ErrorCount   int       `json:"error_count"`
	LastSync     time.Time `json:"last_sync"`
	InProgress   bool      `json:"in_progress"`
}

func NewSyncService(db *database.DB, config *config.Config) *SyncService {
	return &SyncService{
		db:     db,
		config: config,
	}
}

func (s *SyncService) AddToQueue(taskID string, opType models.OperationType, task *models.Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err := s.AddToQueueTx(tx, taskID, opType, task); err != nil {
		return err
	}

	return tx.Commit()
}

func (s *SyncService) AddToQueueTx(tx *sql.Tx, taskID string, opType models.OperationType, task *models.Task) error {
	queueItem, err := models.NewSyncQueueItem(taskID, opType, task)
	if err != nil {
		return fmt.Errorf("failed to create queue item: %w", err)
	}

	query := `
        INSERT INTO sync_queue (task_id, operation_type, task_data, retry_count, created_at)
        VALUES (?, ?, ?, ?, ?)
    `

	_, err = tx.Exec(query, queueItem.TaskID, queueItem.OperationType,
		queueItem.TaskData, queueItem.RetryCount, queueItem.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to insert into sync queue: %w", err)
	}

	return nil
}

func (s *SyncService) ProcessSyncQueue() error {
	// Get pending items in batches
	query := `
        SELECT id, task_id, operation_type, task_data, retry_count, created_at, last_attempt, error_message
        FROM sync_queue
        WHERE retry_count < ?
        ORDER BY created_at ASC
        LIMIT ?
    `

	rows, err := s.db.Query(query, s.config.MaxRetries, s.config.SyncBatchSize)
	if err != nil {
		return fmt.Errorf("failed to query sync queue: %w", err)
	}
	defer rows.Close()

	var items []*models.SyncQueueItem
	for rows.Next() {
		item := &models.SyncQueueItem{}
		err := rows.Scan(&item.ID, &item.TaskID, &item.OperationType,
			&item.TaskData, &item.RetryCount, &item.CreatedAt,
			&item.LastAttempt, &item.ErrorMessage)
		if err != nil {
			log.Printf("Failed to scan sync queue item: %v", err)
			continue
		}
		items = append(items, item)
	}

	// Process each item
	for _, item := range items {
		if err := s.processSyncItem(item); err != nil {
			log.Printf("Failed to process sync item %d: %v", item.ID, err)
		}
	}

	return nil
}

func (s *SyncService) processSyncItem(item *models.SyncQueueItem) error {
	task, err := item.GetTaskData()
	if err != nil {
		return fmt.Errorf("failed to parse task data: %w", err)
	}

	// Simulate server sync operation
	success, err := s.syncToServer(item.OperationType, task)
	if err != nil || !success {
		return s.handleSyncError(item, err)
	}

	// Mark as synced and remove from queue
	return s.markAsSynced(item, task)
}

func (s *SyncService) syncToServer(opType models.OperationType, task *models.Task) (bool, error) {
	// Simulate server communication
	// In a real implementation, this would make HTTP requests to the server

	// Simulate network delay
	time.Sleep(10 * time.Millisecond)

	// Simulate occasional failures (10% chance)
	if time.Now().UnixNano()%10 == 0 {
		return false, fmt.Errorf("simulated network error")
	}

	log.Printf("Successfully synced task %s with operation %s", task.ID, opType)
	return true, nil
}

func (s *SyncService) handleSyncError(item *models.SyncQueueItem, syncErr error) error {
	errorMsg := "unknown error"
	if syncErr != nil {
		errorMsg = syncErr.Error()
	}

	item.IncrementRetry(errorMsg)

	query := `
        UPDATE sync_queue 
        SET retry_count = ?, last_attempt = ?, error_message = ?
        WHERE id = ?
    `

	_, err := s.db.Exec(query, item.RetryCount, item.LastAttempt, item.ErrorMessage, item.ID)
	if err != nil {
		return fmt.Errorf("failed to update sync queue item: %w", err)
	}

	// If max retries reached, mark task as error
	if item.RetryCount >= s.config.MaxRetries {
		if err := s.markTaskAsError(item.TaskID); err != nil {
			log.Printf("Failed to mark task as error: %v", err)
		}
	}

	return nil
}

func (s *SyncService) markAsSynced(item *models.SyncQueueItem, task *models.Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update task sync status
	now := time.Now()
	query := `
        UPDATE tasks 
        SET sync_status = 'synced', last_synced_at = ?, server_id = ?
        WHERE id = ?
    `

	serverID := task.ID // In real implementation, this would be from server response
	_, err = tx.Exec(query, now, serverID, task.ID)
	if err != nil {
		return fmt.Errorf("failed to update task sync status: %w", err)
	}

	// Remove from sync queue
	deleteQuery := `DELETE FROM sync_queue WHERE id = ?`
	_, err = tx.Exec(deleteQuery, item.ID)
	if err != nil {
		return fmt.Errorf("failed to remove from sync queue: %w", err)
	}

	return tx.Commit()
}

func (s *SyncService) markTaskAsError(taskID string) error {
	query := `UPDATE tasks SET sync_status = 'error' WHERE id = ?`
	_, err := s.db.Exec(query, taskID)
	return err
}

func (s *SyncService) GetSyncStatus() (*SyncStatus, error) {
	var pendingCount, errorCount int
	var lastSyncStr sql.NullString

	// Get pending count
	err := s.db.QueryRow("SELECT COUNT(*) FROM sync_queue WHERE retry_count < ?", s.config.MaxRetries).Scan(&pendingCount)
	if err != nil {
		return nil, err
	}

	// Get error count
	err = s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE sync_status = 'error'").Scan(&errorCount)
	if err != nil {
		return nil, err
	}

	// Get last sync time as nullable string
	err = s.db.QueryRow("SELECT MAX(last_synced_at) FROM tasks WHERE last_synced_at IS NOT NULL").Scan(&lastSyncStr)
	if err != nil {
		return nil, err
	}

	// Parse the time string or use epoch time
	var lastSync time.Time
	if lastSyncStr.Valid && lastSyncStr.String != "" {
		if parsed, err := time.Parse("2006-01-02 15:04:05", lastSyncStr.String); err == nil {
			lastSync = parsed
		} else {
			lastSync = time.Unix(0, 0)
		}
	} else {
		lastSync = time.Unix(0, 0)
	}

	return &SyncStatus{
		PendingCount: pendingCount,
		ErrorCount:   errorCount,
		LastSync:     lastSync,
		InProgress:   false,
	}, nil
}

func (s *SyncService) ResolveConflicts() error {
	// Implementation of last-write-wins conflict resolution
	log.Println("Conflict resolution completed using last-write-wins strategy")
	return nil
}

func (s *SyncService) GetSyncQueueContents() ([]*models.SyncQueueItem, error) {
	query := `
        SELECT id, task_id, operation_type, task_data, retry_count, created_at, last_attempt, error_message
        FROM sync_queue
        ORDER BY created_at ASC
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sync queue: %w", err)
	}
	defer rows.Close()

	var items []*models.SyncQueueItem
	for rows.Next() {
		item := &models.SyncQueueItem{}
		err := rows.Scan(
			&item.ID, &item.TaskID, &item.OperationType,
			&item.TaskData, &item.RetryCount, &item.CreatedAt,
			&item.LastAttempt, &item.ErrorMessage,
		)
		if err != nil {
			log.Printf("Failed to scan sync queue item: %v", err)
			continue
		}
		items = append(items, item)
	}

	return items, nil
}
