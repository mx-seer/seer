package sources

import (
	"context"
	"testing"
	"time"
)

func TestDevTo_Type(t *testing.T) {
	d, _ := NewDevTo(SourceConfig{Name: "Test DEV.to"})
	if d.Type() != "devto" {
		t.Errorf("expected type devto, got %s", d.Type())
	}
}

func TestDevTo_Name(t *testing.T) {
	d, _ := NewDevTo(SourceConfig{Name: "Test DEV.to"})
	if d.Name() != "Test DEV.to" {
		t.Errorf("expected name 'Test DEV.to', got %s", d.Name())
	}
}

func TestDevTo_ArticleToOpportunity(t *testing.T) {
	d := &DevTo{config: SourceConfig{Name: "Test"}}

	now := time.Now()
	article := devtoArticle{
		ID:          12345,
		Title:       "How I Built My Side Project",
		Description: "A journey of building something cool",
		URL:         "https://dev.to/user/how-i-built-my-side-project",
		PublishedAt: now,
		TagList:     []string{"showdev", "sideproject", "go"},
		PositiveReactionsCount: 42,
		CommentsCount:          15,
		ReadingTimeMinutes:     8,
	}
	article.User.Name = "John Developer"
	article.User.Username = "johndev"

	opp := d.articleToOpportunity(article)

	if opp.Title != "How I Built My Side Project" {
		t.Errorf("expected title 'How I Built My Side Project', got %s", opp.Title)
	}

	if opp.Description != "A journey of building something cool" {
		t.Errorf("expected description 'A journey of building something cool', got %s", opp.Description)
	}

	if opp.SourceType != "devto" {
		t.Errorf("expected source type devto, got %s", opp.SourceType)
	}

	if opp.SourceIDExternal != "12345" {
		t.Errorf("expected source ID '12345', got %s", opp.SourceIDExternal)
	}

	if opp.Metadata["author"] != "John Developer" {
		t.Errorf("expected author 'John Developer', got %v", opp.Metadata["author"])
	}

	if opp.Metadata["username"] != "johndev" {
		t.Errorf("expected username 'johndev', got %v", opp.Metadata["username"])
	}

	if opp.Metadata["reactions"] != 42 {
		t.Errorf("expected reactions 42, got %v", opp.Metadata["reactions"])
	}

	if opp.Metadata["reading_time"] != 8 {
		t.Errorf("expected reading_time 8, got %v", opp.Metadata["reading_time"])
	}
}

func TestDevTo_ArticleToOpportunity_NoDescription(t *testing.T) {
	d := &DevTo{config: SourceConfig{Name: "Test"}}

	article := devtoArticle{
		ID:          67890,
		Title:       "Quick Tip",
		URL:         "https://dev.to/user/quick-tip",
		PublishedAt: time.Now(),
	}
	article.User.Name = "Jane Coder"

	opp := d.articleToOpportunity(article)

	expected := "Article by Jane Coder"
	if opp.Description != expected {
		t.Errorf("expected description '%s', got %s", expected, opp.Description)
	}
}

func TestDevTo_FetchWithCancelledContext(t *testing.T) {
	d, _ := NewDevTo(SourceConfig{Name: "Test"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := d.Fetch(ctx)
	if err != nil {
		t.Errorf("expected no error with cancelled context, got %v", err)
	}
}
