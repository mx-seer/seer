package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Custom fetches opportunities from a generic JSON API endpoint
type Custom struct {
	name     string
	url      string
	headers  map[string]string
	mapping  customMapping
	client   *http.Client
}

// customMapping defines how to map JSON response to Opportunity
type customMapping struct {
	// JSONPath-like selectors (simplified)
	ItemsPath       string // Path to array of items (e.g., "data.items")
	TitleField      string // Field for title (e.g., "name" or "title")
	DescriptionField string // Field for description
	URLField        string // Field for URL
	IDField         string // Field for external ID
	DateField       string // Field for date (optional)
}

// NewCustom creates a new Custom API source
func NewCustom(cfg SourceConfig) (Source, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("custom source requires a URL")
	}

	// Parse headers from config
	headers := make(map[string]string)
	if auth, ok := cfg.Config["authorization"]; ok {
		headers["Authorization"] = auth
	}
	if apiKey, ok := cfg.Config["api_key"]; ok {
		headers["X-API-Key"] = apiKey
	}

	// Parse field mappings
	mapping := customMapping{
		ItemsPath:        getConfigOrDefault(cfg.Config, "items_path", ""),
		TitleField:       getConfigOrDefault(cfg.Config, "title_field", "title"),
		DescriptionField: getConfigOrDefault(cfg.Config, "description_field", "description"),
		URLField:         getConfigOrDefault(cfg.Config, "url_field", "url"),
		IDField:          getConfigOrDefault(cfg.Config, "id_field", "id"),
		DateField:        getConfigOrDefault(cfg.Config, "date_field", ""),
	}

	return &Custom{
		name:    cfg.Name,
		url:     cfg.URL,
		headers: headers,
		mapping: mapping,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (c *Custom) Type() string {
	return "custom"
}

func (c *Custom) Name() string {
	return c.name
}

func (c *Custom) Fetch(ctx context.Context) ([]Opportunity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json")
	for key, value := range c.headers {
		req.Header.Set(key, value)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("custom API returned status %d", resp.StatusCode)
	}

	var rawResponse any
	if err := json.NewDecoder(resp.Body).Decode(&rawResponse); err != nil {
		return nil, err
	}

	items := c.extractItems(rawResponse)
	if items == nil {
		return nil, fmt.Errorf("could not extract items from response")
	}

	var opportunities []Opportunity
	for _, item := range items {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}

		opp := Opportunity{
			Title:            c.extractString(itemMap, c.mapping.TitleField),
			Description:      c.extractString(itemMap, c.mapping.DescriptionField),
			SourceType:       "custom",
			SourceURL:        c.extractString(itemMap, c.mapping.URLField),
			SourceIDExternal: c.extractString(itemMap, c.mapping.IDField),
			DetectedAt:       time.Now(),
		}

		if c.mapping.DateField != "" {
			if dateStr := c.extractString(itemMap, c.mapping.DateField); dateStr != "" {
				if parsed, err := time.Parse(time.RFC3339, dateStr); err == nil {
					opp.DetectedAt = parsed
				}
			}
		}

		if opp.Title != "" && opp.SourceIDExternal != "" {
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

func (c *Custom) extractItems(data any) []any {
	if c.mapping.ItemsPath == "" {
		// Root is the array
		if arr, ok := data.([]any); ok {
			return arr
		}
		return nil
	}

	// Navigate path like "data.items"
	current := data
	for _, part := range splitAndTrim(c.mapping.ItemsPath, ".") {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return nil
		}
	}

	if arr, ok := current.([]any); ok {
		return arr
	}
	return nil
}

func (c *Custom) extractString(item map[string]any, field string) string {
	if field == "" {
		return ""
	}

	// Handle nested fields like "metadata.title"
	parts := splitAndTrim(field, ".")
	var current any = item

	for _, part := range parts {
		if m, ok := current.(map[string]any); ok {
			current = m[part]
		} else {
			return ""
		}
	}

	switch v := current.(type) {
	case string:
		return v
	case float64:
		return fmt.Sprintf("%.0f", v)
	case int:
		return fmt.Sprintf("%d", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func getConfigOrDefault(config map[string]string, key, defaultVal string) string {
	if val, ok := config[key]; ok && val != "" {
		return val
	}
	return defaultVal
}
