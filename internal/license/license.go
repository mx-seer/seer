package license

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	verifyURL = "https://license.mendex.io/api/verify"
	pingURL   = "https://license.mendex.io/api/ping"
)

// License represents a Pro license
type License struct {
	Key        string    `json:"key"`
	Email      string    `json:"email"`
	Tier       string    `json:"tier"` // "monthly" or "annual"
	ExpiresAt  time.Time `json:"expires_at"`
	VerifiedAt time.Time `json:"verified_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// VerifyResponse is the response from the license server
type VerifyResponse struct {
	Valid     bool      `json:"valid"`
	Email     string    `json:"email"`
	Tier      string    `json:"tier"`
	ExpiresAt time.Time `json:"expires_at"`
	Message   string    `json:"message,omitempty"`
}

// Service manages license operations
type Service struct {
	db     *sql.DB
	client *http.Client
}

// NewService creates a new license service
func NewService(db *sql.DB) *Service {
	return &Service{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetLicense returns the current license
func (s *Service) GetLicense() (*License, error) {
	var lic License
	var expiresAt, verifiedAt, createdAt sql.NullTime

	err := s.db.QueryRow(`
		SELECT key, email, tier, expires_at, verified_at, created_at
		FROM license
		LIMIT 1
	`).Scan(&lic.Key, &lic.Email, &lic.Tier, &expiresAt, &verifiedAt, &createdAt)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	if expiresAt.Valid {
		lic.ExpiresAt = expiresAt.Time
	}
	if verifiedAt.Valid {
		lic.VerifiedAt = verifiedAt.Time
	}
	if createdAt.Valid {
		lic.CreatedAt = createdAt.Time
	}

	return &lic, nil
}

// Activate activates a license key
func (s *Service) Activate(ctx context.Context, key string) (*License, error) {
	// Verify with license server
	resp, err := s.verify(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to verify license: %w", err)
	}

	if !resp.Valid {
		return nil, fmt.Errorf("invalid license: %s", resp.Message)
	}

	// Save to database
	lic := &License{
		Key:        key,
		Email:      resp.Email,
		Tier:       resp.Tier,
		ExpiresAt:  resp.ExpiresAt,
		VerifiedAt: time.Now(),
		CreatedAt:  time.Now(),
	}

	_, err = s.db.Exec(`
		INSERT OR REPLACE INTO license (key, email, tier, expires_at, verified_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`, lic.Key, lic.Email, lic.Tier, lic.ExpiresAt, lic.VerifiedAt, lic.CreatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to save license: %w", err)
	}

	return lic, nil
}

// verify sends a verification request to the license server
func (s *Service) verify(ctx context.Context, key string) (*VerifyResponse, error) {
	body, err := json.Marshal(map[string]string{"key": key})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, verifyURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var verifyResp VerifyResponse
	if err := json.NewDecoder(resp.Body).Decode(&verifyResp); err != nil {
		return nil, err
	}

	return &verifyResp, nil
}

// IsValid checks if the current license is valid
func (s *Service) IsValid() bool {
	lic, err := s.GetLicense()
	if err != nil || lic == nil {
		return false
	}

	// Check expiration
	if time.Now().After(lic.ExpiresAt) {
		return false
	}

	return true
}

// Ping sends a heartbeat to the license server (for analytics)
func (s *Service) Ping(ctx context.Context) error {
	lic, err := s.GetLicense()
	if err != nil || lic == nil {
		return nil // No license, skip ping
	}

	body, err := json.Marshal(map[string]string{
		"key":     lic.Key,
		"version": "1.0.0",
	})
	if err != nil {
		return nil
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, pingURL, bytes.NewReader(body))
	if err != nil {
		return nil
	}
	req.Header.Set("Content-Type", "application/json")

	// Fire and forget - don't wait for response
	go func() {
		resp, err := s.client.Do(req)
		if err == nil {
			resp.Body.Close()
		}
	}()

	return nil
}

// StartPingLoop starts a background loop to periodically ping the license server
func (s *Service) StartPingLoop(ctx context.Context) {
	ticker := time.NewTicker(24 * time.Hour)
	defer ticker.Stop()

	// Initial ping
	s.Ping(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.Ping(ctx)
		}
	}
}

// Deactivate removes the license
func (s *Service) Deactivate() error {
	_, err := s.db.Exec(`DELETE FROM license`)
	return err
}
