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
	npmRegistryAPI = "https://registry.npmjs.org/-/v1/search"
	npmPackageURL  = "https://www.npmjs.com/package/"
)

// NPM fetches new and trending packages from npm registry
type NPM struct {
	config SourceConfig
	client *http.Client
}

// npmSearchResponse represents npm search API response
type npmSearchResponse struct {
	Objects []npmObject `json:"objects"`
}

type npmObject struct {
	Package npmPackage `json:"package"`
	Score   npmScore   `json:"score"`
}

type npmPackage struct {
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Description string    `json:"description"`
	Keywords    []string  `json:"keywords"`
	Date        time.Time `json:"date"`
	Links       struct {
		NPM        string `json:"npm"`
		Homepage   string `json:"homepage"`
		Repository string `json:"repository"`
	} `json:"links"`
	Author struct {
		Name  string `json:"name"`
		Email string `json:"email"`
	} `json:"author"`
}

type npmScore struct {
	Final  float64 `json:"final"`
	Detail struct {
		Quality     float64 `json:"quality"`
		Popularity  float64 `json:"popularity"`
		Maintenance float64 `json:"maintenance"`
	} `json:"detail"`
}

// NewNPM creates a new NPM source
func NewNPM(cfg SourceConfig) (Source, error) {
	return &NPM{
		config: cfg,
		client: &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// Type returns the source type
func (n *NPM) Type() string {
	return "npm"
}

// Name returns the source name
func (n *NPM) Name() string {
	return n.config.Name
}

// Fetch retrieves new packages from npm
func (n *NPM) Fetch(ctx context.Context) ([]Opportunity, error) {
	// Optimized queries for opportunity detection
	queries := []string{
		// Developer tools
		"cli",
		"devtool",
		"developer tool",
		"dev tool",

		// Self-hosted / alternatives
		"self-hosted",
		"selfhosted",
		"alternative",
		"open source",

		// Starters and templates
		"boilerplate",
		"starter",
		"template",
		"scaffold",
		"generator",

		// API/SDK (competition)
		"sdk",
		"api client",
		"wrapper",

		// Specific ecosystems
		"svelte",
		"nuxt",
		"vite plugin",
		"elysia",
		"hono",
		"bun",

		// Utilities
		"logger",
		"validation",
		"auth",
		"database",

		// Trending categories
		"ai",
		"llm",
		"openai",
		"markdown",
		"pdf",
	}

	seen := make(map[string]bool)
	var opportunities []Opportunity
	cutoff := time.Now().AddDate(0, 0, -14) // Only packages from last 14 days

	for _, query := range queries {
		packages, err := n.search(ctx, query)
		if err != nil {
			continue
		}

		for _, pkg := range packages {
			if seen[pkg.Package.Name] {
				continue
			}
			// Filter by date: only keep packages updated in last 30 days
			if pkg.Package.Date.Before(cutoff) {
				continue
			}
			seen[pkg.Package.Name] = true

			opp := n.packageToOpportunity(pkg)
			opportunities = append(opportunities, opp)
		}
	}

	return opportunities, nil
}

func (n *NPM) search(ctx context.Context, query string) ([]npmObject, error) {
	params := url.Values{}
	params.Set("text", query)
	params.Set("size", "25")
	// Boost maintenance to favor actively maintained packages
	params.Set("quality", "0.3")
	params.Set("popularity", "0.3")
	params.Set("maintenance", "0.4")

	reqURL := npmRegistryAPI + "?" + params.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := n.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	var result npmSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return result.Objects, nil
}

func (n *NPM) packageToOpportunity(obj npmObject) Opportunity {
	pkg := obj.Package

	sourceURL := pkg.Links.NPM
	if sourceURL == "" {
		sourceURL = npmPackageURL + pkg.Name
	}

	description := pkg.Description
	if description == "" {
		description = fmt.Sprintf("npm package: %s v%s", pkg.Name, pkg.Version)
	}

	return Opportunity{
		Title:            pkg.Name,
		Description:      description,
		SourceType:       "npm",
		SourceURL:        sourceURL,
		SourceIDExternal: pkg.Name + "@" + pkg.Version,
		DetectedAt:       pkg.Date,
		Metadata: map[string]any{
			"version":     pkg.Version,
			"keywords":    pkg.Keywords,
			"author":      pkg.Author.Name,
			"score":       obj.Score.Final,
			"quality":     obj.Score.Detail.Quality,
			"popularity":  obj.Score.Detail.Popularity,
			"maintenance": obj.Score.Detail.Maintenance,
			"homepage":    pkg.Links.Homepage,
			"repository":  pkg.Links.Repository,
		},
	}
}
