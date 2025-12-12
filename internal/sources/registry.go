package sources

// GetAvailableTypes returns the available source types
func GetAvailableTypes() []string {
	return []string{"hackernews", "github", "npm", "devto", "reddit", "twitter", "custom"}
}
