package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"
)

// OpportunityResponse represents an opportunity in API responses
type OpportunityResponse struct {
	ID               int64     `json:"id"`
	Title            string    `json:"title"`
	Description      string    `json:"description"`
	SourceType       string    `json:"source_type"`
	SourceURL        string    `json:"source_url"`
	SourceIDExternal string    `json:"source_id_external"`
	Score            int       `json:"score"`
	Signals          []string  `json:"signals"`
	DetectedAt       time.Time `json:"detected_at"`
	CreatedAt        time.Time `json:"created_at"`
}

// OpportunitiesHandler handles opportunity-related requests
type OpportunitiesHandler struct {
	db *sql.DB
}

// NewOpportunitiesHandler creates a new opportunities handler
func NewOpportunitiesHandler(db *sql.DB) *OpportunitiesHandler {
	return &OpportunitiesHandler{db: db}
}

// List returns opportunities with optional filters
func (h *OpportunitiesHandler) List(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	sourceType := r.URL.Query().Get("source")
	minScoreStr := r.URL.Query().Get("min_score")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	minScore := 0
	if minScoreStr != "" {
		if v, err := strconv.Atoi(minScoreStr); err == nil {
			minScore = v
		}
	}

	limit := 50
	if limitStr != "" {
		if v, err := strconv.Atoi(limitStr); err == nil && v > 0 && v <= 1000 {
			limit = v
		}
	}

	offset := 0
	if offsetStr != "" {
		if v, err := strconv.Atoi(offsetStr); err == nil && v >= 0 {
			offset = v
		}
	}

	// Build query
	query := `
		SELECT id, title, description, source, source_url, source_id_external, score, signals, detected_at, created_at
		FROM opportunities
		WHERE score >= ?
	`
	args := []any{minScore}

	if sourceType != "" {
		query += " AND source = ?"
		args = append(args, sourceType)
	}

	query += " ORDER BY score DESC, detected_at DESC LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := h.db.Query(query, args...)
	if err != nil {
		http.Error(w, "Failed to query opportunities", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var opportunities []OpportunityResponse
	for rows.Next() {
		var opp OpportunityResponse
		var signalsJSON string
		var description, sourceURL sql.NullString

		err := rows.Scan(
			&opp.ID, &opp.Title, &description, &opp.SourceType,
			&sourceURL, &opp.SourceIDExternal, &opp.Score,
			&signalsJSON, &opp.DetectedAt, &opp.CreatedAt,
		)
		if err != nil {
			continue
		}

		opp.Description = description.String
		opp.SourceURL = sourceURL.String

		// Parse signals JSON
		json.Unmarshal([]byte(signalsJSON), &opp.Signals)
		if opp.Signals == nil {
			opp.Signals = []string{}
		}

		opportunities = append(opportunities, opp)
	}

	if opportunities == nil {
		opportunities = []OpportunityResponse{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(opportunities)
}

// Get returns a single opportunity by ID
func (h *OpportunitiesHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var opp OpportunityResponse
	var signalsJSON string
	var description, sourceURL sql.NullString

	err = h.db.QueryRow(`
		SELECT id, title, description, source, source_url, source_id_external, score, signals, detected_at, created_at
		FROM opportunities
		WHERE id = ?
	`, id).Scan(
		&opp.ID, &opp.Title, &description, &opp.SourceType,
		&sourceURL, &opp.SourceIDExternal, &opp.Score,
		&signalsJSON, &opp.DetectedAt, &opp.CreatedAt,
	)

	if err == sql.ErrNoRows {
		http.Error(w, "Opportunity not found", http.StatusNotFound)
		return
	}
	if err != nil {
		http.Error(w, "Failed to get opportunity", http.StatusInternalServerError)
		return
	}

	opp.Description = description.String
	opp.SourceURL = sourceURL.String
	json.Unmarshal([]byte(signalsJSON), &opp.Signals)
	if opp.Signals == nil {
		opp.Signals = []string{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(opp)
}

// Stats returns opportunity statistics
func (h *OpportunitiesHandler) Stats(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters for filtering
	sourceType := r.URL.Query().Get("source")
	minScoreStr := r.URL.Query().Get("min_score")

	minScore := 0
	if minScoreStr != "" {
		if v, err := strconv.Atoi(minScoreStr); err == nil {
			minScore = v
		}
	}

	stats := struct {
		Total        int            `json:"total"`
		BySource     map[string]int `json:"by_source"`
		AverageScore float64        `json:"average_score"`
		Today        int            `json:"today"`
	}{
		BySource: make(map[string]int),
	}

	// Build WHERE clause
	whereClause := "WHERE score >= ?"
	args := []any{minScore}

	if sourceType != "" {
		whereClause += " AND source = ?"
		args = append(args, sourceType)
	}

	// Total count (filtered)
	h.db.QueryRow("SELECT COUNT(*) FROM opportunities "+whereClause, args...).Scan(&stats.Total)

	// Average score (filtered)
	h.db.QueryRow("SELECT COALESCE(AVG(score), 0) FROM opportunities "+whereClause, args...).Scan(&stats.AverageScore)

	// Today count (filtered) - opportunities detected in last 24 hours
	h.db.QueryRow("SELECT COUNT(*) FROM opportunities "+whereClause+" AND detected_at >= datetime('now', '-24 hours')", args...).Scan(&stats.Today)

	// By source (filtered)
	rows, err := h.db.Query("SELECT source, COUNT(*) FROM opportunities "+whereClause+" GROUP BY source", args...)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var source string
			var count int
			if rows.Scan(&source, &count) == nil {
				stats.BySource[source] = count
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}
