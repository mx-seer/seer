package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mx-seer/seer-pro/internal/api/handlers"
	"github.com/mx-seer/seer-pro/internal/db"
	"github.com/mx-seer/seer-pro/internal/sources"
)

// Server holds the HTTP server dependencies
type Server struct {
	db            *db.DB
	router        *chi.Mux
	sourceManager *sources.Manager
}

// NewServer creates a new API server
func NewServer(database *db.DB, sourceManager *sources.Manager) *Server {
	s := &Server{
		db:            database,
		router:        chi.NewRouter(),
		sourceManager: sourceManager,
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

// setupMiddleware configures common middleware
func (s *Server) setupMiddleware() {
	s.router.Use(middleware.RequestID)
	s.router.Use(middleware.RealIP)
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.Timeout(30 * time.Second))

	// CORS for development
	s.router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

// setupRoutes configures all API routes
func (s *Server) setupRoutes() {
	// Health check at root
	s.router.Get("/health", s.handleHealth)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		r.Get("/health", s.handleHealth)

		// Opportunities
		oppHandler := handlers.NewOpportunitiesHandler(s.db.DB)
		r.Get("/opportunities", oppHandler.List)
		r.Get("/opportunities/stats", oppHandler.Stats)
		r.Get("/opportunities/{id}", oppHandler.Get)

		// Sources
		srcHandler := handlers.NewSourcesHandler(s.db.DB)
		r.Get("/sources", srcHandler.List)
		r.Get("/sources/types", srcHandler.AvailableTypes)
		r.Get("/sources/{id}", srcHandler.Get)
		r.Post("/sources", srcHandler.Create)
		r.Put("/sources/{id}", srcHandler.Update)
		r.Delete("/sources/{id}", srcHandler.Delete)
		r.Post("/sources/{id}/toggle", srcHandler.Toggle)
		r.Post("/sources/fetch", s.handleFetchSources)

		// Prompts
		promptHandler := handlers.NewPromptsHandler(s.db.DB)
		r.Get("/prompts", promptHandler.List)
		r.Post("/prompts", promptHandler.Create)
		r.Post("/prompts/generate", promptHandler.Generate)
		r.Get("/prompts/{id}", promptHandler.Get)
		r.Get("/prompts/{id}/content", promptHandler.GetContent)
	})

	// Serve embedded frontend
	staticFS := StaticFS()
	fileServer := http.FileServer(staticFS)

	s.router.Get("/*", func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file directly
		path := r.URL.Path
		if path == "/" {
			path = "/index.html"
		}

		// Check if file exists
		f, err := staticFS.Open(path[1:]) // Remove leading /
		if err != nil {
			// File not found, serve index.html for SPA routing
			r.URL.Path = "/"
			fileServer.ServeHTTP(w, r)
			return
		}
		f.Close()

		fileServer.ServeHTTP(w, r)
	})
}

// Handler returns the HTTP handler
func (s *Server) Handler() http.Handler {
	return s.router
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Timestamp string `json:"timestamp"`
	Version   string `json:"version"`
}

// handleFetchSources triggers a manual fetch of all enabled sources
func (s *Server) handleFetchSources(w http.ResponseWriter, r *http.Request) {
	if s.sourceManager == nil {
		http.Error(w, "Source manager not available", http.StatusInternalServerError)
		return
	}

	// Run fetch in background and return immediately
	go s.sourceManager.FetchAll(r.Context())

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "started",
		"message": "Fetching opportunities from all enabled sources",
	})
}

// handleHealth handles the health check endpoint
func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "ok",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Version:   "1.0.0",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
