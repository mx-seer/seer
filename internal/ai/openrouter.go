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
	Register("openrouter", NewOpenRouter)
}

// OpenRouter implements the Provider interface for OpenRouter
type OpenRouter struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// OpenRouter uses OpenAI-compatible API format
type openRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewOpenRouter creates a new OpenRouter provider
func NewOpenRouter(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openrouter requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://openrouter.ai/api/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "meta-llama/llama-3.2-3b-instruct:free"
	}

	return &OpenRouter{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (o *OpenRouter) Name() string {
	return "openrouter"
}

func (o *OpenRouter) Available() bool {
	return o.apiKey != ""
}

func (o *OpenRouter) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := openRouterRequest{
		Model: o.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("HTTP-Referer", "https://seer.mendex.io")
	req.Header.Set("X-Title", "Seer")

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openrouter returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openRouterResp openRouterResponse
	if err := json.NewDecoder(resp.Body).Decode(&openRouterResp); err != nil {
		return "", err
	}

	if openRouterResp.Error != nil {
		return "", fmt.Errorf("openrouter error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return "", fmt.Errorf("openrouter returned no choices")
	}

	return openRouterResp.Choices[0].Message.Content, nil
}
