package services

import (
	"context"
	"fmt"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
)

type mockZettelSearchRepo struct {
	fail bool
}

func (m *mockZettelSearchRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockZettelSearchRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if m.fail {
		return nil, fmt.Errorf("search failed")
	}
	return []*domain.Zettel{{ID: "z1", Title: "Result"}}, nil
}

func TestSearchService(t *testing.T) {
	repo := &mockZettelSearchRepo{}
	service := NewSearchService(repo)
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		results, err := service.Search(ctx, "test", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 result, got %d", len(results))
		}
	})

	t.Run("Empty Query", func(t *testing.T) {
		_, err := service.Search(ctx, "", 5)
		if err == nil {
			t.Errorf("expected error for empty query, got nil")
		}
	})

	t.Run("Default Limit", func(t *testing.T) {
		_, err := service.Search(ctx, "test", 0)
		if err != nil {
			t.Errorf("expected success with default limit, got %v", err)
		}
	})

	t.Run("Repo Error", func(t *testing.T) {
		repo.fail = true
		defer func() { repo.fail = false }()
		_, err := service.Search(ctx, "test", 5)
		if err == nil {
			t.Errorf("expected repo error, got nil")
		}
	})
}
