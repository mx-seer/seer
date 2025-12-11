package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// DB wraps the sql.DB connection
type DB struct {
	*sql.DB
}

// migrations contains all database migrations in order
var migrations = []string{
	// Migration 1: Initial schema
	`CREATE TABLE IF NOT EXISTS sources (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		url TEXT,
		config TEXT DEFAULT '{}',
		enabled BOOLEAN DEFAULT true,
		is_builtin BOOLEAN DEFAULT false,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,

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
	);`,

	`CREATE TABLE IF NOT EXISTS settings (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL
	);`,

	`CREATE TABLE IF NOT EXISTS reports (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		period_start DATETIME NOT NULL,
		period_end DATETIME NOT NULL,
		opportunity_count INTEGER DEFAULT 0,
		content_human TEXT,
		content_prompt TEXT,
		ai_analysis TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,

	`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,

	// Migration 2: Alerts table (Pro feature)
	`CREATE TABLE IF NOT EXISTS alerts (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		type TEXT NOT NULL,
		name TEXT NOT NULL,
		destination TEXT NOT NULL,
		min_score INTEGER DEFAULT 50,
		enabled BOOLEAN DEFAULT true,
		config TEXT DEFAULT '{}',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,

	// Migration 3: License table (Pro feature)
	`CREATE TABLE IF NOT EXISTS license (
		key TEXT PRIMARY KEY,
		email TEXT,
		tier TEXT,
		expires_at DATETIME,
		verified_at DATETIME,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`,
}

// New creates a new database connection and runs migrations
func New(dbPath string) (*DB, error) {
	// Ensure data directory exists
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Open database with WAL mode for better concurrency
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)")
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	wrapper := &DB{db}

	// Run migrations
	if err := wrapper.migrate(); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return wrapper, nil
}

// migrate runs all pending migrations
func (db *DB) migrate() error {
	// Create migrations table if not exists
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (
		version INTEGER PRIMARY KEY,
		applied_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}

	// Get current version
	var currentVersion int
	row := db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations")
	if err := row.Scan(&currentVersion); err != nil {
		return err
	}

	// Run pending migrations
	for i, migration := range migrations {
		version := i + 1
		if version <= currentVersion {
			continue
		}

		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", version, err)
		}

		if _, err := db.Exec("INSERT INTO schema_migrations (version) VALUES (?)", version); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", version, err)
		}
	}

	return nil
}

// Close closes the database connection
func (db *DB) Close() error {
	return db.DB.Close()
}
