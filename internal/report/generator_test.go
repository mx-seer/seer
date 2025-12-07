package report

import (
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	gen := New()
	if gen == nil {
		t.Fatal("expected generator to be created")
	}
}

func TestGenerate_EmptyOpportunities(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	report := gen.Generate([]Opportunity{}, start, end)

	if report.OpportunityCount != 0 {
		t.Errorf("expected 0 opportunities, got %d", report.OpportunityCount)
	}

	if !strings.Contains(report.ContentHuman, "No opportunities found") {
		t.Error("expected human content to mention no opportunities")
	}
}

func TestGenerate_WithOpportunities(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	opportunities := []Opportunity{
		{
			ID:          1,
			Title:       "High Score Opportunity",
			Description: "This is a great opportunity",
			SourceType:  "hackernews",
			SourceURL:   "https://example.com/1",
			Score:       90,
			Signals:     []string{"problem_mention", "solution_seeking"},
			DetectedAt:  time.Now(),
		},
		{
			ID:          2,
			Title:       "Medium Score Opportunity",
			Description: "This is an okay opportunity",
			SourceType:  "github",
			SourceURL:   "https://example.com/2",
			Score:       50,
			Signals:     []string{"technical"},
			DetectedAt:  time.Now(),
		},
		{
			ID:          3,
			Title:       "Low Score Opportunity",
			Description: "This is a low priority opportunity",
			SourceType:  "devto",
			SourceURL:   "https://example.com/3",
			Score:       20,
			Signals:     []string{},
			DetectedAt:  time.Now(),
		},
	}

	report := gen.Generate(opportunities, start, end)

	if report.OpportunityCount != 3 {
		t.Errorf("expected 3 opportunities, got %d", report.OpportunityCount)
	}

	// Should be sorted by score descending
	if report.Opportunities[0].Score != 90 {
		t.Error("expected opportunities to be sorted by score descending")
	}

	// Human readable should contain title
	if !strings.Contains(report.ContentHuman, "High Score Opportunity") {
		t.Error("expected human content to contain opportunity title")
	}

	// Prompt should contain analysis instructions
	if !strings.Contains(report.ContentPrompt, "market analyst") {
		t.Error("expected prompt to contain analyst role")
	}
}

func TestGenerate_SortsOpportunities(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	opportunities := []Opportunity{
		{ID: 1, Title: "Low", Score: 10},
		{ID: 2, Title: "High", Score: 90},
		{ID: 3, Title: "Medium", Score: 50},
	}

	report := gen.Generate(opportunities, start, end)

	if report.Opportunities[0].Title != "High" {
		t.Errorf("expected first opportunity to be 'High', got '%s'", report.Opportunities[0].Title)
	}

	if report.Opportunities[1].Title != "Medium" {
		t.Errorf("expected second opportunity to be 'Medium', got '%s'", report.Opportunities[1].Title)
	}

	if report.Opportunities[2].Title != "Low" {
		t.Errorf("expected third opportunity to be 'Low', got '%s'", report.Opportunities[2].Title)
	}
}

func TestGenerate_HumanReadableContent(t *testing.T) {
	gen := New()
	start := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	opportunities := []Opportunity{
		{
			ID:          1,
			Title:       "Test Opportunity",
			Description: "A test description",
			SourceType:  "hackernews",
			SourceURL:   "https://example.com",
			Score:       75,
			Signals:     []string{"signal1", "signal2"},
		},
	}

	report := gen.Generate(opportunities, start, end)

	// Check structure
	if !strings.Contains(report.ContentHuman, "# Seer Opportunity Report") {
		t.Error("expected human content to have title")
	}

	if !strings.Contains(report.ContentHuman, "**Period:**") {
		t.Error("expected human content to have period")
	}

	if !strings.Contains(report.ContentHuman, "**Score:** 75/100") {
		t.Error("expected human content to have score")
	}

	if !strings.Contains(report.ContentHuman, "signal1, signal2") {
		t.Error("expected human content to have signals")
	}

	if !strings.Contains(report.ContentHuman, "## Summary by Source") {
		t.Error("expected human content to have source summary")
	}
}

func TestGenerate_PromptContent(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	opportunities := []Opportunity{
		{
			ID:          1,
			Title:       "Test Opportunity",
			Description: "A test description",
			SourceType:  "hackernews",
			SourceURL:   "https://example.com",
			Score:       75,
			Signals:     []string{"signal1"},
		},
	}

	report := gen.Generate(opportunities, start, end)

	// Check prompt structure
	if !strings.Contains(report.ContentPrompt, "expert market analyst") {
		t.Error("expected prompt to set analyst role")
	}

	if !strings.Contains(report.ContentPrompt, "indie developers") {
		t.Error("expected prompt to mention indie developers")
	}

	if !strings.Contains(report.ContentPrompt, "=== OPPORTUNITIES ===") {
		t.Error("expected prompt to have opportunities section")
	}

	if !strings.Contains(report.ContentPrompt, "actionable") {
		t.Error("expected prompt to request actionable output")
	}
}

func TestGenerate_TruncatesLongDescriptions(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	longDesc := strings.Repeat("x", 1000)
	opportunities := []Opportunity{
		{
			ID:          1,
			Title:       "Test",
			Description: longDesc,
			SourceURL:   "https://example.com",
			Score:       50,
		},
	}

	report := gen.Generate(opportunities, start, end)

	// Human readable truncates at 300
	if strings.Contains(report.ContentHuman, longDesc) {
		t.Error("expected human content to truncate long descriptions")
	}

	if !strings.Contains(report.ContentHuman, "...") {
		t.Error("expected human content to have ellipsis for truncated descriptions")
	}
}

func TestGetTopOpportunities(t *testing.T) {
	report := &Report{
		Opportunities: []Opportunity{
			{ID: 1, Score: 90},
			{ID: 2, Score: 80},
			{ID: 3, Score: 70},
			{ID: 4, Score: 60},
			{ID: 5, Score: 50},
		},
	}

	top3 := report.GetTopOpportunities(3)
	if len(top3) != 3 {
		t.Errorf("expected 3 opportunities, got %d", len(top3))
	}

	if top3[0].ID != 1 || top3[1].ID != 2 || top3[2].ID != 3 {
		t.Error("expected top 3 opportunities by order")
	}
}

func TestGetTopOpportunities_MoreThanAvailable(t *testing.T) {
	report := &Report{
		Opportunities: []Opportunity{
			{ID: 1, Score: 90},
			{ID: 2, Score: 80},
		},
	}

	top10 := report.GetTopOpportunities(10)
	if len(top10) != 2 {
		t.Errorf("expected 2 opportunities (all available), got %d", len(top10))
	}
}

func TestGenerate_SourceSummary(t *testing.T) {
	gen := New()
	start := time.Now().Add(-24 * time.Hour)
	end := time.Now()

	opportunities := []Opportunity{
		{ID: 1, SourceType: "hackernews", Score: 80},
		{ID: 2, SourceType: "hackernews", Score: 60},
		{ID: 3, SourceType: "github", Score: 70},
	}

	report := gen.Generate(opportunities, start, end)

	if !strings.Contains(report.ContentHuman, "hackernews") {
		t.Error("expected summary to contain hackernews")
	}

	if !strings.Contains(report.ContentHuman, "github") {
		t.Error("expected summary to contain github")
	}

	// hackernews should show 2 opportunities
	if !strings.Contains(report.ContentHuman, "2 opportunities") {
		t.Error("expected summary to show correct count for hackernews")
	}
}
