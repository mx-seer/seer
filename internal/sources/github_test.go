package sources

import (
	"context"
	"testing"
	"time"
)

func TestGitHub_Type(t *testing.T) {
	gh, _ := NewGitHub(SourceConfig{Name: "Test GitHub"})
	if gh.Type() != "github" {
		t.Errorf("expected type github, got %s", gh.Type())
	}
}

func TestGitHub_Name(t *testing.T) {
	gh, _ := NewGitHub(SourceConfig{Name: "Test GitHub"})
	if gh.Name() != "Test GitHub" {
		t.Errorf("expected name Test GitHub, got %s", gh.Name())
	}
}

func TestGitHub_RepoToOpportunity(t *testing.T) {
	gh := &GitHub{config: SourceConfig{Name: "Test"}}

	now := time.Now()
	repo := ghRepo{
		ID:              12345,
		Name:            "awesome-project",
		FullName:        "user/awesome-project",
		Description:     "An awesome open source project",
		HTMLURL:         "https://github.com/user/awesome-project",
		StargazersCount: 1000,
		ForksCount:      100,
		OpenIssuesCount: 25,
		Language:        "Go",
		Topics:          []string{"golang", "cli"},
		CreatedAt:       now.Add(-30 * 24 * time.Hour),
		PushedAt:        now,
	}

	opp := gh.repoToOpportunity(repo)

	if opp.Title != "user/awesome-project" {
		t.Errorf("expected title 'user/awesome-project', got %s", opp.Title)
	}

	if opp.Description != "An awesome open source project" {
		t.Errorf("expected description 'An awesome open source project', got %s", opp.Description)
	}

	if opp.SourceType != "github" {
		t.Errorf("expected source type github, got %s", opp.SourceType)
	}

	if opp.SourceIDExternal != "12345" {
		t.Errorf("expected source ID 12345, got %s", opp.SourceIDExternal)
	}

	if opp.SourceURL != "https://github.com/user/awesome-project" {
		t.Errorf("expected URL https://github.com/user/awesome-project, got %s", opp.SourceURL)
	}

	if opp.Metadata["stars"] != 1000 {
		t.Errorf("expected stars 1000, got %v", opp.Metadata["stars"])
	}

	if opp.Metadata["language"] != "Go" {
		t.Errorf("expected language Go, got %v", opp.Metadata["language"])
	}
}

func TestGitHub_RepoToOpportunity_NoDescription(t *testing.T) {
	gh := &GitHub{config: SourceConfig{Name: "Test"}}

	repo := ghRepo{
		ID:       67890,
		Name:     "mystery-project",
		FullName: "dev/mystery-project",
		HTMLURL:  "https://github.com/dev/mystery-project",
		PushedAt: time.Now(),
	}

	opp := gh.repoToOpportunity(repo)

	expected := "Repository: dev/mystery-project"
	if opp.Description != expected {
		t.Errorf("expected description '%s', got %s", expected, opp.Description)
	}
}

func TestGitHub_FetchWithCancelledContext(t *testing.T) {
	gh, _ := NewGitHub(SourceConfig{Name: "Test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should return empty results without error
	_, err := gh.Fetch(ctx)
	if err != nil {
		t.Errorf("expected no error with cancelled context, got %v", err)
	}
}
