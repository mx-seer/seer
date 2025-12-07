//go:build !pro

package ai

import "fmt"

func init() {
	Register("openrouter", NewOpenRouter)
}

// NewOpenRouter returns an error in CE edition
func NewOpenRouter(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("openrouter provider requires Pro edition")
}
