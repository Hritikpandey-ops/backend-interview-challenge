package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

type DB struct {
	*sql.DB
}

func NewSQLiteDB(dbPath string) (*DB, error) {
	var dsn string

	if dbPath == ":memory:" {
		// Use shared cache for in-memory databases to allow multiple connections
		dsn = "file:memdb1?mode=memory&cache=shared"
	} else {
		// Create directory if it doesn't exist for file databases
		dir := filepath.Dir(dbPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create database directory: %w", err)
		}
		dsn = dbPath
	}

	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable foreign keys and other optimizations
	pragmas := []string{
		"PRAGMA foreign_keys = ON",
		"PRAGMA journal_mode = WAL",
		"PRAGMA synchronous = NORMAL",
		"PRAGMA temp_store = memory",
	}

	for _, pragma := range pragmas {
		if _, err := db.Exec(pragma); err != nil {
			return nil, fmt.Errorf("failed to execute pragma %s: %w", pragma, err)
		}
	}

	dbConn := &DB{db}
	if err := dbConn.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return dbConn, nil
}

func (db *DB) migrate() error {
	// Enable foreign keys first
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	migrations := []string{
		`CREATE TABLE IF NOT EXISTS tasks (
            id TEXT PRIMARY KEY,
            title TEXT NOT NULL,
            description TEXT,
            completed BOOLEAN NOT NULL DEFAULT 0,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            is_deleted BOOLEAN NOT NULL DEFAULT 0,
            sync_status TEXT NOT NULL DEFAULT 'pending',
            server_id TEXT,
            last_synced_at DATETIME,
            CONSTRAINT chk_sync_status CHECK (sync_status IN ('pending', 'synced', 'error'))
        )`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_sync_status ON tasks(sync_status)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_is_deleted ON tasks(is_deleted)`,
		`CREATE INDEX IF NOT EXISTS idx_tasks_updated_at ON tasks(updated_at)`,
		`CREATE TABLE IF NOT EXISTS sync_queue (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            task_id TEXT NOT NULL,
            operation_type TEXT NOT NULL,
            task_data TEXT NOT NULL,
            retry_count INTEGER NOT NULL DEFAULT 0,
            created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
            last_attempt DATETIME,
            error_message TEXT,
            CONSTRAINT chk_operation_type CHECK (operation_type IN ('create', 'update', 'delete')),
            FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
        )`,
		`CREATE INDEX IF NOT EXISTS idx_sync_queue_retry_count ON sync_queue(retry_count)`,
		`CREATE INDEX IF NOT EXISTS idx_sync_queue_created_at ON sync_queue(created_at)`,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}

	return nil
}

func (db *DB) Close() error {
	return db.DB.Close()
}
