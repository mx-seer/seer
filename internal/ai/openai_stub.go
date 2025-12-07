//go:build !pro

package ai

import "fmt"

func init() {
	Register("openai", NewOpenAI)
}

// NewOpenAI returns an error in CE edition
func NewOpenAI(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("openai provider requires Pro edition")
}
