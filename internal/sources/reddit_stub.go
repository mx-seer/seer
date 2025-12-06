//go:build !pro

package sources

import "fmt"

// NewReddit returns an error in CE edition
func NewReddit(cfg SourceConfig) (Source, error) {
	return nil, fmt.Errorf("reddit source requires Pro edition")
}
