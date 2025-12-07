//go:build !pro

package ai

import "fmt"

func init() {
	Register("ollama", NewOllama)
}

// NewOllama returns an error in CE edition
func NewOllama(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("ollama provider requires Pro edition")
}
