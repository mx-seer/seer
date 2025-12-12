package sources

import (
	"testing"
)

func TestGetAvailableTypes(t *testing.T) {
	types := GetAvailableTypes()

	expected := []string{"hackernews", "github", "npm", "devto", "reddit", "twitter", "custom"}
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
}
