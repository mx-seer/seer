//go:build pro

package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Reddit fetches opportunities from Reddit subreddits
type Reddit struct {
	name       string
	subreddits []string
	keywords   []string
	client     *http.Client
}

// redditListing represents Reddit API response
type redditListing struct {
	Data struct {
		Children []struct {
			Data redditPost `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

type redditPost struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Selftext  string  `json:"selftext"`
	URL       string  `json:"url"`
	Permalink string  `json:"permalink"`
	Subreddit string  `json:"subreddit"`
	Score     int     `json:"score"`
	Created   float64 `json:"created_utc"`
	NumComments int   `json:"num_comments"`
}

// NewReddit creates a new Reddit source
func NewReddit(cfg SourceConfig) (Source, error) {
	subreddits := []string{"SideProject", "startups", "Entrepreneur", "SaaS", "indiehackers"}
	if subs, ok := cfg.Config["subreddits"]; ok && subs != "" {
		// Parse comma-separated subreddits
		subreddits = parseCSV(subs)
	}

	var keywords []string
	if kw, ok := cfg.Config["keywords"]; ok && kw != "" {
		keywords = parseCSV(kw)
	}

	return &Reddit{
		name:       cfg.Name,
		subreddits: subreddits,
		keywords:   keywords,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (r *Reddit) Type() string {
	return "reddit"
}

func (r *Reddit) Name() string {
	return r.name
}

func (r *Reddit) Fetch(ctx context.Context) ([]Opportunity, error) {
	var opportunities []Opportunity

	for _, subreddit := range r.subreddits {
		posts, err := r.fetchSubreddit(ctx, subreddit)
		if err != nil {
			continue // Skip failed subreddits
		}
		opportunities = append(opportunities, posts...)
	}

	return opportunities, nil
}

func (r *Reddit) fetchSubreddit(ctx context.Context, subreddit string) ([]Opportunity, error) {
	url := fmt.Sprintf("https://www.reddit.com/r/%s/new.json?limit=50", subreddit)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	// Reddit requires a User-Agent
	req.Header.Set("User-Agent", "Seer/1.0 (market opportunity detector)")

	resp, err := r.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("reddit API returned status %d", resp.StatusCode)
	}

	var listing redditListing
	if err := json.NewDecoder(resp.Body).Decode(&listing); err != nil {
		return nil, err
	}

	var opportunities []Opportunity
	for _, child := range listing.Data.Children {
		post := child.Data

		// Filter by keywords if configured
		if len(r.keywords) > 0 && !containsAnyKeyword(post.Title+" "+post.Selftext, r.keywords) {
			continue
		}

		opportunities = append(opportunities, Opportunity{
			Title:            post.Title,
			Description:      truncate(post.Selftext, 500),
			SourceType:       "reddit",
			SourceURL:        "https://reddit.com" + post.Permalink,
			SourceIDExternal: post.ID,
			DetectedAt:       time.Unix(int64(post.Created), 0),
			Metadata: map[string]any{
				"subreddit":    post.Subreddit,
				"score":        post.Score,
				"num_comments": post.NumComments,
			},
		})
	}

	return opportunities, nil
}

// parseCSV splits a comma-separated string into a slice
func parseCSV(s string) []string {
	var result []string
	for _, part := range splitAndTrim(s, ",") {
		if part != "" {
			result = append(result, part)
		}
	}
	return result
}

func splitAndTrim(s, sep string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if i+len(sep) <= len(s) && s[i:i+len(sep)] == sep {
			part := trimSpace(s[start:i])
			result = append(result, part)
			start = i + len(sep)
			i += len(sep) - 1
		}
	}
	result = append(result, trimSpace(s[start:]))
	return result
}

func trimSpace(s string) string {
	start := 0
	end := len(s)
	for start < end && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n') {
		start++
	}
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n') {
		end--
	}
	return s[start:end]
}
