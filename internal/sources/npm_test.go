package sources

import (
	"context"
	"testing"
	"time"
)

func TestNPM_Type(t *testing.T) {
	npm, _ := NewNPM(SourceConfig{Name: "Test npm"})
	if npm.Type() != "npm" {
		t.Errorf("expected type npm, got %s", npm.Type())
	}
}

func TestNPM_Name(t *testing.T) {
	npm, _ := NewNPM(SourceConfig{Name: "Test npm"})
	if npm.Name() != "Test npm" {
		t.Errorf("expected name 'Test npm', got %s", npm.Name())
	}
}

func TestNPM_PackageToOpportunity(t *testing.T) {
	npm := &NPM{config: SourceConfig{Name: "Test"}}

	now := time.Now()
	obj := npmObject{
		Package: npmPackage{
			Name:        "awesome-cli",
			Version:     "1.2.3",
			Description: "An awesome CLI tool",
			Keywords:    []string{"cli", "tool", "awesome"},
			Date:        now,
		},
		Score: npmScore{
			Final: 0.85,
			Detail: struct {
				Quality     float64 `json:"quality"`
				Popularity  float64 `json:"popularity"`
				Maintenance float64 `json:"maintenance"`
			}{
				Quality:     0.9,
				Popularity:  0.8,
				Maintenance: 0.85,
			},
		},
	}
	obj.Package.Links.NPM = "https://www.npmjs.com/package/awesome-cli"
	obj.Package.Author.Name = "developer"

	opp := npm.packageToOpportunity(obj)

	if opp.Title != "awesome-cli" {
		t.Errorf("expected title 'awesome-cli', got %s", opp.Title)
	}

	if opp.Description != "An awesome CLI tool" {
		t.Errorf("expected description 'An awesome CLI tool', got %s", opp.Description)
	}

	if opp.SourceType != "npm" {
		t.Errorf("expected source type npm, got %s", opp.SourceType)
	}

	if opp.SourceIDExternal != "awesome-cli@1.2.3" {
		t.Errorf("expected source ID 'awesome-cli@1.2.3', got %s", opp.SourceIDExternal)
	}

	if opp.Metadata["version"] != "1.2.3" {
		t.Errorf("expected version 1.2.3, got %v", opp.Metadata["version"])
	}

	if opp.Metadata["author"] != "developer" {
		t.Errorf("expected author 'developer', got %v", opp.Metadata["author"])
	}

	if opp.Metadata["score"] != 0.85 {
		t.Errorf("expected score 0.85, got %v", opp.Metadata["score"])
	}
}

func TestNPM_PackageToOpportunity_NoDescription(t *testing.T) {
	npm := &NPM{config: SourceConfig{Name: "Test"}}

	obj := npmObject{
		Package: npmPackage{
			Name:    "mystery-package",
			Version: "0.0.1",
			Date:    time.Now(),
		},
	}

	opp := npm.packageToOpportunity(obj)

	expected := "npm package: mystery-package v0.0.1"
	if opp.Description != expected {
		t.Errorf("expected description '%s', got %s", expected, opp.Description)
	}
}

func TestNPM_PackageToOpportunity_NoNPMLink(t *testing.T) {
	npm := &NPM{config: SourceConfig{Name: "Test"}}

	obj := npmObject{
		Package: npmPackage{
			Name:    "my-package",
			Version: "1.0.0",
			Date:    time.Now(),
		},
	}

	opp := npm.packageToOpportunity(obj)

	expected := npmPackageURL + "my-package"
	if opp.SourceURL != expected {
		t.Errorf("expected URL '%s', got %s", expected, opp.SourceURL)
	}
}

func TestNPM_FetchWithCancelledContext(t *testing.T) {
	npm, _ := NewNPM(SourceConfig{Name: "Test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := npm.Fetch(ctx)
	if err != nil {
		t.Errorf("expected no error with cancelled context, got %v", err)
	}
}
