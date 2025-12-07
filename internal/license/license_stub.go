//go:build !pro

package license

import (
	"context"
	"database/sql"
	"time"
)

// License represents a Pro license (stub for CE)
type License struct {
	Key        string    `json:"key"`
	Email      string    `json:"email"`
	Tier       string    `json:"tier"`
	ExpiresAt  time.Time `json:"expires_at"`
	VerifiedAt time.Time `json:"verified_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// VerifyResponse is the response from the license server (stub for CE)
type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	Email     string    `json:"email"`
	Tier      string    `json:"tier"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message,omitempty"`
}

// Service manages license operations (stub for CE)
type Service struct {
	db *sql.DB
}

// NewService creates a new license service (stub for CE)
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// GetLicense returns nil in CE edition
func (s *Service) GetLicense() (*License, error) {
	return nil, nil
}

// Activate does nothing in CE edition
func (s *Service) Activate(ctx context.Context, key string) (*License, error) {
	return nil, nil
}

// IsValid returns false in CE edition (no Pro features)
func (s *Service) IsValid() bool {
	return false
}

// Ping does nothing in CE edition
func (s *Service) Ping(ctx context.Context) error {
	return nil
}

// StartPingLoop does nothing in CE edition
func (s *Service) StartPingLoop(ctx context.Context) {
	// No-op in CE
}

// Deactivate does nothing in CE edition
func (s *Service) Deactivate() error {
	return nil
}
