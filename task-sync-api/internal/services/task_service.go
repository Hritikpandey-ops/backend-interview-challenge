package services

import (
	"database/sql"
	"fmt"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/database"
	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/models"
)

type TaskService struct {
	db          *database.DB
	syncService *SyncService
}

func NewTaskService(db *database.DB, syncService *SyncService) *TaskService {
	return &TaskService{
		db:          db,
		syncService: syncService,
	}
}

func (s *TaskService) GetAllTasks() ([]*models.Task, error) {
	query := `
        SELECT id, title, description, completed, created_at, updated_at, 
               is_deleted, sync_status, server_id, last_synced_at
        FROM tasks 
        WHERE is_deleted = 0
        ORDER BY updated_at DESC, created_at DESC
    `

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query tasks: %w", err)
	}
	defer rows.Close()

	var tasks []*models.Task
	for rows.Next() {
		task := &models.Task{}
		var description, serverID sql.NullString
		var lastSyncedAt sql.NullTime

		err := rows.Scan(
			&task.ID, &task.Title, &description, &task.Completed,
			&task.CreatedAt, &task.UpdatedAt, &task.IsDeleted,
			&task.SyncStatus, &serverID, &lastSyncedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan task: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			task.Description = &description.String
		}
		if serverID.Valid {
			task.ServerID = &serverID.String
		}
		if lastSyncedAt.Valid {
			task.LastSyncedAt = &lastSyncedAt.Time
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (s *TaskService) GetTaskByID(id string) (*models.Task, error) {
	query := `
        SELECT id, title, description, completed, created_at, updated_at, 
               is_deleted, sync_status, server_id, last_synced_at
        FROM tasks 
        WHERE id = ? AND is_deleted = 0
    `

	task := &models.Task{}
	var description, serverID sql.NullString
	var lastSyncedAt sql.NullTime

	err := s.db.QueryRow(query, id).Scan(
		&task.ID, &task.Title, &description, &task.Completed,
		&task.CreatedAt, &task.UpdatedAt, &task.IsDeleted,
		&task.SyncStatus, &serverID, &lastSyncedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("task not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get task: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		task.Description = &description.String
	}
	if serverID.Valid {
		task.ServerID = &serverID.String
	}
	if lastSyncedAt.Valid {
		task.LastSyncedAt = &lastSyncedAt.Time
	}

	return task, nil
}

func (s *TaskService) CreateTask(req *models.CreateTaskRequest) (*models.Task, error) {
	task := models.NewTask(req.Title, req.Description)

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert task
	query := `
        INSERT INTO tasks (id, title, description, completed, created_at, updated_at, 
                          is_deleted, sync_status, server_id, last_synced_at)
        VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
    `

	_, err = tx.Exec(query, task.ID, task.Title, task.Description, task.Completed,
		task.CreatedAt, task.UpdatedAt, task.IsDeleted, task.SyncStatus,
		task.ServerID, task.LastSyncedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to insert task: %w", err)
	}

	// Add to sync queue
	if err := s.syncService.AddToQueueTx(tx, task.ID, models.OperationTypeCreate, task); err != nil {
		return nil, fmt.Errorf("failed to add to sync queue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return task, nil
}

func (s *TaskService) UpdateTask(id string, req *models.UpdateTaskRequest) (*models.Task, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing task
	task, err := s.GetTaskByID(id)
	if err != nil {
		return nil, err
	}

	// Update task
	task.Update(req)

	query := `
        UPDATE tasks 
        SET title = ?, description = ?, completed = ?, updated_at = ?, sync_status = ?
        WHERE id = ? AND is_deleted = 0
    `

	result, err := tx.Exec(query, task.Title, task.Description, task.Completed,
		task.UpdatedAt, task.SyncStatus, id)
	if err != nil {
		return nil, fmt.Errorf("failed to update task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return nil, fmt.Errorf("task not found")
	}

	// Add to sync queue
	if err := s.syncService.AddToQueueTx(tx, task.ID, models.OperationTypeUpdate, task); err != nil {
		return nil, fmt.Errorf("failed to add to sync queue: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return task, nil
}

func (s *TaskService) DeleteTask(id string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Get existing task
	task, err := s.GetTaskByID(id)
	if err != nil {
		return err
	}

	// Soft delete
	task.SoftDelete()

	query := `
        UPDATE tasks 
        SET is_deleted = 1, updated_at = ?, sync_status = ?
        WHERE id = ?
    `

	result, err := tx.Exec(query, task.UpdatedAt, task.SyncStatus, id)
	if err != nil {
		return fmt.Errorf("failed to delete task: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("task not found")
	}

	// Add to sync queue
	if err := s.syncService.AddToQueueTx(tx, task.ID, models.OperationTypeDelete, task); err != nil {
		return fmt.Errorf("failed to add to sync queue: %w", err)
	}

	return tx.Commit()
}
