//go:build !pro

package ai

import "fmt"

func init() {
	Register("google", NewGoogle)
}

// NewGoogle returns an error in CE edition
func NewGoogle(cfg ProviderConfig) (Provider, error) {
	return nil, fmt.Errorf("google provider requires Pro edition")
}
