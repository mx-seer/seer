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
	Register("anthropic", NewAnthropic)
}

// Anthropic implements the Provider interface for Anthropic Claude
type Anthropic struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type anthropicRequest struct {
	Model     string    `json:"model"`
	MaxTokens int       `json:"max_tokens"`
	Messages  []Message `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewAnthropic creates a new Anthropic provider
func NewAnthropic(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("anthropic requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "claude-3-haiku-20240307"
	}

	return &Anthropic{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (a *Anthropic) Name() string {
	return "anthropic"
}

func (a *Anthropic) Available() bool {
	return a.apiKey != ""
}

func (a *Anthropic) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := anthropicRequest{
		Model:     a.model,
		MaxTokens: 4096,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("anthropic returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", err
	}

	if anthropicResp.Error != nil {
		return "", fmt.Errorf("anthropic error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return "", fmt.Errorf("anthropic returned no content")
	}

	return anthropicResp.Content[0].Text, nil
}
