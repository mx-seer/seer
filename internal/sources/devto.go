package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const devtoAPI = "https://dev.to/api/articles"

// DevTo fetches articles from DEV.to
type DevTo struct {
	config SourceConfig
	client *http.Client
}

// devtoArticle represents a DEV.to article
type devtoArticle struct {
	ID                 int64     `json:"id"`
	Title              string    `json:"title"`
	Description        string    `json:"description"`
	URL                string    `json:"url"`
	PublishedAt        time.Time `json:"published_at"`
	PublishedTimestamp string    `json:"published_timestamp"`
	TagList            []string  `json:"tag_list"`
	User               struct {
		Name     string `json:"name"`
		Username string `json:"username"`
	} `json:"user"`
	PositiveReactionsCount int `json:"positive_reactions_count"`
	CommentsCount          int `json:"comments_count"`
	ReadingTimeMinutes     int `json:"reading_time_minutes"`
}

// NewDevTo creates a new DEV.to source
func NewDevTo(cfg SourceConfig) (Source, error) {
	return &DevTo{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Type returns the source type
func (d *DevTo) Type() string {
	return "devto"
}

// Name returns the source name
func (d *DevTo) Name() string {
	return d.config.Name
}

// Fetch retrieves recent articles from DEV.to
func (d *DevTo) Fetch(ctx context.Context) ([]Opportunity, error) {
	// Optimized tags for opportunity detection
	tags := []string{
		// Project showcases
		"showdev",
		"opensource",
		"sideproject",

		// Startup/business
		"startup",
		"entrepreneurship",
		"indiehackers",
		"buildinpublic",

		// Developer experience
		"productivity",
		"devtools",
		"tooling",

		// Technical categories
		"selfhosted",
		"docker",
		"api",
		"cli",

		// Discussion/discovery
		"discuss",
		"watercooler",
		"news",

		// Learning (pain points often surface here)
		"tutorial",
		"beginners",
		"webdev",
		"programming",
	}

	seen := make(map[int64]bool)
	var opportunities []Opportunity
	cutoff := time.Now().AddDate(0, 0, -7) // Only articles from last 7 days

	// Fetch by tags
	for _, tag := range tags {
		articles, err := d.fetchArticles(ctx, tag)
		if err != nil {
			continue
		}

		for _, article := range articles {
			if seen[article.ID] {
				continue
			}
			// Filter by date: only keep articles from last 7 days
			if article.PublishedAt.Before(cutoff) {
				continue
			}
			seen[article.ID] = true

			opp := d.articleToOpportunity(article)
			opportunities = append(opportunities, opp)
		}
	}

	// Also search by opportunity keywords in titles
	keywords := []string{
		"built",
		"launched",
		"alternative",
		"self-hosted",
		"open source",
		"side project",
		"weekend project",
	}

	for _, keyword := range keywords {
		articles, err := d.fetchLatestArticles(ctx)
		if err != nil {
			continue
		}

		keywordLower := strings.ToLower(keyword)
		for _, article := range articles {
			if seen[article.ID] {
				continue
			}
			if article.PublishedAt.Before(cutoff) {
				continue
			}
			// Filter by keyword in title or description
			titleLower := strings.ToLower(article.Title)
			descLower := strings.ToLower(article.Description)
			if !strings.Contains(titleLower, keywordLower) && !strings.Contains(descLower, keywordLower) {
				continue
			}
			seen[article.ID] = true

			opp := d.articleToOpportunity(article)
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

func (d *DevTo) fetchArticles(ctx context.Context, tag string) ([]devtoArticle, error) {
	params := url.Values{}
	params.Set("tag", tag)
	params.Set("per_page", "20")
	params.Set("state", "rising")

	reqURL := devtoAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Seer/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var articles []devtoArticle
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return articles, nil
}

func (d *DevTo) fetchLatestArticles(ctx context.Context) ([]devtoArticle, error) {
	params := url.Values{}
	params.Set("per_page", "50")
	params.Set("state", "fresh")

	reqURL := devtoAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Seer/1.0")

	resp, err := d.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var articles []devtoArticle
	if err := json.NewDecoder(resp.Body).Decode(&articles); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return articles, nil
}

func (d *DevTo) articleToOpportunity(article devtoArticle) Opportunity {
	description := article.Description
	if description == "" {
		description = fmt.Sprintf("Article by %s", article.User.Name)
	}

	return Opportunity{
		Title:            article.Title,
		Description:      description,
		SourceType:       "devto",
		SourceURL:        article.URL,
		SourceIDExternal: fmt.Sprintf("%d", article.ID),
		DetectedAt:       article.PublishedAt,
		Metadata: map[string]any{
			"author":        article.User.Name,
			"username":      article.User.Username,
			"tags":          article.TagList,
			"reactions":     article.PositiveReactionsCount,
			"comments":      article.CommentsCount,
			"reading_time":  article.ReadingTimeMinutes,
		},
	}
}
