package services

import (
	"context"
	"fmt"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type SearchService struct {
	zettels repositories.ZettelRepository
}

func NewSearchService(zettels repositories.ZettelRepository) *SearchService {
	return &SearchService{zettels: zettels}
}

func (s *SearchService) Search(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}
	if limit <= 0 {
		limit = 10
	}
	return s.zettels.SearchZettels(ctx, query, limit)
}
