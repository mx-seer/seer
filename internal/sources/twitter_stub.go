//go:build !pro

package sources

import "fmt"

// NewTwitter returns an error in CE edition
func NewTwitter(cfg SourceConfig) (Source, error) {
	return nil, fmt.Errorf("twitter source requires Pro edition")
}
