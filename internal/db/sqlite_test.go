package db

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNew(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	defer db.Close()

	// Verify file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("database file was not created")
	}

	// Verify tables exist
	tables := []string{"sources", "opportunities", "settings", "reports", "schema_migrations"}
	for _, table := range tables {
		var name string
		err := db.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("table %s was not created: %v", table, err)
		}
	}
}

func TestNew_CreatesDataDirectory(t *testing.T) {
	tmpDir := t.TempDir()
	nestedPath := filepath.Join(tmpDir, "nested", "dir", "test.db")

	db, err := New(nestedPath)
	if err != nil {
		t.Fatalf("failed to create database in nested directory: %v", err)
	}
	defer db.Close()

	// Verify directory was created
	dir := filepath.Dir(nestedPath)
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("nested directory was not created")
	}
}

func TestMigrations_Idempotent(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// First connection
	db1, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}
	db1.Close()

	// Second connection - should not fail on migrations
	db2, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to reconnect to database: %v", err)
	}
	defer db2.Close()

	// Verify migrations table has entries
	var count int
	err = db2.QueryRow("SELECT COUNT(*) FROM schema_migrations").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query migrations: %v", err)
	}

	if count == 0 {
		t.Error("no migrations were recorded")
	}
}

func TestClose(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to create database: %v", err)
	}

	if err := db.Close(); err != nil {
		t.Errorf("failed to close database: %v", err)
	}

	// Verify connection is closed
	if err := db.Ping(); err == nil {
		t.Error("expected ping to fail after close")
	}
}
