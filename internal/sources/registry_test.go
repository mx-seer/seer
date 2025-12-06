//go:build !pro

package sources

import (
	"testing"
)

func TestGetAvailableTypes_CE(t *testing.T) {
	types := GetAvailableTypes()

	expected := []string{"hackernews", "github", "npm", "devto", "rss"}
	if len(types) != len(expected) {
		t.Errorf("expected %d types, got %d", len(expected), len(types))
	}

	typeMap := make(map[string]bool)
	for _, tp := range types {
		typeMap[tp] = true
	}

	for _, exp := range expected {
		if !typeMap[exp] {
			t.Errorf("expected type %s to be available", exp)
		}
	}

	// Pro-only types should not be available
	proTypes := []string{"reddit", "twitter", "custom"}
	for _, pro := range proTypes {
		if typeMap[pro] {
			t.Errorf("type %s should not be available in CE", pro)
		}
	}
}

func TestMaxRSSFeeds_CE(t *testing.T) {
	max := MaxRSSFeeds()
	if max != 2 {
		t.Errorf("expected max 2 RSS feeds in CE, got %d", max)
	}
}

func TestIsPro_CE(t *testing.T) {
	if IsPro() {
		t.Error("expected IsPro() to be false in CE edition")
	}
}
