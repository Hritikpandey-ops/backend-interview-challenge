package tests

import (
	"testing"

	"github.com/pearlthoughts/backend-interview-challenge-1/task-sync-api/internal/database"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDatabaseSetup(t *testing.T) {
	db, err := database.NewSQLiteDB(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Check tables exist
	tables := []string{"tasks", "sync_queue"}
	for _, table := range tables {
		var count int
		err = db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count, "Table %s should exist", table)

		// Try to query the table
		err = db.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count)
		require.NoError(t, err, "Should be able to query table %s", table)
	}

	t.Log("Database setup successful - all tables created and accessible")
}
