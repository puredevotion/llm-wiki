package services

import (
	"context"
	"fmt"
	"sort"

	"llm-wiki/apps/backend/internal/clients/embeddings"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type SearchService struct {
	zettels repositories.ZettelRepository
	vectors repositories.VectorRepository
	embeds  embeddings.Client
}

func NewSearchService(zettels repositories.ZettelRepository, vectors repositories.VectorRepository, embeds embeddings.Client) *SearchService {
	return &SearchService{zettels: zettels, vectors: vectors, embeds: embeds}
}

func (s *SearchService) Search(ctx context.Context, query string, limit int) ([]*domain.SearchResult, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 10
	}

	// 1. Keyword Search (FTS5)
	keywordResults, kErr := s.zettels.SearchZettels(ctx, query, limit*2)
	if kErr != nil {
		keywordResults = []*domain.Zettel{}
	}

	// 2. Semantic Search
	var semanticIDs []string
	var sErr error
	if s.embeds != nil && s.vectors != nil {
		var queryVec domain.Vector
		queryVec, sErr = s.embeds.Generate(ctx, query)
		if sErr == nil {
			semanticIDs, _ = s.vectors.Search(ctx, "zettel", queryVec, limit*2)
		}
	}

	if kErr != nil && sErr != nil {
		return nil, fmt.Errorf("hybrid search failed: keyword(%w), semantic(%w)", kErr, sErr)
	}

	// 3. Hybrid RRF Ranking
	type hybridResult struct {
		z     *domain.Zettel
		score float64
	}
	combined := make(map[string]*hybridResult)
	k := 60.0 // RRF constant

	// Add keyword ranks
	for i, z := range keywordResults {
		combined[z.ID] = &hybridResult{
			z:     z,
			score: 1.0 / (k + float64(i+1)),
		}
	}

	// Add semantic ranks
	for i, id := range semanticIDs {
		if res, ok := combined[id]; ok {
			res.score += 1.0 / (k + float64(i+1))
		} else {
			// Fetch missing Zettel
			z, err := s.zettels.FindByID(ctx, id)
			if err == nil && z != nil {
				combined[id] = &hybridResult{
					z:     z,
					score: 1.0 / (k + float64(i+1)),
				}
			}
		}
	}

	// Convert to slice and sort
	final := make([]*hybridResult, 0, len(combined))
	for _, v := range combined {
		final = append(final, v)
	}

	sort.Slice(final, func(i, j int) bool {
		return final[i].score > final[j].score
	})

	if len(final) > limit {
		final = final[:limit]
	}

	results := make([]*domain.SearchResult, 0, len(final))
	for _, f := range final {
		snippet := f.z.Body
		if len(snippet) > 200 {
			snippet = snippet[:197] + "..."
		}
		results = append(results, &domain.SearchResult{
			ID:        f.z.ID,
			Title:     f.z.Title,
			Snippet:   snippet,
			Lifecycle: f.z.Lifecycle,
			Score:     f.score,
		})
	}

	// Fallback to keyword only if semantic failed entirely
	if len(results) == 0 && len(keywordResults) > 0 {
		count := len(keywordResults)
		if count > limit {
			count = limit
		}
		for i := 0; i < count; i++ {
			z := keywordResults[i]
			snippet := z.Body
			if len(snippet) > 200 {
				snippet = snippet[:197] + "..."
			}
			results = append(results, &domain.SearchResult{
				ID:        z.ID,
				Title:     z.Title,
				Snippet:   snippet,
				Lifecycle: z.Lifecycle,
				Score:     0,
			})
		}
	}

	return results, nil
}
