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
	ghAPIBase   = "https://api.github.com"
	ghSearchAPI = ghAPIBase + "/search/repositories"
)

// GitHub fetches trending repositories and issues from GitHub
type GitHub struct {
	config SourceConfig
	client *http.Client
}

// ghSearchResponse represents GitHub search API response
type ghSearchResponse struct {
	Items []ghRepo `json:"items"`
}

type ghRepo struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Description     string    `json:"description"`
	HTMLURL         string    `json:"html_url"`
	StargazersCount int       `json:"stargazers_count"`
	ForksCount      int       `json:"forks_count"`
	OpenIssuesCount int       `json:"open_issues_count"`
	Language        string    `json:"language"`
	Topics          []string  `json:"topics"`
	CreatedAt       time.Time `json:"created_at"`
	PushedAt        time.Time `json:"pushed_at"`
}

// NewGitHub creates a new GitHub source
func NewGitHub(cfg SourceConfig) (Source, error) {
	return &GitHub{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Type returns the source type
func (g *GitHub) Type() string {
	return "github"
}

// Name returns the source name
func (g *GitHub) Name() string {
	return g.config.Name
}

// Fetch retrieves trending repositories from GitHub
func (g *GitHub) Fetch(ctx context.Context) ([]Opportunity, error) {
	// Search for recently created repos with good engagement
	queries := []string{
		"stars:>10 created:>" + time.Now().Add(-7*24*time.Hour).Format("2006-01-02"),
		"help wanted good first issue",
		"looking for contributors",
	}

	seen := make(map[int64]bool)
	var opportunities []Opportunity

	for _, query := range queries {
		repos, err := g.searchRepos(ctx, query)
		if err != nil {
			continue
		}

		for _, repo := range repos {
			if seen[repo.ID] {
				continue
			}
			seen[repo.ID] = true

			opp := g.repoToOpportunity(repo)
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

func (g *GitHub) searchRepos(ctx context.Context, query string) ([]ghRepo, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("sort", "stars")
	params.Set("order", "desc")
	params.Set("per_page", "20")

	reqURL := ghSearchAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "Seer/1.0")

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result ghSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Items, nil
}

func (g *GitHub) repoToOpportunity(repo ghRepo) Opportunity {
	description := repo.Description
	if description == "" {
		description = fmt.Sprintf("Repository: %s", repo.FullName)
	}

	return Opportunity{
		Title:            repo.FullName,
		Description:      description,
		SourceType:       "github",
		SourceURL:        repo.HTMLURL,
		SourceIDExternal: fmt.Sprintf("%d", repo.ID),
		DetectedAt:       repo.PushedAt,
		Metadata: map[string]any{
			"stars":        repo.StargazersCount,
			"forks":        repo.ForksCount,
			"open_issues":  repo.OpenIssuesCount,
			"language":     repo.Language,
			"topics":       repo.Topics,
			"created_at":   repo.CreatedAt,
		},
	}
}
