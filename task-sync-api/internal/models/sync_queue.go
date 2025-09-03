package models

import (
	"encoding/json"
	"time"
)

type OperationType string

const (
	OperationTypeCreate OperationType = "create"
	OperationTypeUpdate OperationType = "update"
	OperationTypeDelete OperationType = "delete"
)

type SyncQueueItem struct {
	ID            int           `json:"id" db:"id"`
	TaskID        string        `json:"task_id" db:"task_id"`
	OperationType OperationType `json:"operation_type" db:"operation_type"`
	TaskData      string        `json:"task_data" db:"task_data"`
	RetryCount    int           `json:"retry_count" db:"retry_count"`
	CreatedAt     time.Time     `json:"created_at" db:"created_at"`
	LastAttempt   *time.Time    `json:"last_attempt" db:"last_attempt"`
	ErrorMessage  *string       `json:"error_message" db:"error_message"`
}

func NewSyncQueueItem(taskID string, opType OperationType, task *Task) (*SyncQueueItem, error) {
	taskData, err := json.Marshal(task)
	if err != nil {
		return nil, err
	}

	return &SyncQueueItem{
		TaskID:        taskID,
		OperationType: opType,
		TaskData:      string(taskData),
		RetryCount:    0,
		CreatedAt:     time.Now(),
	}, nil
}

func (sq *SyncQueueItem) GetTaskData() (*Task, error) {
	var task Task
	err := json.Unmarshal([]byte(sq.TaskData), &task)
	return &task, err
}

func (sq *SyncQueueItem) IncrementRetry(errorMsg string) {
	sq.RetryCount++
	now := time.Now()
	sq.LastAttempt = &now
	sq.ErrorMessage = &errorMsg
}
