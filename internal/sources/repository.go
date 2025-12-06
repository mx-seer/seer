package sources

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"
)

// Repository handles source persistence
type Repository struct {
	db *sql.DB
}

// SourceRecord represents a source in the database
type SourceRecord struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	URL       string    `json:"url"`
	Config    string    `json:"config"`
	Enabled   bool      `json:"enabled"`
	IsBuiltin bool      `json:"is_builtin"`
	CreatedAt time.Time `json:"created_at"`
}

// NewRepository creates a new source repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// GetAll returns all sources
func (r *Repository) GetAll() ([]SourceRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, type, name, url, config, enabled, is_builtin, created_at
		FROM sources
		ORDER BY is_builtin DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query sources: %w", err)
	}
	defer rows.Close()

	var sources []SourceRecord
	for rows.Next() {
		var s SourceRecord
		var url, config sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &s.Name, &url, &config, &s.Enabled, &s.IsBuiltin, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		s.URL = url.String
		s.Config = config.String
		if s.Config == "" {
			s.Config = "{}"
		}
		sources = append(sources, s)
	}

	return sources, rows.Err()
}

// GetEnabled returns all enabled sources
func (r *Repository) GetEnabled() ([]SourceRecord, error) {
	rows, err := r.db.Query(`
		SELECT id, type, name, url, config, enabled, is_builtin, created_at
		FROM sources
		WHERE enabled = true
		ORDER BY is_builtin DESC, name ASC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query enabled sources: %w", err)
	}
	defer rows.Close()

	var sources []SourceRecord
	for rows.Next() {
		var s SourceRecord
		var url, config sql.NullString
		if err := rows.Scan(&s.ID, &s.Type, &s.Name, &url, &config, &s.Enabled, &s.IsBuiltin, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan source: %w", err)
		}
		s.URL = url.String
		s.Config = config.String
		if s.Config == "" {
			s.Config = "{}"
		}
		sources = append(sources, s)
	}

	return sources, rows.Err()
}

// GetByID returns a source by ID
func (r *Repository) GetByID(id int64) (*SourceRecord, error) {
	var s SourceRecord
	var url, config sql.NullString

	err := r.db.QueryRow(`
		SELECT id, type, name, url, config, enabled, is_builtin, created_at
		FROM sources
		WHERE id = ?
	`, id).Scan(&s.ID, &s.Type, &s.Name, &url, &config, &s.Enabled, &s.IsBuiltin, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get source: %w", err)
	}

	s.URL = url.String
	s.Config = config.String
	if s.Config == "" {
		s.Config = "{}"
	}

	return &s, nil
}

// Create creates a new source
func (r *Repository) Create(s *SourceRecord) error {
	result, err := r.db.Exec(`
		INSERT INTO sources (type, name, url, config, enabled, is_builtin)
		VALUES (?, ?, ?, ?, ?, ?)
	`, s.Type, s.Name, s.URL, s.Config, s.Enabled, s.IsBuiltin)

	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	s.ID = id
	return nil
}

// Update updates an existing source
func (r *Repository) Update(s *SourceRecord) error {
	_, err := r.db.Exec(`
		UPDATE sources
		SET name = ?, url = ?, config = ?, enabled = ?
		WHERE id = ? AND is_builtin = false
	`, s.Name, s.URL, s.Config, s.Enabled, s.ID)

	if err != nil {
		return fmt.Errorf("failed to update source: %w", err)
	}

	return nil
}

// SetEnabled updates the enabled status of a source
func (r *Repository) SetEnabled(id int64, enabled bool) error {
	_, err := r.db.Exec(`UPDATE sources SET enabled = ? WHERE id = ?`, enabled, id)
	if err != nil {
		return fmt.Errorf("failed to update source enabled status: %w", err)
	}
	return nil
}

// Delete deletes a non-builtin source
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec(`DELETE FROM sources WHERE id = ? AND is_builtin = false`, id)
	if err != nil {
		return fmt.Errorf("failed to delete source: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("source not found or is builtin")
	}

	return nil
}

// CountByType returns the count of sources by type
func (r *Repository) CountByType(sourceType string) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sources WHERE type = ?`, sourceType).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count sources: %w", err)
	}
	return count, nil
}

// Seed creates default builtin sources if they don't exist
func (r *Repository) Seed() error {
	// Check if already seeded
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM sources WHERE is_builtin = true`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check sources: %w", err)
	}

	if count > 0 {
		return nil // Already seeded
	}

	defaultSources := []SourceRecord{
		{Type: "hackernews", Name: "Hacker News", Enabled: true, IsBuiltin: true, Config: "{}"},
		{Type: "github", Name: "GitHub Trending", Enabled: true, IsBuiltin: true, Config: "{}"},
		{Type: "npm", Name: "npm Registry", Enabled: true, IsBuiltin: true, Config: "{}"},
		{Type: "devto", Name: "DEV.to", Enabled: true, IsBuiltin: true, Config: "{}"},
	}

	for _, s := range defaultSources {
		if err := r.Create(&s); err != nil {
			return fmt.Errorf("failed to seed source %s: %w", s.Name, err)
		}
	}

	return nil
}

// ToConfig converts a SourceRecord to SourceConfig
func (s *SourceRecord) ToConfig() SourceConfig {
	var config map[string]string
	json.Unmarshal([]byte(s.Config), &config)
	if config == nil {
		config = make(map[string]string)
	}

	return SourceConfig{
		ID:        s.ID,
		Type:      s.Type,
		Name:      s.Name,
		URL:       s.URL,
		Config:    config,
		Enabled:   s.Enabled,
		IsBuiltin: s.IsBuiltin,
	}
}
