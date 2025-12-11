package sources

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	db, err := sql.Open("sqlite3", dbPath+"?_journal_mode=WAL&_foreign_keys=on")
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}

	// Create tables
	migrations := []string{
		`CREATE TABLE IF NOT EXISTS sources (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			type TEXT NOT NULL,
			name TEXT NOT NULL,
			url TEXT,
			config TEXT DEFAULT '{}',
			enabled BOOLEAN DEFAULT true,
			is_builtin BOOLEAN DEFAULT false,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE TABLE IF NOT EXISTS opportunities (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			source_id INTEGER REFERENCES sources(id),
			title TEXT NOT NULL,
			description TEXT,
			source TEXT NOT NULL,
			source_url TEXT NOT NULL,
			source_id_external TEXT NOT NULL,
			score INTEGER DEFAULT 0,
			signals TEXT DEFAULT '[]',
			ai_analysis TEXT,
			detected_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(source, source_id_external)
		)`,
	}

	for _, m := range migrations {
		if _, err := db.Exec(m); err != nil {
			t.Fatalf("failed to run migration: %v", err)
		}
	}

	t.Cleanup(func() {
		db.Close()
	})

	return db
}

func TestManager_RegisterFactory(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Check default factories are registered
	if _, ok := m.factories["hackernews"]; !ok {
		t.Error("hackernews factory not registered")
	}
	if _, ok := m.factories["github"]; !ok {
		t.Error("github factory not registered")
	}
	if _, ok := m.factories["npm"]; !ok {
		t.Error("npm factory not registered")
	}
	if _, ok := m.factories["devto"]; !ok {
		t.Error("devto factory not registered")
	}
}

func TestManager_SeedSources(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Seed sources
	if err := m.repo.Seed(); err != nil {
		t.Fatalf("failed to seed sources: %v", err)
	}

	// Verify sources were created
	sources, err := m.repo.GetAll()
	if err != nil {
		t.Fatalf("failed to get sources: %v", err)
	}

	if len(sources) != 4 {
		t.Errorf("expected 4 default sources, got %d", len(sources))
	}

	// Check source types
	types := make(map[string]bool)
	for _, s := range sources {
		types[s.Type] = true
		if !s.IsBuiltin {
			t.Errorf("expected source %s to be builtin", s.Name)
		}
		if !s.Enabled {
			t.Errorf("expected source %s to be enabled", s.Name)
		}
	}

	expectedTypes := []string{"hackernews", "github", "npm", "devto"}
	for _, et := range expectedTypes {
		if !types[et] {
			t.Errorf("expected source type %s to be present", et)
		}
	}
}

func TestManager_SeedIdempotent(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Seed twice
	if err := m.repo.Seed(); err != nil {
		t.Fatalf("first seed failed: %v", err)
	}
	if err := m.repo.Seed(); err != nil {
		t.Fatalf("second seed failed: %v", err)
	}

	// Should still have only 4 sources
	sources, err := m.repo.GetAll()
	if err != nil {
		t.Fatalf("failed to get sources: %v", err)
	}

	if len(sources) != 4 {
		t.Errorf("expected 4 sources after double seed, got %d", len(sources))
	}
}

func TestManager_StartStop(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Start
	if err := m.Start(); err != nil {
		t.Fatalf("failed to start manager: %v", err)
	}

	if !m.isRunning {
		t.Error("expected manager to be running")
	}

	// Stop
	m.Stop()

	if m.isRunning {
		t.Error("expected manager to be stopped")
	}
}

func TestManager_SaveOpportunity(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Seed to get a source
	if err := m.repo.Seed(); err != nil {
		t.Fatalf("failed to seed: %v", err)
	}

	sources, _ := m.repo.GetAll()
	if len(sources) == 0 {
		t.Fatal("no sources found")
	}

	opp := Opportunity{
		Title:            "Test Opportunity",
		Description:      "A test opportunity",
		SourceType:       "hackernews",
		SourceURL:        "https://example.com",
		SourceIDExternal: "test-123",
	}

	if err := m.saveOpportunity(sources[0].ID, opp); err != nil {
		t.Fatalf("failed to save opportunity: %v", err)
	}

	// Verify it was saved
	var count int
	err := db.QueryRow("SELECT COUNT(*) FROM opportunities WHERE source_id_external = ?", "test-123").Scan(&count)
	if err != nil {
		t.Fatalf("failed to query opportunities: %v", err)
	}

	if count != 1 {
		t.Errorf("expected 1 opportunity, got %d", count)
	}
}

func TestManager_FetchAll_EmptySources(t *testing.T) {
	db := setupTestDB(t)
	m := NewManager(db, 60)

	// Don't seed - no sources
	err := m.FetchAll(context.Background())
	if err != nil {
		t.Errorf("expected no error with empty sources, got %v", err)
	}
}
