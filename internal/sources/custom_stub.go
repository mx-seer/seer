//go:build !pro

package sources

import "fmt"

// NewCustom returns an error in CE edition
func NewCustom(cfg SourceConfig) (Source, error) {
	return nil, fmt.Errorf("custom API source requires Pro edition")
}
