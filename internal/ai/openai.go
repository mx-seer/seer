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
	Register("openai", NewOpenAI)
}

// OpenAI implements the Provider interface for OpenAI
type OpenAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type openAIRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type openAIResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewOpenAI creates a new OpenAI provider
func NewOpenAI(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("openai requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &OpenAI{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (o *OpenAI) Name() string {
	return "openai"
}

func (o *OpenAI) Available() bool {
	return o.apiKey != ""
}

func (o *OpenAI) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := openAIRequest{
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

	resp, err := o.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var openAIResp openAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
		return "", err
	}

	if openAIResp.Error != nil {
		return "", fmt.Errorf("openai error: %s", openAIResp.Error.Message)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}

	return openAIResp.Choices[0].Message.Content, nil
}
