package report

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// Opportunity represents an opportunity for the report
type Opportunity struct {
	ID          int64
	Title       string
	Description string
	SourceType  string
	SourceURL   string
	Score       int
	Signals     []string
	DetectedAt  time.Time
}

// Report contains the generated report
type Report struct {
	PeriodStart      time.Time
	PeriodEnd        time.Time
	OpportunityCount int
	ContentHuman     string
	ContentPrompt    string
	Opportunities    []Opportunity
}

// Generator creates reports from opportunities
type Generator struct{}

// New creates a new report generator
func New() *Generator {
	return &Generator{}
}

// Generate creates a report from opportunities
func (g *Generator) Generate(opportunities []Opportunity, periodStart, periodEnd time.Time) *Report {
	// Sort by score descending
	sorted := make([]Opportunity, len(opportunities))
	copy(sorted, opportunities)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Score > sorted[j].Score
	})

	report := &Report{
		PeriodStart:      periodStart,
		PeriodEnd:        periodEnd,
		OpportunityCount: len(opportunities),
		Opportunities:    sorted,
	}

	report.ContentHuman = g.generateHumanReadable(sorted, periodStart, periodEnd)
	report.ContentPrompt = g.generatePrompt(sorted, periodStart, periodEnd)

	return report
}

// generateHumanReadable creates a human-readable report
func (g *Generator) generateHumanReadable(opportunities []Opportunity, start, end time.Time) string {
	var sb strings.Builder

	sb.WriteString("# Seer Opportunity Report\n\n")
	sb.WriteString(fmt.Sprintf("**Period:** %s to %s\n", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006")))
	sb.WriteString(fmt.Sprintf("**Total Opportunities:** %d\n\n", len(opportunities)))

	if len(opportunities) == 0 {
		sb.WriteString("No opportunities found in this period.\n")
		return sb.String()
	}

	sb.WriteString("---\n\n")

	// Top opportunities
	sb.WriteString("## Top Opportunities\n\n")

	limit := 20
	if len(opportunities) < limit {
		limit = len(opportunities)
	}

	for i, opp := range opportunities[:limit] {
		sb.WriteString(fmt.Sprintf("### %d. %s\n\n", i+1, opp.Title))
		sb.WriteString(fmt.Sprintf("**Score:** %d/100 | **Source:** %s\n\n", opp.Score, opp.SourceType))

		if opp.Description != "" {
			desc := opp.Description
			if len(desc) > 300 {
				desc = desc[:300] + "..."
			}
			sb.WriteString(fmt.Sprintf("%s\n\n", desc))
		}

		if len(opp.Signals) > 0 {
			sb.WriteString(fmt.Sprintf("**Signals:** %s\n\n", strings.Join(opp.Signals, ", ")))
		}

		sb.WriteString(fmt.Sprintf("**Link:** %s\n\n", opp.SourceURL))
		sb.WriteString("---\n\n")
	}

	// Summary by source
	sb.WriteString("## Summary by Source\n\n")

	sourceCounts := make(map[string]int)
	sourceScores := make(map[string]int)
	for _, opp := range opportunities {
		sourceCounts[opp.SourceType]++
		sourceScores[opp.SourceType] += opp.Score
	}

	for source, count := range sourceCounts {
		avgScore := 0
		if count > 0 {
			avgScore = sourceScores[source] / count
		}
		sb.WriteString(fmt.Sprintf("- **%s:** %d opportunities (avg score: %d)\n", source, count, avgScore))
	}

	return sb.String()
}

// generatePrompt creates an AI-optimized prompt
func (g *Generator) generatePrompt(opportunities []Opportunity, start, end time.Time) string {
	var sb strings.Builder

	sb.WriteString("You are an expert market analyst specializing in opportunities for indie developers and bootstrapped startups.\n\n")
	sb.WriteString("Analyze the following market opportunities detected from various sources and provide:\n")
	sb.WriteString("1. A summary of the most promising opportunities\n")
	sb.WriteString("2. Common themes and patterns you notice\n")
	sb.WriteString("3. Specific actionable ideas for indie developers\n")
	sb.WriteString("4. Any emerging trends worth watching\n\n")

	sb.WriteString(fmt.Sprintf("Report Period: %s to %s\n", start.Format("2006-01-02"), end.Format("2006-01-02")))
	sb.WriteString(fmt.Sprintf("Total Opportunities: %d\n\n", len(opportunities)))

	sb.WriteString("=== OPPORTUNITIES ===\n\n")

	limit := 30 // More items for AI analysis
	if len(opportunities) < limit {
		limit = len(opportunities)
	}

	for i, opp := range opportunities[:limit] {
		sb.WriteString(fmt.Sprintf("[%d] %s\n", i+1, opp.Title))
		sb.WriteString(fmt.Sprintf("Source: %s | Score: %d/100\n", opp.SourceType, opp.Score))

		if opp.Description != "" {
			desc := opp.Description
			if len(desc) > 500 {
				desc = desc[:500] + "..."
			}
			sb.WriteString(fmt.Sprintf("Description: %s\n", desc))
		}

		if len(opp.Signals) > 0 {
			sb.WriteString(fmt.Sprintf("Signals: %s\n", strings.Join(opp.Signals, ", ")))
		}

		sb.WriteString(fmt.Sprintf("URL: %s\n", opp.SourceURL))
		sb.WriteString("\n")
	}

	sb.WriteString("=== END OPPORTUNITIES ===\n\n")
	sb.WriteString("Please provide your analysis in a structured, actionable format.")

	return sb.String()
}

// GetTopOpportunities returns the top N opportunities by score
func (r *Report) GetTopOpportunities(n int) []Opportunity {
	if n > len(r.Opportunities) {
		n = len(r.Opportunities)
	}
	return r.Opportunities[:n]
}
