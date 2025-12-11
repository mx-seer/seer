package sources

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/mx-seer/seer-pro/internal/scoring"
	"github.com/robfig/cron/v3"
)

// Manager coordinates source fetching and scheduling
type Manager struct {
	db            *sql.DB
	repo          *Repository
	cron          *cron.Cron
	factories     map[string]SourceFactory
	scorer        *scoring.Scorer
	mu            sync.RWMutex
	isRunning     bool
	fetchInterval int // Interval in minutes between fetches
}

// NewManager creates a new source manager
// fetchIntervalMinutes is the interval in minutes between fetches (default: 60)
func NewManager(db *sql.DB, fetchIntervalMinutes int) *Manager {
	if fetchIntervalMinutes <= 0 {
		fetchIntervalMinutes = 60 // Default to 1 hour
	}
	m := &Manager{
		db:            db,
		repo:          NewRepository(db),
		cron:          cron.New(),
		factories:     make(map[string]SourceFactory),
		scorer:        scoring.New(),
		fetchInterval: fetchIntervalMinutes,
	}

	// Register default factories
	m.RegisterFactory("hackernews", NewHackerNews)
	m.RegisterFactory("github", NewGitHub)
	m.RegisterFactory("npm", NewNPM)
	m.RegisterFactory("devto", NewDevTo)

	return m
}

// RegisterFactory registers a source factory
func (m *Manager) RegisterFactory(sourceType string, factory SourceFactory) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.factories[sourceType] = factory
}

// Start starts the scheduler
func (m *Manager) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.isRunning {
		return nil
	}

	// Seed default sources
	if err := m.repo.Seed(); err != nil {
		return fmt.Errorf("failed to seed sources: %w", err)
	}

	// Schedule fetching at configured interval
	cronExpr := fmt.Sprintf("@every %dm", m.fetchInterval)
	_, err := m.cron.AddFunc(cronExpr, func() {
		if err := m.FetchAll(context.Background()); err != nil {
			log.Printf("Error fetching sources: %v", err)
		}
	})
	log.Printf("Source fetch scheduled every %d minutes", m.fetchInterval)
	if err != nil {
		return fmt.Errorf("failed to schedule fetch job: %w", err)
	}

	m.cron.Start()
	m.isRunning = true

	log.Println("Source manager started")

	// Run initial fetch in background
	go func() {
		if err := m.FetchAll(context.Background()); err != nil {
			log.Printf("Initial fetch error: %v", err)
		}
	}()

	return nil
}

// Stop stops the scheduler
func (m *Manager) Stop() {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isRunning {
		return
	}

	ctx := m.cron.Stop()
	<-ctx.Done()
	m.isRunning = false

	log.Println("Source manager stopped")
}

// FetchAll fetches opportunities from all enabled sources
func (m *Manager) FetchAll(ctx context.Context) error {
	sources, err := m.repo.GetEnabled()
	if err != nil {
		return fmt.Errorf("failed to get enabled sources: %w", err)
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(sources))

	for _, src := range sources {
		wg.Add(1)
		go func(s SourceRecord) {
			defer wg.Done()
			if err := m.fetchSource(ctx, s); err != nil {
				errChan <- fmt.Errorf("source %s: %w", s.Name, err)
			}
		}(src)
	}

	wg.Wait()
	close(errChan)

	// Collect errors
	var errs []error
	for err := range errChan {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		log.Printf("Fetch completed with %d errors", len(errs))
		for _, e := range errs {
			log.Printf("  - %v", e)
		}
	}

	return nil
}

// fetchSource fetches opportunities from a single source
func (m *Manager) fetchSource(ctx context.Context, record SourceRecord) error {
	m.mu.RLock()
	factory, ok := m.factories[record.Type]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown source type: %s", record.Type)
	}

	cfg := record.ToConfig()
	source, err := factory(cfg)
	if err != nil {
		return fmt.Errorf("failed to create source: %w", err)
	}

	opportunities, err := source.Fetch(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch: %w", err)
	}

	// Save opportunities to database
	for _, opp := range opportunities {
		if err := m.saveOpportunity(record.ID, opp); err != nil {
			log.Printf("Failed to save opportunity %s: %v", opp.Title, err)
		}
	}

	log.Printf("Fetched %d opportunities from %s", len(opportunities), record.Name)
	return nil
}

// saveOpportunity saves an opportunity to the database
func (m *Manager) saveOpportunity(sourceID int64, opp Opportunity) error {
	// Convert to scoring.Opportunity for scoring
	scoringOpp := scoring.Opportunity{
		Title:       opp.Title,
		Description: opp.Description,
		SourceType:  opp.SourceType,
		DetectedAt:  opp.DetectedAt,
		Metadata:    opp.Metadata,
	}

	// Calculate score
	result := m.scorer.Score(scoringOpp)

	// Get matched signal names
	matchedSignals := result.GetMatchedSignals()
	signalNames := make([]string, len(matchedSignals))
	for i, sig := range matchedSignals {
		signalNames[i] = sig.Name
	}
	signalsJSON, _ := json.Marshal(signalNames)

	_, err := m.db.Exec(`
		INSERT INTO opportunities (source_id, title, description, source, source_url, source_id_external, score, signals, detected_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(source, source_id_external) DO UPDATE SET
			title = excluded.title,
			description = excluded.description,
			source_url = excluded.source_url,
			score = excluded.score,
			signals = excluded.signals,
			detected_at = excluded.detected_at
	`, sourceID, opp.Title, opp.Description, opp.SourceType, opp.SourceURL, opp.SourceIDExternal, result.Score, string(signalsJSON), opp.DetectedAt)

	if err != nil {
		return fmt.Errorf("failed to insert opportunity: %w", err)
	}

	return nil
}

// GetRepository returns the source repository
func (m *Manager) GetRepository() *Repository {
	return m.repo
}
