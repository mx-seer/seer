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
	Register("deepseek", NewDeepSeek)
}

// DeepSeek implements the Provider interface for DeepSeek
type DeepSeek struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// DeepSeek uses OpenAI-compatible API format
type deepSeekRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type deepSeekResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewDeepSeek creates a new DeepSeek provider
func NewDeepSeek(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("deepseek requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}

	model := cfg.Model
	if model == "" {
		model = "deepseek-chat"
	}

	return &DeepSeek{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (d *DeepSeek) Name() string {
	return "deepseek"
}

func (d *DeepSeek) Available() bool {
	return d.apiKey != ""
}

func (d *DeepSeek) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := deepSeekRequest{
		Model: d.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, d.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+d.apiKey)

	resp, err := d.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("deepseek returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var deepSeekResp deepSeekResponse
	if err := json.NewDecoder(resp.Body).Decode(&deepSeekResp); err != nil {
		return "", err
	}

	if deepSeekResp.Error != nil {
		return "", fmt.Errorf("deepseek error: %s", deepSeekResp.Error.Message)
	}

	if len(deepSeekResp.Choices) == 0 {
		return "", fmt.Errorf("deepseek returned no choices")
	}

	return deepSeekResp.Choices[0].Message.Content, nil
}
