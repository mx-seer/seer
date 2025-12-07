//go:build !pro

package ai

import "fmt"

func init() {
	Register("groq", NewGroq)
}

// NewGroq returns an error in CE edition
func NewGroq(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("groq provider requires Pro edition")
}
