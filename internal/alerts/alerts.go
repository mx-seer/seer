//go:build pro

package alerts

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypeEmail   AlertType = "email"
	AlertTypeWebhook AlertType = "webhook"
	AlertTypeSlack   AlertType = "slack"
)

// Alert represents an alert configuration
type Alert struct {
	ID          int64             `json:"id"`
	Type        AlertType         `json:"type"`
	Name        string            `json:"name"`
	Destination string            `json:"destination"` // email, webhook URL, or slack webhook
	MinScore    int               `json:"min_score"`   // Minimum score to trigger
	Enabled     bool              `json:"enabled"`
	Config      map[string]string `json:"config,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
}

// AlertPayload is the data sent when an alert triggers
type AlertPayload struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Score       int       `json:"score"`
	Source      string    `json:"source"`
	URL         string    `json:"url"`
	DetectedAt  time.Time `json:"detected_at"`
}

// AlertService manages alerts
type AlertService struct {
	db     *sql.DB
	client *http.Client
}

// NewAlertService creates a new alert service
func NewAlertService(db *sql.DB) *AlertService {
	return &AlertService{
		db: db,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// GetAlerts returns all configured alerts
func (s *AlertService) GetAlerts() ([]Alert, error) {
	rows, err := s.db.Query(`
		SELECT id, type, name, destination, min_score, enabled, config, created_at
		FROM alerts
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		var configJSON string
		err := rows.Scan(&a.ID, &a.Type, &a.Name, &a.Destination, &a.MinScore, &a.Enabled, &configJSON, &a.CreatedAt)
		if err != nil {
			return nil, err
		}

		if configJSON != "" {
			json.Unmarshal([]byte(configJSON), &a.Config)
		}

		alerts = append(alerts, a)
	}

	return alerts, nil
}

// CreateAlert creates a new alert
func (s *AlertService) CreateAlert(alert *Alert) error {
	configJSON, err := json.Marshal(alert.Config)
	if err != nil {
		configJSON = []byte("{}")
	}

	result, err := s.db.Exec(`
		INSERT INTO alerts (type, name, destination, min_score, enabled, config)
		VALUES (?, ?, ?, ?, ?, ?)
	`, alert.Type, alert.Name, alert.Destination, alert.MinScore, alert.Enabled, string(configJSON))
	if err != nil {
		return err
	}

	alert.ID, _ = result.LastInsertId()
	return nil
}

// DeleteAlert deletes an alert
func (s *AlertService) DeleteAlert(id int64) error {
	_, err := s.db.Exec(`DELETE FROM alerts WHERE id = ?`, id)
	return err
}

// ToggleAlert enables/disables an alert
func (s *AlertService) ToggleAlert(id int64) error {
	_, err := s.db.Exec(`UPDATE alerts SET enabled = NOT enabled WHERE id = ?`, id)
	return err
}

// Send sends an alert for the given payload
func (s *AlertService) Send(ctx context.Context, alert Alert, payload AlertPayload) error {
	switch alert.Type {
	case AlertTypeWebhook:
		return s.sendWebhook(ctx, alert.Destination, payload)
	case AlertTypeSlack:
		return s.sendSlack(ctx, alert.Destination, payload)
	case AlertTypeEmail:
		// Email requires SMTP config - simplified for now
		return fmt.Errorf("email alerts not yet implemented")
	default:
		return fmt.Errorf("unknown alert type: %s", alert.Type)
	}
}

func (s *AlertService) sendWebhook(ctx context.Context, url string, payload AlertPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}

func (s *AlertService) sendSlack(ctx context.Context, webhookURL string, payload AlertPayload) error {
	slackPayload := map[string]any{
		"text": fmt.Sprintf("*New Opportunity Detected!*\n\n*%s*\n%s\n\nScore: %d | Source: %s\n<%s|View →>",
			payload.Title,
			payload.Description,
			payload.Score,
			payload.Source,
			payload.URL,
		),
	}

	body, err := json.Marshal(slackPayload)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, webhookURL, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("slack webhook returned status %d", resp.StatusCode)
	}

	return nil
}

// CheckAndSend checks if any alerts should be triggered for a new opportunity
func (s *AlertService) CheckAndSend(ctx context.Context, payload AlertPayload) error {
	alerts, err := s.GetAlerts()
	if err != nil {
		return err
	}

	for _, alert := range alerts {
		if !alert.Enabled {
			continue
		}

		if payload.Score >= alert.MinScore {
			if err := s.Send(ctx, alert, payload); err != nil {
				// Log error but continue with other alerts
				fmt.Printf("Failed to send alert %s: %v\n", alert.Name, err)
			}
		}
	}

	return nil
}
