package ai

import (
	"context"
	"fmt"
)

// Provider defines the interface for AI providers
type Provider interface {
	// Name returns the provider name
	Name() string

	// Analyze sends a prompt to the AI and returns the analysis
	Analyze(ctx context.Context, prompt string) (string, error)

	// Available checks if the provider is properly configured
	Available() bool
}

// ProviderConfig holds configuration for an AI provider
type ProviderConfig struct {
	Type     string            `json:"type"`
	APIKey   string            `json:"api_key,omitempty"`
	BaseURL  string            `json:"base_url,omitempty"`
	Model    string            `json:"model,omitempty"`
	Options  map[string]string `json:"options,omitempty"`
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ProviderFactory creates a Provider from configuration
type ProviderFactory func(cfg ProviderConfig) (Provider, error)

// registry holds all registered provider factories
var registry = make(map[string]ProviderFactory)

// Register registers a provider factory
func Register(name string, factory ProviderFactory) {
	registry[name] = factory
}

// New creates a new provider from configuration
func New(cfg ProviderConfig) (Provider, error) {
	factory, ok := registry[cfg.Type]
	if !ok {
		return nil, fmt.Errorf("unknown AI provider: %s", cfg.Type)
	}
	return factory(cfg)
}

// AvailableProviders returns a list of registered provider types
func AvailableProviders() []string {
	var providers []string
	for name := range registry {
		providers = append(providers, name)
	}
	return providers
}
