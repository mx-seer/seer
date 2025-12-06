//go:build !pro

package sources

// GetAvailableTypes returns the source types available in CE edition
func GetAvailableTypes() []string {
	return []string{"hackernews", "github", "npm", "devto", "rss"}
}

// MaxRSSFeeds returns the maximum number of RSS feeds allowed in CE
func MaxRSSFeeds() int {
	return 2
}

// IsPro returns false for CE edition
func IsPro() bool {
	return false
}
