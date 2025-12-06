package sources

import (
	"context"
	"time"
)

// Opportunity represents a detected market opportunity
type Opportunity struct {
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	SourceType       string    `json:"source_type"`
	SourceURL        string    `json:"source_url"`
	SourceIDExternal string    `json:"source_id_external"`
	DetectedAt       time.Time `json:"detected_at"`
	Metadata         map[string]any `json:"metadata,omitempty"`
}

// Source defines the interface all sources must implement
type Source interface {
	// Type returns the source type identifier
	Type() string

	// Name returns the human-readable source name
	Name() string

	// Fetch retrieves opportunities from the source
	Fetch(ctx context.Context) ([]Opportunity, error)
}

// SourceConfig holds configuration for a source instance
type SourceConfig struct {
	ID        int64             `json:"id"`
	Type      string            `json:"type"`
	Name      string            `json:"name"`
	URL       string            `json:"url,omitempty"`
	Config    map[string]string `json:"config,omitempty"`
	Enabled   bool              `json:"enabled"`
	IsBuiltin bool              `json:"is_builtin"`
}

// SourceFactory creates a Source from configuration
type SourceFactory func(cfg SourceConfig) (Source, error)
