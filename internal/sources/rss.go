package sources

import (
	"context"
	"crypto/md5"
	"fmt"
	"net/http"
	"time"

	"github.com/mmcdole/gofeed"
)

// RSS fetches articles from RSS/Atom feeds
type RSS struct {
	config SourceConfig
	parser *gofeed.Parser
	client *http.Client
}

// NewRSS creates a new RSS source
func NewRSS(cfg SourceConfig) (Source, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("RSS source requires a URL")
	}

	return &RSS{
		config: cfg,
		parser: gofeed.NewParser(),
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Type returns the source type
func (r *RSS) Type() string {
	return "rss"
}

// Name returns the source name
func (r *RSS) Name() string {
	return r.config.Name
}

// Fetch retrieves items from the RSS feed
func (r *RSS) Fetch(ctx context.Context) ([]Opportunity, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, r.config.URL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Seer/1.0")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch feed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	feed, err := r.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed: %w", err)
	}

	var opportunities []Opportunity
	for _, item := range feed.Items {
		opp := r.itemToOpportunity(item, feed.Title)
		opportunities = append(opportunities, opp)
	}

	return opportunities, nil
}

func (r *RSS) itemToOpportunity(item *gofeed.Item, feedTitle string) Opportunity {
	// Generate unique ID from link or GUID
	externalID := item.GUID
	if externalID == "" {
		externalID = item.Link
	}
	if externalID == "" {
		// Hash the title as fallback
		hash := md5.Sum([]byte(item.Title + item.Published))
		externalID = fmt.Sprintf("%x", hash)
	}

	description := item.Description
	if description == "" && item.Content != "" {
		// Truncate content if too long
		if len(item.Content) > 500 {
			description = item.Content[:500] + "..."
		} else {
			description = item.Content
		}
	}

	var publishedAt time.Time
	if item.PublishedParsed != nil {
		publishedAt = *item.PublishedParsed
	} else if item.UpdatedParsed != nil {
		publishedAt = *item.UpdatedParsed
	} else {
		publishedAt = time.Now()
	}

	var categories []string
	for _, cat := range item.Categories {
		categories = append(categories, cat)
	}

	var author string
	if item.Author != nil {
		author = item.Author.Name
	}

	return Opportunity{
		Title:            item.Title,
		Description:      description,
		SourceType:       "rss",
		SourceURL:        item.Link,
		SourceIDExternal: externalID,
		DetectedAt:       publishedAt,
		Metadata: map[string]any{
			"feed_title": feedTitle,
			"feed_url":   r.config.URL,
			"author":     author,
			"categories": categories,
		},
	}
}
