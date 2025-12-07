//go:build pro

package sources

import (
	"database/sql"
	"encoding/json"
)

// KeywordsConfig stores custom keywords configuration for Pro users
type KeywordsConfig struct {
	// Include filters - opportunities must match at least one
	IncludeKeywords []string `json:"include_keywords"`

	// Exclude filters - opportunities matching any are filtered out
	ExcludeKeywords []string `json:"exclude_keywords"`

	// Boost keywords - add score bonus for matching
	BoostKeywords []string `json:"boost_keywords"`
}

// KeywordsRepository manages custom keywords in the database
type KeywordsRepository struct {
	db *sql.DB
}

// NewKeywordsRepository creates a new keywords repository
func NewKeywordsRepository(db *sql.DB) *KeywordsRepository {
	return &KeywordsRepository{db: db}
}

// GetKeywords retrieves the keywords configuration
func (r *KeywordsRepository) GetKeywords() (*KeywordsConfig, error) {
	var value string
	err := r.db.QueryRow(`SELECT value FROM settings WHERE key = 'custom_keywords'`).Scan(&value)
	if err == sql.ErrNoRows {
		return &KeywordsConfig{}, nil
	}
	if err != nil {
		return nil, err
	}

	var config KeywordsConfig
	if err := json.Unmarshal([]byte(value), &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// SaveKeywords saves the keywords configuration
func (r *KeywordsRepository) SaveKeywords(config *KeywordsConfig) error {
	data, err := json.Marshal(config)
	if err != nil {
		return err
	}

	_, err = r.db.Exec(`
		INSERT INTO settings (key, value) VALUES ('custom_keywords', ?)
		ON CONFLICT(key) DO UPDATE SET value = excluded.value
	`, string(data))

	return err
}

// FilterOpportunities applies keyword filters to opportunities
func FilterOpportunities(opps []Opportunity, config *KeywordsConfig) []Opportunity {
	if config == nil || (len(config.IncludeKeywords) == 0 && len(config.ExcludeKeywords) == 0) {
		return opps
	}

	var filtered []Opportunity
	for _, opp := range opps {
		text := opp.Title + " " + opp.Description

		// Check exclude keywords first
		if len(config.ExcludeKeywords) > 0 && containsAnyKeyword(text, config.ExcludeKeywords) {
			continue
		}

		// Check include keywords if configured
		if len(config.IncludeKeywords) > 0 && !containsAnyKeyword(text, config.IncludeKeywords) {
			continue
		}

		filtered = append(filtered, opp)
	}

	return filtered
}

// CalculateKeywordBoost returns a score boost for matching boost keywords
func CalculateKeywordBoost(opp Opportunity, config *KeywordsConfig) int {
	if config == nil || len(config.BoostKeywords) == 0 {
		return 0
	}

	text := opp.Title + " " + opp.Description
	boost := 0

	for _, kw := range config.BoostKeywords {
		if containsAnyKeyword(text, []string{kw}) {
			boost += 5 // +5 points per matching boost keyword
		}
	}

	// Cap boost at 20 points
	if boost > 20 {
		boost = 20
	}

	return boost
}
