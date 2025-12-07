//go:build !pro

package sources

import "database/sql"

// KeywordsConfig stores custom keywords configuration (stub for CE)
type KeywordsConfig struct {
	IncludeKeywords []string `json:"include_keywords"`
	ExcludeKeywords []string `json:"exclude_keywords"`
	BoostKeywords   []string `json:"boost_keywords"`
}

// KeywordsRepository manages custom keywords (stub for CE)
type KeywordsRepository struct {
	db *sql.DB
}

// NewKeywordsRepository creates a new keywords repository (stub for CE)
func NewKeywordsRepository(db *sql.DB) *KeywordsRepository {
	return &KeywordsRepository{db: db}
}

// GetKeywords returns empty config in CE edition
func (r *KeywordsRepository) GetKeywords() (*KeywordsConfig, error) {
	return &KeywordsConfig{}, nil
}

// SaveKeywords does nothing in CE edition
func (r *KeywordsRepository) SaveKeywords(config *KeywordsConfig) error {
	return nil
}

// FilterOpportunities returns opportunities unchanged in CE edition
func FilterOpportunities(opps []Opportunity, config *KeywordsConfig) []Opportunity {
	return opps
}

// CalculateKeywordBoost returns 0 in CE edition
func CalculateKeywordBoost(opp Opportunity, config *KeywordsConfig) int {
	return 0
}
