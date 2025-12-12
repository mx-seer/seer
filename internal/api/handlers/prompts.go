package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mx-seer/seer/internal/report"
)

// PromptResponse represents a prompt in API responses
type PromptResponse struct {
	ID               int64     `json:"id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	OpportunityCount int       `json:"opportunity_count"`
	ContentHuman     string    `json:"content_human"`
	ContentPrompt    string    `json:"content_prompt"`
	AIAnalysis       string    `json:"ai_analysis,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// PromptsHandler handles prompt-related requests
type PromptsHandler struct {
	db        *sql.DB
	generator *report.Generator
}

// NewPromptsHandler creates a new prompts handler
func NewPromptsHandler(db *sql.DB) *PromptsHandler {
	return &PromptsHandler{
		db:        db,
		generator: report.New(),
	}
}

// Generate generates a new prompt
func (h *PromptsHandler) Generate(w http.ResponseWriter, r *http.Request) {
	// Default to last 24 hours
	periodEnd := time.Now()
	periodStart := periodEnd.Add(-24 * time.Hour)

	// Allow custom period via query params
	if startStr := r.URL.Query().Get("start"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			periodStart = t
		}
	}
	if endStr := r.URL.Query().Get("end"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			periodEnd = t.Add(24*time.Hour - time.Second) // End of day
		}
	}

	// Get opportunities from the period
	rows, err := h.db.Query(`
		SELECT id, title, description, source, source_url, score, signals, detected_at
		FROM opportunities
		WHERE detected_at BETWEEN ? AND ?
		ORDER BY score DESC
	`, periodStart, periodEnd)
	if err != nil {
		http.Error(w, "Failed to query opportunities", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var opportunities []report.Opportunity
	for rows.Next() {
		var opp report.Opportunity
		var signalsJSON string
		var description, sourceURL sql.NullString

		err := rows.Scan(
			&opp.ID, &opp.Title, &description, &opp.SourceType,
			&sourceURL, &opp.Score, &signalsJSON, &opp.DetectedAt,
		)
		if err != nil {
			continue
		}

		opp.Description = description.String
		opp.SourceURL = sourceURL.String
		json.Unmarshal([]byte(signalsJSON), &opp.Signals)

		opportunities = append(opportunities, opp)
	}

	// Generate prompt
	rep := h.generator.Generate(opportunities, periodStart, periodEnd)

	// Save to database (using reports table for backward compatibility)
	result, err := h.db.Exec(`
		INSERT INTO reports (period_start, period_end, opportunity_count, content_human, content_prompt)
		VALUES (?, ?, ?, ?, ?)
	`, rep.PeriodStart, rep.PeriodEnd, rep.OpportunityCount, rep.ContentHuman, rep.ContentPrompt)

	var promptID int64
	if err == nil {
		promptID, _ = result.LastInsertId()
	}

	response := PromptResponse{
		ID:               promptID,
		PeriodStart:      rep.PeriodStart,
		PeriodEnd:        rep.PeriodEnd,
		OpportunityCount: rep.OpportunityCount,
		ContentHuman:     rep.ContentHuman,
		ContentPrompt:    rep.ContentPrompt,
		CreatedAt:        time.Now(),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// List returns recent prompts
func (h *PromptsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, period_start, period_end, opportunity_count, created_at
		FROM reports
		ORDER BY created_at DESC
		LIMIT 20
	`)
	if err != nil {
		http.Error(w, "Failed to query prompts", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var prompts []PromptResponse
	for rows.Next() {
		var p PromptResponse
		err := rows.Scan(&p.ID, &p.PeriodStart, &p.PeriodEnd, &p.OpportunityCount, &p.CreatedAt)
		if err != nil {
			continue
		}
		prompts = append(prompts, p)
	}

	if prompts == nil {
		prompts = []PromptResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(prompts)
}

// Get returns a single prompt by ID
func (h *PromptsHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int64
	if _, err := json.Number(idStr).Int64(); err == nil {
		id, _ = json.Number(idStr).Int64()
	}

	var p PromptResponse
	var contentHuman, contentPrompt, aiAnalysis sql.NullString

	err := h.db.QueryRow(`
		SELECT id, period_start, period_end, opportunity_count, content_human, content_prompt, ai_analysis, created_at
		FROM reports
		WHERE id = ?
	`, id).Scan(
		&p.ID, &p.PeriodStart, &p.PeriodEnd, &p.OpportunityCount,
		&contentHuman, &contentPrompt, &aiAnalysis, &p.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Prompt not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get prompt", http.StatusInternalServerError)
		return
	}

	p.ContentHuman = contentHuman.String
	p.ContentPrompt = contentPrompt.String
	p.AIAnalysis = aiAnalysis.String

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// GetContent returns just the prompt content for copying
func (h *PromptsHandler) GetContent(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int64
	if _, err := json.Number(idStr).Int64(); err == nil {
		id, _ = json.Number(idStr).Int64()
	}

	var contentPrompt string
	err := h.db.QueryRow(`SELECT content_prompt FROM reports WHERE id = ?`, id).Scan(&contentPrompt)

	if err == sql.ErrNoRows {
		http.Error(w, "Prompt not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get prompt", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(contentPrompt))
}

// CreatePromptRequest represents the request body for creating a prompt
type CreatePromptRequest struct {
	OpportunityCount int    `json:"opportunity_count"`
	ContentPrompt    string `json:"content_prompt"`
}

// Create creates a new prompt
func (h *PromptsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePromptRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ContentPrompt == "" {
		http.Error(w, "content_prompt is required", http.StatusBadRequest)
		return
	}

	now := time.Now()

	result, err := h.db.Exec(`
		INSERT INTO reports (period_start, period_end, opportunity_count, content_prompt)
		VALUES (?, ?, ?, ?)
	`, now, now, req.OpportunityCount, req.ContentPrompt)

	if err != nil {
		http.Error(w, "Failed to create prompt", http.StatusInternalServerError)
		return
	}

	promptID, _ := result.LastInsertId()

	response := PromptResponse{
		ID:               promptID,
		PeriodStart:      now,
		PeriodEnd:        now,
		OpportunityCount: req.OpportunityCount,
		ContentPrompt:    req.ContentPrompt,
		CreatedAt:        now,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}
