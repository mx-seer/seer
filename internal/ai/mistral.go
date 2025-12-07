//go:build pro

package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

func init() {
	Register("mistral", NewMistral)
}

// Mistral implements the Provider interface for Mistral AI
type Mistral struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// Mistral uses OpenAI-compatible API format
type mistralRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type mistralResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewMistral creates a new Mistral provider
func NewMistral(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("mistral requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.mistral.ai/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "mistral-small-latest"
	}

	return &Mistral{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (m *Mistral) Name() string {
	return "mistral"
}

func (m *Mistral) Available() bool {
	return m.apiKey != ""
}

func (m *Mistral) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := mistralRequest{
		Model: m.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, m.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+m.apiKey)

	resp, err := m.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("mistral returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var mistralResp mistralResponse
	if err := json.NewDecoder(resp.Body).Decode(&mistralResp); err != nil {
		return "", err
	}

	if mistralResp.Error != nil {
		return "", fmt.Errorf("mistral error: %s", mistralResp.Error.Message)
	}

	if len(mistralResp.Choices) == 0 {
		return "", fmt.Errorf("mistral returned no choices")
	}

	return mistralResp.Choices[0].Message.Content, nil
}
