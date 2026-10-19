package services

import (
	"context"
	"testing"
	"llm-wiki/apps/backend/internal/domain"
)

type mockGraphFetchRepo struct {
	data *domain.GraphData
}

func (m *mockGraphFetchRepo) UpsertNode(ctx context.Context, id, label string, props map[string]any) error {
	return nil
}
func (m *mockGraphFetchRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	return nil
}
func (m *mockGraphFetchRepo) FetchGraph(ctx context.Context) (*domain.GraphData, error) {
	return m.data, nil
}

func TestGraphService(t *testing.T) {
	data := &domain.GraphData{
		Nodes: []domain.GraphNode{{ID: "1", Name: "Node 1"}},
		Links: []domain.GraphLink{{Source: "1", Target: "2"}},
	}
	repo := &mockGraphFetchRepo{data: data}
	svc := NewGraphService(repo)

	t.Run("GetGraph", func(t *testing.T) {
		res, err := svc.GetGraph(context.Background())
		if err != nil {
			t.Fatalf("GetGraph failed: %v", err)
		}
		if len(res.Nodes) != 1 {
			t.Errorf("expected 1 node, got %d", len(res.Nodes))
		}
	})
}
