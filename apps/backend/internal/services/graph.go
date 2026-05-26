package services

import (
	"context"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type GraphService struct {
	graph repositories.GraphRepository
}

func NewGraphService(graph repositories.GraphRepository) *GraphService {
	return &GraphService{graph: graph}
}

func (s *GraphService) GetGraph(ctx context.Context) (*domain.GraphData, error) {
	return s.graph.FetchGraph(ctx)
}
