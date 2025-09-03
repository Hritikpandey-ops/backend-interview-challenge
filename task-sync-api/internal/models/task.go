package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type SyncStatus string

const (
	SyncStatusPending SyncStatus = "pending"
	SyncStatusSynced  SyncStatus = "synced"
	SyncStatusError   SyncStatus = "error"
)

type Task struct {
	ID           string     `json:"id" db:"id"`
	Title        string     `json:"title" db:"title"`
	Description  *string    `json:"description" db:"description"`
	Completed    bool       `json:"completed" db:"completed"`
	IsDeleted    bool       `json:"is_deleted" db:"is_deleted"`
	SyncStatus   SyncStatus `json:"sync_status" db:"sync_status"`
	ServerID     *string    `json:"server_id" db:"server_id"`
	LastSyncedAt *time.Time `json:"last_synced_at" db:"last_synced_at"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at" db:"updated_at"`
}

func (t *Task) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		ID           string     `json:"id"`
		Title        string     `json:"title"`
		Description  *string    `json:"description"`
		Completed    bool       `json:"completed"`
		IsDeleted    bool       `json:"is_deleted"`
		SyncStatus   SyncStatus `json:"sync_status"`
		ServerID     *string    `json:"server_id"`
		LastSyncedAt *string    `json:"last_synced_at"`
		CreatedAt    string     `json:"created_at"`
		UpdatedAt    string     `json:"updated_at"`
	}{
		ID:           t.ID,
		Title:        t.Title,
		Description:  t.Description,
		Completed:    t.Completed,
		IsDeleted:    t.IsDeleted,
		SyncStatus:   t.SyncStatus,
		ServerID:     t.ServerID,
		LastSyncedAt: formatTimePtr(t.LastSyncedAt),
		CreatedAt:    t.CreatedAt.Format(time.RFC3339),
		UpdatedAt:    t.UpdatedAt.Format(time.RFC3339),
	})
}

func formatTimePtr(t *time.Time) *string {
	if t == nil {
		return nil
	}
	formatted := t.Format(time.RFC3339)
	return &formatted
}

type CreateTaskRequest struct {
	Title       string  `json:"title" binding:"required"`
	Description *string `json:"description"`
}

type UpdateTaskRequest struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Completed   *bool   `json:"completed"`
}

func NewTask(title string, description *string) *Task {
	now := time.Now()
	return &Task{
		ID:          uuid.New().String(),
		Title:       title,
		Description: description,
		Completed:   false,
		CreatedAt:   now,
		UpdatedAt:   now,
		IsDeleted:   false,
		SyncStatus:  SyncStatusPending,
	}
}

func (t *Task) Update(req *UpdateTaskRequest) {
	if req.Title != nil {
		t.Title = *req.Title
	}
	if req.Description != nil {
		t.Description = req.Description
	}
	if req.Completed != nil {
		t.Completed = *req.Completed
	}
	t.UpdatedAt = time.Now()
	t.SyncStatus = SyncStatusPending
}

func (t *Task) SoftDelete() {
	t.IsDeleted = true
	t.UpdatedAt = time.Now()
	t.SyncStatus = SyncStatusPending
}
