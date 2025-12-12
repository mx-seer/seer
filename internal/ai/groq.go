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
	Register("groq", NewGroq)
}

// Groq implements the Provider interface for Groq
type Groq struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

// Groq uses OpenAI-compatible API format
type groqRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type groqResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// NewGroq creates a new Groq provider
func NewGroq(cfg ProviderConfig) (Provider, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("groq requires api_key")
	}

	baseURL := cfg.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com/openai/v1"
	}

	model := cfg.Model
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}

	return &Groq{
		apiKey:  cfg.APIKey,
		baseURL: baseURL,
		model:   model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}, nil
}

func (g *Groq) Name() string {
	return "groq"
}

func (g *Groq) Available() bool {
	return g.apiKey != ""
}

func (g *Groq) Analyze(ctx context.Context, prompt string) (string, error) {
	reqBody := groqRequest{
		Model: g.model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.apiKey)

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("groq returned status %d: %s", resp.StatusCode, string(respBody))
	}

	var groqResp groqResponse
	if err := json.NewDecoder(resp.Body).Decode(&groqResp); err != nil {
		return "", err
	}

	if groqResp.Error != nil {
		return "", fmt.Errorf("groq error: %s", groqResp.Error.Message)
	}

	if len(groqResp.Choices) == 0 {
		return "", fmt.Errorf("groq returned no choices")
	}

	return groqResp.Choices[0].Message.Content, nil
}
