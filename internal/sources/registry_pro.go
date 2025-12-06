//go:build pro

package sources

// GetAvailableTypes returns the source types available in Pro edition
func GetAvailableTypes() []string {
	return []string{"hackernews", "github", "npm", "devto", "rss", "reddit", "twitter", "custom"}
}

// MaxRSSFeeds returns unlimited RSS feeds for Pro (-1 means unlimited)
func MaxRSSFeeds() int {
	return -1
}

// IsPro returns true for Pro edition
func IsPro() bool {
	return true
}
