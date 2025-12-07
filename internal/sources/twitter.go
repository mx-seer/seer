//go:build pro

package sources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Twitter fetches opportunities from Twitter/X API
type Twitter struct {
	name       string
	keywords   []string
	bearerToken string
	client     *http.Client
}

// twitterSearchResponse represents Twitter API v2 search response
type twitterSearchResponse struct {
	Data []twitterTweet `json:"data"`
	Meta struct {
		ResultCount int    `json:"result_count"`
		NextToken   string `json:"next_token"`
	} `json:"meta"`
}

type twitterTweet struct {
	ID        string `json:"id"`
	Text      string `json:"text"`
	AuthorID  string `json:"author_id"`
	CreatedAt string `json:"created_at"`
	PublicMetrics struct {
		RetweetCount int `json:"retweet_count"`
		LikeCount    int `json:"like_count"`
		ReplyCount   int `json:"reply_count"`
	} `json:"public_metrics"`
}

// NewTwitter creates a new Twitter source
func NewTwitter(cfg SourceConfig) (Source, error) {
	bearerToken := ""
	if token, ok := cfg.Config["bearer_token"]; ok {
		bearerToken = token
	}

	if bearerToken == "" {
		return nil, fmt.Errorf("twitter source requires bearer_token in config")
	}

	keywords := []string{"looking for", "need a tool", "wish there was", "anyone know", "alternative to"}
	if kw, ok := cfg.Config["keywords"]; ok && kw != "" {
		keywords = parseCSV(kw)
	}

	return &Twitter{
		name:        cfg.Name,
		keywords:    keywords,
		bearerToken: bearerToken,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

func (t *Twitter) Type() string {
	return "twitter"
}

func (t *Twitter) Name() string {
	return t.name
}

func (t *Twitter) Fetch(ctx context.Context) ([]Opportunity, error) {
	var opportunities []Opportunity

	for _, keyword := range t.keywords {
		tweets, err := t.searchTweets(ctx, keyword)
		if err != nil {
			continue // Skip failed searches
		}
		opportunities = append(opportunities, tweets...)
	}

	return opportunities, nil
}

func (t *Twitter) searchTweets(ctx context.Context, query string) ([]Opportunity, error) {
	endpoint := "https://api.twitter.com/2/tweets/search/recent"

	params := url.Values{}
	params.Set("query", query+" -is:retweet lang:en")
	params.Set("max_results", "50")
	params.Set("tweet.fields", "created_at,public_metrics,author_id")

	reqURL := fmt.Sprintf("%s?%s", endpoint, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+t.bearerToken)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("twitter API returned status %d", resp.StatusCode)
	}

	var searchResp twitterSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return nil, err
	}

	var opportunities []Opportunity
	for _, tweet := range searchResp.Data {
		createdAt, _ := time.Parse(time.RFC3339, tweet.CreatedAt)

		opportunities = append(opportunities, Opportunity{
			Title:            truncate(tweet.Text, 100),
			Description:      tweet.Text,
			SourceType:       "twitter",
			SourceURL:        fmt.Sprintf("https://twitter.com/i/web/status/%s", tweet.ID),
			SourceIDExternal: tweet.ID,
			DetectedAt:       createdAt,
			Metadata: map[string]any{
				"author_id":     tweet.AuthorID,
				"retweet_count": tweet.PublicMetrics.RetweetCount,
				"like_count":    tweet.PublicMetrics.LikeCount,
				"reply_count":   tweet.PublicMetrics.ReplyCount,
				"keyword":       query,
			},
		})
	}

	return opportunities, nil
}
