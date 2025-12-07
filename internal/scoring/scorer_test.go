package scoring

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	scorer := New()
	if scorer == nil {
		t.Fatal("expected scorer to be created")
	}
	if len(scorer.signals) == 0 {
		t.Error("expected default signals to be registered")
	}
}

func TestScore_HighScore(t *testing.T) {
	scorer := New()

	opp := Opportunity{
		Title:       "Show HN: I built a tool to solve the problem of API documentation",
		Description: "Looking for feedback on my indie project. I was frustrated with existing solutions...",
		SourceType:  "hackernews",
		DetectedAt:  time.Now(),
		Metadata: map[string]any{
			"points":       100,
			"num_comments": 50,
		},
	}

	result := scorer.Score(opp)

	if result.Score < 50 {
		t.Errorf("expected high score (>50), got %d", result.Score)
	}

	matched := result.GetMatchedSignals()
	if len(matched) < 3 {
		t.Errorf("expected at least 3 matched signals, got %d", len(matched))
	}
}

func TestScore_LowScore(t *testing.T) {
	scorer := New()

	opp := Opportunity{
		Title:       "Random news article",
		Description: "Some generic content here",
		SourceType:  "rss",
		DetectedAt:  time.Now().Add(-48 * time.Hour),
		Metadata:    map[string]any{},
	}

	result := scorer.Score(opp)

	if result.Score > 30 {
		t.Errorf("expected low score (<30), got %d", result.Score)
	}
}

func TestScore_ProblemMention(t *testing.T) {
	scorer := New()

	opp := Opportunity{
		Title:       "I have a problem with my database",
		Description: "It's frustrating that this doesn't work",
		SourceType:  "hackernews",
		DetectedAt:  time.Now(),
		Metadata:    map[string]any{},
	}

	result := scorer.Score(opp)

	found := false
	for _, s := range result.Signals {
		if s.Name == "problem_mention" && s.Matched {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected problem_mention signal to match")
	}
}

func TestScore_SolutionSeeking(t *testing.T) {
	scorer := New()

	opp := Opportunity{
		Title:       "How do I deploy my app?",
		Description: "Looking for the best way to handle this",
		SourceType:  "hackernews",
		DetectedAt:  time.Now(),
		Metadata:    map[string]any{},
	}

	result := scorer.Score(opp)

	found := false
	for _, s := range result.Signals {
		if s.Name == "solution_seeking" && s.Matched {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected solution_seeking signal to match")
	}
}

func TestScore_ShowProject(t *testing.T) {
	scorer := New()

	opp := Opportunity{
		Title:       "Show HN: My new side project",
		Description: "I built this over the weekend",
		SourceType:  "hackernews",
		DetectedAt:  time.Now(),
		Metadata:    map[string]any{},
	}

	result := scorer.Score(opp)

	found := false
	for _, s := range result.Signals {
		if s.Name == "show_project" && s.Matched {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected show_project signal to match")
	}
}

func TestScore_HighEngagement(t *testing.T) {
	scorer := New()

	testCases := []struct {
		name     string
		metadata map[string]any
	}{
		{"high_points", map[string]any{"points": 100}},
		{"high_comments", map[string]any{"num_comments": 50}},
		{"high_stars", map[string]any{"stars": 200}},
		{"high_reactions", map[string]any{"reactions": 30}},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			opp := Opportunity{
				Title:       "Test",
				Description: "Test",
				DetectedAt:  time.Now().Add(-48 * time.Hour), // Not recent to isolate engagement
				Metadata:    tc.metadata,
			}

			result := scorer.Score(opp)

			found := false
			for _, s := range result.Signals {
				if s.Name == "high_engagement" && s.Matched {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("expected high_engagement signal to match for %s", tc.name)
			}
		})
	}
}

func TestScore_Recency(t *testing.T) {
	scorer := New()

	// Recent opportunity
	recent := Opportunity{
		Title:      "Test",
		DetectedAt: time.Now(),
		Metadata:   map[string]any{},
	}

	// Old opportunity
	old := Opportunity{
		Title:      "Test",
		DetectedAt: time.Now().Add(-48 * time.Hour),
		Metadata:   map[string]any{},
	}

	recentResult := scorer.Score(recent)
	oldResult := scorer.Score(old)

	recentMatched := false
	for _, s := range recentResult.Signals {
		if s.Name == "recent" && s.Matched {
			recentMatched = true
			break
		}
	}

	oldMatched := false
	for _, s := range oldResult.Signals {
		if s.Name == "recent" && s.Matched {
			oldMatched = true
			break
		}
	}

	if !recentMatched {
		t.Error("expected recent signal to match for recent opportunity")
	}

	if oldMatched {
		t.Error("expected recent signal NOT to match for old opportunity")
	}
}

func TestGetMatchedSignals(t *testing.T) {
	result := Result{
		Score: 50,
		Signals: []Signal{
			{Name: "a", Matched: true},
			{Name: "b", Matched: false},
			{Name: "c", Matched: true},
			{Name: "d", Matched: false},
		},
	}

	matched := result.GetMatchedSignals()

	if len(matched) != 2 {
		t.Errorf("expected 2 matched signals, got %d", len(matched))
	}

	for _, s := range matched {
		if !s.Matched {
			t.Error("GetMatchedSignals returned unmatched signal")
		}
	}
}

func TestContainsAny(t *testing.T) {
	testCases := []struct {
		text     string
		keywords []string
		expected bool
	}{
		{"Hello World", []string{"hello"}, true},
		{"HELLO WORLD", []string{"hello"}, true},
		{"Hello World", []string{"foo", "bar"}, false},
		{"I have a PROBLEM", []string{"problem"}, true},
		{"", []string{"test"}, false},
		{"test", []string{}, false},
	}

	for _, tc := range testCases {
		result := containsAny(tc.text, tc.keywords)
		if result != tc.expected {
			t.Errorf("containsAny(%q, %v) = %v, expected %v", tc.text, tc.keywords, result, tc.expected)
		}
	}
}

func TestScore_Normalized(t *testing.T) {
	scorer := New()

	// Even with all signals matching, score should be <= 100
	opp := Opportunity{
		Title:       "Show HN: I built an open source API tool to solve a problem",
		Description: "Looking for feedback. How do I get customers for my indie SaaS startup?",
		SourceType:  "hackernews",
		DetectedAt:  time.Now(),
		Metadata: map[string]any{
			"points":       200,
			"num_comments": 100,
		},
	}

	result := scorer.Score(opp)

	if result.Score > 100 {
		t.Errorf("score should be normalized to max 100, got %d", result.Score)
	}

	if result.Score < 0 {
		t.Errorf("score should not be negative, got %d", result.Score)
	}
}
