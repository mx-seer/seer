//go:build !pro

package ai

import "fmt"

func init() {
	Register("anthropic", NewAnthropic)
}

// NewAnthropic returns an error in CE edition
func NewAnthropic(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("anthropic provider requires Pro edition")
}
