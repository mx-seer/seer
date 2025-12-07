package report

// CopyablePrompt generates a prompt optimized for copying to external AI
func CopyablePrompt(r *Report) string {
	return r.ContentPrompt
}

// DefaultPromptTemplate returns the default prompt template
func DefaultPromptTemplate() string {
	return `You are an expert market analyst specializing in opportunities for indie developers and bootstrapped startups.

Analyze the following market opportunities detected from various sources and provide:
1. A summary of the most promising opportunities
2. Common themes and patterns you notice
3. Specific actionable ideas for indie developers
4. Any emerging trends worth watching

Please provide your analysis in a structured, actionable format.`
}
