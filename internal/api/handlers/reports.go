package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"time"

	"github.com/mx-seer/seer/internal/report"
)

// ReportResponse represents a report in API responses
type ReportResponse struct {
	ID               int64     `json:"id"`
	PeriodStart      time.Time `json:"period_start"`
	PeriodEnd        time.Time `json:"period_end"`
	OpportunityCount int       `json:"opportunity_count"`
	ContentHuman     string    `json:"content_human"`
	ContentPrompt    string    `json:"content_prompt"`
	AIAnalysis       string    `json:"ai_analysis,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}

// ReportsHandler handles report-related requests
type ReportsHandler struct {
	db        *sql.DB
	generator *report.Generator
}

// NewReportsHandler creates a new reports handler
func NewReportsHandler(db *sql.DB) *ReportsHandler {
	return &ReportsHandler{
		db:        db,
		generator: report.New(),
	}
}

// Generate generates a new report
func (h *ReportsHandler) Generate(w http.ResponseWriter, r *http.Request) {
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

	// Generate report
	rep := h.generator.Generate(opportunities, periodStart, periodEnd)

	// Save to database
	result, err := h.db.Exec(`
		INSERT INTO reports (period_start, period_end, opportunity_count, content_human, content_prompt)
		VALUES (?, ?, ?, ?, ?)
	`, rep.PeriodStart, rep.PeriodEnd, rep.OpportunityCount, rep.ContentHuman, rep.ContentPrompt)

	var reportID int64
	if err == nil {
		reportID, _ = result.LastInsertId()
	}

	response := ReportResponse{
		ID:               reportID,
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

// List returns recent reports
func (h *ReportsHandler) List(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(`
		SELECT id, period_start, period_end, opportunity_count, created_at
		FROM reports
		ORDER BY created_at DESC
		LIMIT 20
	`)
	if err != nil {
		http.Error(w, "Failed to query reports", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []ReportResponse
	for rows.Next() {
		var rep ReportResponse
		err := rows.Scan(&rep.ID, &rep.PeriodStart, &rep.PeriodEnd, &rep.OpportunityCount, &rep.CreatedAt)
		if err != nil {
			continue
		}
		reports = append(reports, rep)
	}

	if reports == nil {
		reports = []ReportResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(reports)
}

// Get returns a single report by ID
func (h *ReportsHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int64
	if _, err := json.Number(idStr).Int64(); err == nil {
		id, _ = json.Number(idStr).Int64()
	}

	var rep ReportResponse
	var contentHuman, contentPrompt, aiAnalysis sql.NullString

	err := h.db.QueryRow(`
		SELECT id, period_start, period_end, opportunity_count, content_human, content_prompt, ai_analysis, created_at
		FROM reports
		WHERE id = ?
	`, id).Scan(
		&rep.ID, &rep.PeriodStart, &rep.PeriodEnd, &rep.OpportunityCount,
		&contentHuman, &contentPrompt, &aiAnalysis, &rep.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}

	rep.ContentHuman = contentHuman.String
	rep.ContentPrompt = contentPrompt.String
	rep.AIAnalysis = aiAnalysis.String

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rep)
}

// GetPrompt returns just the prompt content for copying
func (h *ReportsHandler) GetPrompt(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int64
	if _, err := json.Number(idStr).Int64(); err == nil {
		id, _ = json.Number(idStr).Int64()
	}

	var contentPrompt string
	err := h.db.QueryRow(`SELECT content_prompt FROM reports WHERE id = ?`, id).Scan(&contentPrompt)

	if err == sql.ErrNoRows {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get report", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(contentPrompt))
}
