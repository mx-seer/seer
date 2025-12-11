package handlers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/mx-seer/seer-pro/internal/sources"
)

// SourceResponse represents a source in API responses
type SourceResponse struct {
	ID        int64     `json:"id"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	URL       string    `json:"url,omitempty"`
	Enabled   bool      `json:"enabled"`
	IsBuiltin bool      `json:"is_builtin"`
	CreatedAt time.Time `json:"created_at"`
}

// SourceRequest represents a request to create/update a source
type SourceRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	Enabled *bool  `json:"enabled,omitempty"`
}

// SourcesHandler handles source-related requests
type SourcesHandler struct {
	repo *sources.Repository
}

// NewSourcesHandler creates a new sources handler
func NewSourcesHandler(db *sql.DB) *SourcesHandler {
	return &SourcesHandler{
		repo: sources.NewRepository(db),
	}
}

// List returns all sources
func (h *SourcesHandler) List(w http.ResponseWriter, r *http.Request) {
	records, err := h.repo.GetAll()
	if err != nil {
		http.Error(w, "Failed to get sources", http.StatusInternalServerError)
		return
	}

	response := make([]SourceResponse, len(records))
	for i, rec := range records {
		response[i] = SourceResponse{
			ID:        rec.ID,
			Type:      rec.Type,
			Name:      rec.Name,
			URL:       rec.URL,
			Enabled:   rec.Enabled,
			IsBuiltin: rec.IsBuiltin,
			CreatedAt: rec.CreatedAt,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Get returns a single source by ID
func (h *SourcesHandler) Get(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	rec, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Failed to get source", http.StatusInternalServerError)
		return
	}
	if rec == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	response := SourceResponse{
		ID:        rec.ID,
		Type:      rec.Type,
		Name:      rec.Name,
		URL:       rec.URL,
		Enabled:   rec.Enabled,
		IsBuiltin: rec.IsBuiltin,
		CreatedAt: rec.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Create creates a new source
func (h *SourcesHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req SourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate type
	availableTypes := sources.GetAvailableTypes()
	validType := false
	for _, t := range availableTypes {
		if t == req.Type {
			validType = true
			break
		}
	}
	if !validType {
		http.Error(w, "Invalid source type", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}

	record := &sources.SourceRecord{
		Type:      req.Type,
		Name:      req.Name,
		URL:       req.URL,
		Enabled:   enabled,
		IsBuiltin: false,
		Config:    "{}",
	}

	if err := h.repo.Create(record); err != nil {
		http.Error(w, "Failed to create source", http.StatusInternalServerError)
		return
	}

	response := SourceResponse{
		ID:        record.ID,
		Type:      record.Type,
		Name:      record.Name,
		URL:       record.URL,
		Enabled:   record.Enabled,
		IsBuiltin: record.IsBuiltin,
		CreatedAt: record.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// Update updates an existing source
func (h *SourcesHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Failed to get source", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	if existing.IsBuiltin {
		http.Error(w, "Cannot modify builtin sources", http.StatusForbidden)
		return
	}

	var req SourceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Update fields
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.URL != "" {
		existing.URL = req.URL
	}
	if req.Enabled != nil {
		existing.Enabled = *req.Enabled
	}

	if err := h.repo.Update(existing); err != nil {
		http.Error(w, "Failed to update source", http.StatusInternalServerError)
		return
	}

	response := SourceResponse{
		ID:        existing.ID,
		Type:      existing.Type,
		Name:      existing.Name,
		URL:       existing.URL,
		Enabled:   existing.Enabled,
		IsBuiltin: existing.IsBuiltin,
		CreatedAt: existing.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Delete deletes a source
func (h *SourcesHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Failed to get source", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	if existing.IsBuiltin {
		http.Error(w, "Cannot delete builtin sources", http.StatusForbidden)
		return
	}

	if err := h.repo.Delete(id); err != nil {
		http.Error(w, "Failed to delete source", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Toggle toggles the enabled status of a source
func (h *SourcesHandler) Toggle(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(id)
	if err != nil {
		http.Error(w, "Failed to get source", http.StatusInternalServerError)
		return
	}
	if existing == nil {
		http.Error(w, "Source not found", http.StatusNotFound)
		return
	}

	newEnabled := !existing.Enabled
	if err := h.repo.SetEnabled(id, newEnabled); err != nil {
		http.Error(w, "Failed to toggle source", http.StatusInternalServerError)
		return
	}

	existing.Enabled = newEnabled
	response := SourceResponse{
		ID:        existing.ID,
		Type:      existing.Type,
		Name:      existing.Name,
		URL:       existing.URL,
		Enabled:   existing.Enabled,
		IsBuiltin: existing.IsBuiltin,
		CreatedAt: existing.CreatedAt,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// AvailableTypes returns the available source types
func (h *SourcesHandler) AvailableTypes(w http.ResponseWriter, r *http.Request) {
	types := sources.GetAvailableTypes()

	response := struct {
		Types []string `json:"types"`
	}{
		Types: types,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
