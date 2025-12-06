package sources

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mmcdole/gofeed"
)

func TestRSS_Type(t *testing.T) {
	r, _ := NewRSS(SourceConfig{Name: "Test RSS", URL: "http://example.com/feed.xml"})
	if r.Type() != "rss" {
		t.Errorf("expected type rss, got %s", r.Type())
	}
}

func TestRSS_Name(t *testing.T) {
	r, _ := NewRSS(SourceConfig{Name: "Test RSS", URL: "http://example.com/feed.xml"})
	if r.Name() != "Test RSS" {
		t.Errorf("expected name 'Test RSS', got %s", r.Name())
	}
}

func TestRSS_NewRSS_RequiresURL(t *testing.T) {
	_, err := NewRSS(SourceConfig{Name: "Test RSS"})
	if err == nil {
		t.Error("expected error when URL is empty")
	}
}

func TestRSS_Fetch(t *testing.T) {
	feedXML := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>http://example.com</link>
    <item>
      <title>Test Article</title>
      <link>http://example.com/article1</link>
      <description>This is a test article</description>
      <guid>article-1</guid>
      <pubDate>Mon, 01 Jan 2024 12:00:00 GMT</pubDate>
    </item>
    <item>
      <title>Another Article</title>
      <link>http://example.com/article2</link>
      <description>Another test article</description>
      <guid>article-2</guid>
      <pubDate>Mon, 01 Jan 2024 10:00:00 GMT</pubDate>
    </item>
  </channel>
</rss>`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/xml")
		w.Write([]byte(feedXML))
	}))
	defer server.Close()

	rss, err := NewRSS(SourceConfig{Name: "Test", URL: server.URL})
	if err != nil {
		t.Fatalf("failed to create RSS source: %v", err)
	}

	opportunities, err := rss.Fetch(context.Background())
	if err != nil {
		t.Fatalf("failed to fetch: %v", err)
	}

	if len(opportunities) != 2 {
		t.Errorf("expected 2 opportunities, got %d", len(opportunities))
	}

	if opportunities[0].Title != "Test Article" {
		t.Errorf("expected title 'Test Article', got %s", opportunities[0].Title)
	}

	if opportunities[0].SourceIDExternal != "article-1" {
		t.Errorf("expected source ID 'article-1', got %s", opportunities[0].SourceIDExternal)
	}

	if opportunities[0].Metadata["feed_title"] != "Test Feed" {
		t.Errorf("expected feed_title 'Test Feed', got %v", opportunities[0].Metadata["feed_title"])
	}
}

func TestRSS_ItemToOpportunity(t *testing.T) {
	r := &RSS{config: SourceConfig{Name: "Test", URL: "http://example.com/feed.xml"}}

	now := time.Now()
	item := &gofeed.Item{
		Title:           "Great Article",
		Description:     "A great description",
		Link:            "http://example.com/great-article",
		GUID:            "great-article-123",
		PublishedParsed: &now,
		Author:          &gofeed.Person{Name: "Author Name"},
		Categories:      []string{"tech", "news"},
	}

	opp := r.itemToOpportunity(item, "My Feed")

	if opp.Title != "Great Article" {
		t.Errorf("expected title 'Great Article', got %s", opp.Title)
	}

	if opp.Description != "A great description" {
		t.Errorf("expected description 'A great description', got %s", opp.Description)
	}

	if opp.SourceType != "rss" {
		t.Errorf("expected source type rss, got %s", opp.SourceType)
	}

	if opp.SourceIDExternal != "great-article-123" {
		t.Errorf("expected source ID 'great-article-123', got %s", opp.SourceIDExternal)
	}

	if opp.Metadata["author"] != "Author Name" {
		t.Errorf("expected author 'Author Name', got %v", opp.Metadata["author"])
	}

	if opp.Metadata["feed_title"] != "My Feed" {
		t.Errorf("expected feed_title 'My Feed', got %v", opp.Metadata["feed_title"])
	}
}

func TestRSS_ItemToOpportunity_NoGUID(t *testing.T) {
	r := &RSS{config: SourceConfig{Name: "Test", URL: "http://example.com/feed.xml"}}

	item := &gofeed.Item{
		Title: "No GUID Article",
		Link:  "http://example.com/no-guid",
	}

	opp := r.itemToOpportunity(item, "Feed")

	// Should use link as external ID
	if opp.SourceIDExternal != "http://example.com/no-guid" {
		t.Errorf("expected source ID to be link, got %s", opp.SourceIDExternal)
	}
}

func TestRSS_ItemToOpportunity_LongContent(t *testing.T) {
	r := &RSS{config: SourceConfig{Name: "Test", URL: "http://example.com/feed.xml"}}

	longContent := ""
	for i := 0; i < 600; i++ {
		longContent += "x"
	}

	item := &gofeed.Item{
		Title:   "Long Content Article",
		Link:    "http://example.com/long",
		GUID:    "long-123",
		Content: longContent,
	}

	opp := r.itemToOpportunity(item, "Feed")

	// Should be truncated to 500 chars + "..."
	if len(opp.Description) != 503 {
		t.Errorf("expected description length 503, got %d", len(opp.Description))
	}
}

func TestRSS_FetchWithCancelledContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.Write([]byte(`<?xml version="1.0"?><rss><channel></channel></rss>`))
	}))
	defer server.Close()

	rss, _ := NewRSS(SourceConfig{Name: "Test", URL: server.URL})

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := rss.Fetch(ctx)
	if err == nil {
		t.Error("expected error with cancelled context")
	}
}
