//go:build !pro

package alerts

import (
	"context"
	"database/sql"
	"time"
)

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeEmail   AlertType = "email"
	AlertTypeWebhook AlertType = "webhook"
	AlertTypeSlack   AlertType = "slack"
)

// Alert represents an alert configuration (stub for CE)
type Alert struct {
	ID          int64             `json:"id"`
	Type        AlertType         `json:"type"`
	Name        string            `json:"name"`
	Destination string            `json:"destination"`
	MinScore    int               `json:"min_score"`
	Enabled     bool              `json:"enabled"`
	Config      map[string]string `json:"config,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// AlertPayload is the data sent when an alert triggers (stub for CE)
type AlertPayload struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Score       int       `json:"score"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	DetectedAt  time.Time `json:"detected_at"`
}

// AlertService manages alerts (stub for CE)
type AlertService struct {
	db *sql.DB
}

// NewAlertService creates a new alert service (stub for CE)
func NewAlertService(db *sql.DB) *AlertService {
	return &AlertService{db: db}
}

// GetAlerts returns empty list in CE edition
func (s *AlertService) GetAlerts() ([]Alert, error) {
	return []Alert{}, nil
}

// CreateAlert does nothing in CE edition
func (s *AlertService) CreateAlert(alert *Alert) error {
	return nil
}

// DeleteAlert does nothing in CE edition
func (s *AlertService) DeleteAlert(id int64) error {
	return nil
}

// ToggleAlert does nothing in CE edition
func (s *AlertService) ToggleAlert(id int64) error {
	return nil
}

// Send does nothing in CE edition
func (s *AlertService) Send(ctx context.Context, alert Alert, payload AlertPayload) error {
	return nil
}

// CheckAndSend does nothing in CE edition
func (s *AlertService) CheckAndSend(ctx context.Context, payload AlertPayload) error {
	return nil
}
