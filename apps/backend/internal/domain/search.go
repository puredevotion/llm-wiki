package domain

// SearchResult represents a single entry in search results, enriched with metadata.
type SearchResult struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	Snippet   string  `json:"snippet"`
	Lifecycle string  `json:"lifecycle"`
	Score     float64 `json:"score"`
}
