//go:build !pro

package ai

import "fmt"

func init() {
	Register("mistral", NewMistral)
}

// NewMistral returns an error in CE edition
func NewMistral(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("mistral provider requires Pro edition")
}
