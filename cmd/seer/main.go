package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/mx-seer/seer/internal/api"
	"github.com/mx-seer/seer/internal/config"
	"github.com/mx-seer/seer/internal/db"
	"github.com/mx-seer/seer/internal/sources"
)

func main() {
	// Parse command line flags
	configPath := flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	log.Printf("Starting Seer on %s", cfg.Address())

	// Initialize database
	database, err := db.New(cfg.Database.Path)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.Close()

	log.Println("Database initialized")

	// Initialize source manager
	sourceManager := sources.NewManager(database.DB)
	if err := sourceManager.Start(); err != nil {
		log.Fatalf("Failed to start source manager: %v", err)
	}
	defer sourceManager.Stop()

	// Create API server
	server := api.NewServer(database)

	// Start HTTP server
	go func() {
		log.Printf("Server listening on %s", cfg.Address())
		if err := http.ListenAndServe(cfg.Address(), server.Handler()); err != nil {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
