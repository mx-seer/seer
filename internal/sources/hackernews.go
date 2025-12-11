package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const (
	hnAlgoliaAPI = "https://hn.algolia.com/api/v1/search_by_date"
	hnBaseURL    = "https://news.ycombinator.com/item?id="
)

// HackerNews fetches opportunities from Hacker News via Algolia API
type HackerNews struct {
	config SourceConfig
	client *http.Client
}

// hnSearchResponse represents the Algolia API response
type hnSearchResponse struct {
	Hits []hnHit `json:"hits"`
}

type hnHit struct {
	ObjectID    string   `json:"objectID"`
	Title       string   `json:"title"`
	URL         string   `json:"url"`
	StoryText   string   `json:"story_text"`
	Author      string   `json:"author"`
	Points      int      `json:"points"`
	NumComments int      `json:"num_comments"`
	CreatedAt   string   `json:"created_at"`
	Tags        []string `json:"_tags"`
}

// NewHackerNews creates a new HackerNews source
func NewHackerNews(cfg SourceConfig) (Source, error) {
	return &HackerNews{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Type returns the source type
func (h *HackerNews) Type() string {
	return "hackernews"
}

// Name returns the source name
func (h *HackerNews) Name() string {
	return h.config.Name
}

// Fetch retrieves recent opportunities from Hacker News
func (h *HackerNews) Fetch(ctx context.Context) ([]Opportunity, error) {
	// Optimized queries to detect market opportunities
	queries := []string{
		// Direct opportunities - people stating needs
		"I wish",
		"I need",
		"looking for",
		"searching for",

		// Pain points - frustration signals
		"frustrated with",
		"annoyed by",
		"hate using",
		"problem with",
		"issue with",
		"struggle with",

		// Requests for alternatives
		"alternative to",
		"replacement for",
		"instead of",
		"better than",
		"competitor to",

		// Validation signals - willingness to pay
		"would pay for",
		"shut up and take my money",
		"take my money",
		"happy to pay",

		// Discovery requests
		"what do you use for",
		"how do you handle",
		"recommend a",
		"suggest a",
		"does anyone know",
		"is there a",
		"why isn't there",

		// Build signals
		"someone should build",
		"why hasn't anyone",
		"idea for a startup",
		"business idea",

		// Show HN - new launches to analyze
		"Show HN",

		// Ask HN - direct questions with opportunities
		"Ask HN",
	}

	seen := make(map[string]bool)
	var opportunities []Opportunity

	for _, query := range queries {
		hits, err := h.search(ctx, query)
		if err != nil {
			continue // Skip failed queries
		}

		for _, hit := range hits {
			if seen[hit.ObjectID] {
				continue
			}
			seen[hit.ObjectID] = true

			opp := h.hitToOpportunity(hit)
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

func (h *HackerNews) search(ctx context.Context, query string) ([]hnHit, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Set("tags", "story")
	params.Set("hitsPerPage", "20")
	params.Set("numericFilters", "created_at_i>"+fmt.Sprintf("%d", time.Now().Add(-24*time.Hour).Unix()))

	reqURL := hnAlgoliaAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := h.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result hnSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Hits, nil
}

func (h *HackerNews) hitToOpportunity(hit hnHit) Opportunity {
	// Always use HN discussion URL so users go to where the topic is being discussed
	sourceURL := hnBaseURL + hit.ObjectID

	description := hit.StoryText
	if description == "" && hit.URL != "" {
		description = hit.URL
	}

	createdAt, _ := time.Parse(time.RFC3339, hit.CreatedAt)
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return Opportunity{
		Title:            hit.Title,
		Description:      description,
		SourceType:       "hackernews",
		SourceURL:        sourceURL,
		SourceIDExternal: hit.ObjectID,
		DetectedAt:       createdAt,
		Metadata: map[string]any{
			"author":       hit.Author,
			"points":       hit.Points,
			"num_comments": hit.NumComments,
			"hn_url":       hnBaseURL + hit.ObjectID,
		},
	}
}
