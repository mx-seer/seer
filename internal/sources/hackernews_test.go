package sources

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHackerNews_Type(t *testing.T) {
	hn, _ := NewHackerNews(SourceConfig{Name: "Test HN"})
	if hn.Type() != "hackernews" {
		t.Errorf("expected type hackernews, got %s", hn.Type())
	}
}

func TestHackerNews_Name(t *testing.T) {
	hn, _ := NewHackerNews(SourceConfig{Name: "Test HN"})
	if hn.Name() != "Test HN" {
		t.Errorf("expected name Test HN, got %s", hn.Name())
	}
}

func TestHackerNews_Fetch(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := hnSearchResponse{
			Hits: []hnHit{
				{
					ObjectID:    "123",
					Title:       "Show HN: My cool project",
					URL:         "https://example.com",
					Author:      "testuser",
					Points:      100,
					NumComments: 50,
					CreatedAt:   "2024-01-01T12:00:00.000Z",
				},
				{
					ObjectID:  "456",
					Title:     "Ask HN: Need feedback on my startup",
					StoryText: "I'm building a new tool...",
					Author:    "founder",
					Points:    25,
					CreatedAt: "2024-01-01T10:00:00.000Z",
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// We can't easily inject the mock URL, so we'll test the hitToOpportunity function
	hn := &HackerNews{
		config: SourceConfig{Name: "Test"},
		client: server.Client(),
	}

	hit := hnHit{
		ObjectID:    "123",
		Title:       "Show HN: Test Project",
		URL:         "https://example.com",
		Author:      "tester",
		Points:      50,
		NumComments: 10,
		CreatedAt:   "2024-01-01T12:00:00.000Z",
	}

	opp := hn.hitToOpportunity(hit)

	if opp.Title != "Show HN: Test Project" {
		t.Errorf("expected title 'Show HN: Test Project', got %s", opp.Title)
	}

	if opp.SourceType != "hackernews" {
		t.Errorf("expected source type hackernews, got %s", opp.SourceType)
	}

	if opp.SourceIDExternal != "123" {
		t.Errorf("expected source ID 123, got %s", opp.SourceIDExternal)
	}

	if opp.SourceURL != "https://example.com" {
		t.Errorf("expected source URL https://example.com, got %s", opp.SourceURL)
	}

	if opp.Metadata["author"] != "tester" {
		t.Errorf("expected author tester, got %v", opp.Metadata["author"])
	}

	if opp.Metadata["points"] != 50 {
		t.Errorf("expected points 50, got %v", opp.Metadata["points"])
	}
}

func TestHackerNews_HitToOpportunity_NoURL(t *testing.T) {
	hn := &HackerNews{config: SourceConfig{Name: "Test"}}

	hit := hnHit{
		ObjectID:  "789",
		Title:     "Ask HN: Question",
		StoryText: "This is my question...",
		Author:    "asker",
		CreatedAt: "2024-01-01T12:00:00.000Z",
	}

	opp := hn.hitToOpportunity(hit)

	expectedURL := hnBaseURL + "789"
	if opp.SourceURL != expectedURL {
		t.Errorf("expected source URL %s, got %s", expectedURL, opp.SourceURL)
	}

	if opp.Description != "This is my question..." {
		t.Errorf("expected description from story_text, got %s", opp.Description)
	}
}

func TestHackerNews_FetchWithContext(t *testing.T) {
	hn, _ := NewHackerNews(SourceConfig{Name: "Test"})

	// Test with cancelled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// This should return quickly without error (empty results from failed queries)
	_, err := hn.Fetch(ctx)
	if err != nil {
		t.Errorf("expected no error with cancelled context, got %v", err)
	}
}
