package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDefault(t *testing.T) {
	cfg := Default()

	if cfg.Server.Host != "0.0.0.0" {
		t.Errorf("expected host 0.0.0.0, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != 8080 {
		t.Errorf("expected port 8080, got %d", cfg.Server.Port)
	}

	if cfg.Database.Path != "./data/seer.db" {
		t.Errorf("expected database path ./data/seer.db, got %s", cfg.Database.Path)
	}

	if cfg.AI.Provider != "" {
		t.Errorf("expected empty AI provider, got %s", cfg.AI.Provider)
	}
}

func TestLoad_NonExistent(t *testing.T) {
	cfg, err := Load("nonexistent.yaml")
	if err != nil {
		t.Fatalf("expected no error for nonexistent file, got %v", err)
	}

	// Should return defaults
	if cfg.Server.Port != 8080 {
		t.Errorf("expected default port 8080, got %d", cfg.Server.Port)
	}
}

func TestLoad_ValidFile(t *testing.T) {
	// Create temp config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `
server:
  host: "127.0.0.1"
  port: 3000
database:
  path: "/tmp/test.db"
ai:
  provider: "ollama"
  endpoint: "http://localhost:11434"
  model: "llama3.2"
`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	cfg, err := Load(configPath)
	if err != nil {
		t.Fatalf("failed to load config: %v", err)
	}

	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}

	if cfg.Server.Port != 3000 {
		t.Errorf("expected port 3000, got %d", cfg.Server.Port)
	}

	if cfg.Database.Path != "/tmp/test.db" {
		t.Errorf("expected database path /tmp/test.db, got %s", cfg.Database.Path)
	}

	if cfg.AI.Provider != "ollama" {
		t.Errorf("expected AI provider ollama, got %s", cfg.AI.Provider)
	}

	if cfg.AI.Model != "llama3.2" {
		t.Errorf("expected AI model llama3.2, got %s", cfg.AI.Model)
	}
}

func TestLoad_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")

	content := `invalid: yaml: content:`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp config: %v", err)
	}

	_, err := Load(configPath)
	if err == nil {
		t.Error("expected error for invalid YAML, got nil")
	}
}

func TestAddress(t *testing.T) {
	cfg := &Config{
		Server: ServerConfig{
			Host: "localhost",
			Port: 9000,
		},
	}

	expected := "localhost:9000"
	if cfg.Address() != expected {
		t.Errorf("expected %s, got %s", expected, cfg.Address())
	}
}
