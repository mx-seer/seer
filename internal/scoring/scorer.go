package scoring

import (
	"strings"
	"time"
)

// Signal represents a scoring signal with its weight
type Signal struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Weight      float64 `json:"weight"`
	Matched     bool    `json:"matched"`
}

// Opportunity represents the data to be scored
type Opportunity struct {
	Title       string
	Description string
	SourceType  string
	DetectedAt  time.Time
	Metadata    map[string]any
}

// Result contains the scoring result
type Result struct {
	Score   int      `json:"score"`
	Signals []Signal `json:"signals"`
}

// Scorer calculates opportunity scores based on signals
type Scorer struct {
	signals []signalCheck
}

type signalCheck struct {
	signal Signal
	check  func(o Opportunity) bool
}

// New creates a new Scorer with default signals
func New() *Scorer {
	s := &Scorer{}
	s.registerDefaultSignals()
	return s
}

func (s *Scorer) registerDefaultSignals() {
	// Problem indicators
	s.addSignal(Signal{
		Name:        "problem_mention",
		Description: "Mentions a problem or pain point",
		Weight:      15,
	}, func(o Opportunity) bool {
		keywords := []string{"problem", "issue", "frustrated", "annoying", "hate", "wish", "need", "looking for", "struggling"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})

	// Solution seeking
	s.addSignal(Signal{
		Name:        "solution_seeking",
		Description: "Actively seeking a solution",
		Weight:      20,
	}, func(o Opportunity) bool {
		keywords := []string{"how do i", "how to", "best way", "recommend", "alternative", "looking for", "need help", "any suggestions"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})

	// Show HN / Side project
	s.addSignal(Signal{
		Name:        "show_project",
		Description: "Someone showing their project",
		Weight:      10,
	}, func(o Opportunity) bool {
		keywords := []string{"show hn", "showhn", "i built", "i made", "my project", "side project", "launching", "just launched"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})

	// Technical focus
	s.addSignal(Signal{
		Name:        "technical",
		Description: "Technical content (dev tools, APIs, etc)",
		Weight:      10,
	}, func(o Opportunity) bool {
		keywords := []string{"api", "sdk", "library", "framework", "tool", "cli", "developer", "devtool", "open source"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})

	// SaaS/Business keywords
	s.addSignal(Signal{
		Name:        "business_opportunity",
		Description: "Business/SaaS opportunity indicators",
		Weight:      15,
	}, func(o Opportunity) bool {
		keywords := []string{"saas", "startup", "business", "revenue", "customers", "users", "subscription", "pricing", "monetize"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})

	// High engagement (from metadata)
	s.addSignal(Signal{
		Name:        "high_engagement",
		Description: "High engagement (comments, stars, reactions)",
		Weight:      10,
	}, func(o Opportunity) bool {
		// Check various engagement metrics from metadata
		if points, ok := o.Metadata["points"].(int); ok && points > 50 {
			return true
		}
		if comments, ok := o.Metadata["num_comments"].(int); ok && comments > 20 {
			return true
		}
		if stars, ok := o.Metadata["stars"].(int); ok && stars > 100 {
			return true
		}
		if reactions, ok := o.Metadata["reactions"].(int); ok && reactions > 20 {
			return true
		}
		return false
	})

	// Recency bonus
	s.addSignal(Signal{
		Name:        "recent",
		Description: "Posted within last 24 hours",
		Weight:      10,
	}, func(o Opportunity) bool {
		return time.Since(o.DetectedAt) < 24*time.Hour
	})

	// Indie/Solo keywords
	s.addSignal(Signal{
		Name:        "indie_focus",
		Description: "Relevant to indie developers",
		Weight:      10,
	}, func(o Opportunity) bool {
		keywords := []string{"indie", "solo", "bootstrapped", "self-funded", "side project", "maker", "indiehacker", "solopreneur"}
		return containsAny(o.Title+" "+o.Description, keywords)
	})
}

func (s *Scorer) addSignal(signal Signal, check func(o Opportunity) bool) {
	s.signals = append(s.signals, signalCheck{signal: signal, check: check})
}

// Score calculates the score for an opportunity
func (s *Scorer) Score(o Opportunity) Result {
	var totalScore float64
	var matchedSignals []Signal

	for _, sc := range s.signals {
		signal := sc.signal
		signal.Matched = sc.check(o)

		if signal.Matched {
			totalScore += signal.Weight
		}

		matchedSignals = append(matchedSignals, signal)
	}

	// Normalize to 0-100 scale
	maxPossible := 0.0
	for _, sc := range s.signals {
		maxPossible += sc.signal.Weight
	}

	normalizedScore := int((totalScore / maxPossible) * 100)
	if normalizedScore > 100 {
		normalizedScore = 100
	}

	return Result{
		Score:   normalizedScore,
		Signals: matchedSignals,
	}
}

// GetMatchedSignals returns only the matched signals from a result
func (r *Result) GetMatchedSignals() []Signal {
	var matched []Signal
	for _, s := range r.Signals {
		if s.Matched {
			matched = append(matched, s)
		}
	}
	return matched
}

// containsAny checks if text contains any of the keywords (case-insensitive)
func containsAny(text string, keywords []string) bool {
	lower := strings.ToLower(text)
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}
