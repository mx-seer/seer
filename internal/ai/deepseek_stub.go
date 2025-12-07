//go:build !pro

package ai

import "fmt"

func init() {
	Register("deepseek", NewDeepSeek)
}

// NewDeepSeek returns an error in CE edition
func NewDeepSeek(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("deepseek provider requires Pro edition")
}
