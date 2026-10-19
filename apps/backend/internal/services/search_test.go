package services

import (
	"context"
	"fmt"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
)

type mockSearchZettelRepo struct {
	results []*domain.Zettel
	zettels map[string]*domain.Zettel
	fail    bool
}

func (m *mockSearchZettelRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockSearchZettelRepo) FindByID(ctx context.Context, id string) (*domain.Zettel, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return m.zettels[id], nil
}
func (m *mockSearchZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if m.fail {
		return nil, fmt.Errorf("search failed")
	}
	return m.results, nil
}

func TestSearchService(t *testing.T) {
	zRepo := &mockSearchZettelRepo{
		zettels: map[string]*domain.Zettel{
			"z1": {ID: "z1", Title: "Z1", Body: "Content 1", Lifecycle: "zettel"},
			"z2": {ID: "z2", Title: "Z2", Body: "Content 2", Lifecycle: "evergreen"},
		},
		results: []*domain.Zettel{{ID: "z1", Title: "Keyword Match", Body: "Match body", Lifecycle: "zettel"}},
	}
	zRepo.fail = false
	
	vRepo := &mockVectorRepo{
		ids: []string{"z1", "z2"},
	}
	embeds := &mockEmbeddingsClient{}
	
	service := NewSearchService(zRepo, vRepo, embeds)
	ctx := context.Background()

	t.Run("Hybrid Search Success", func(t *testing.T) {
		results, err := service.Search(ctx, "test", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) < 2 {
			t.Errorf("expected at least 2 results (z1, z2), got %d", len(results))
		}
	})

	t.Run("Keyword Only (Semantic Nil)", func(t *testing.T) {
		s2 := NewSearchService(zRepo, nil, nil)
		res, _ := s2.Search(ctx, "test", 5)
		if len(res) == 0 {
			t.Error("expected results")
		}
	})

	t.Run("Limit Keywords", func(t *testing.T) {
		zRepo.results = []*domain.Zettel{
			{ID: "1", Body: "1"}, {ID: "2", Body: "2"}, {ID: "3", Body: "3"},
		}
		s2 := NewSearchService(zRepo, nil, nil)
		res, _ := s2.Search(ctx, "test", 2)
		if len(res) != 2 {
			t.Errorf("expected 2 results, got %d", len(res))
		}
	})

	t.Run("Semantic Returns No Results", func(t *testing.T) {
		zRepo.results = []*domain.Zettel{{ID: "z1", Body: "B"}}
		vRepo.ids = []string{}
		results, err := service.Search(ctx, "test", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected keyword fallback results")
		}
		vRepo.ids = []string{"z1", "z2"} // restore
	})

	t.Run("Keyword Returns No Results", func(t *testing.T) {
		zRepo.results = []*domain.Zettel{}
		results, err := service.Search(ctx, "test", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected semantic results")
		}
	})

	t.Run("Semantic Returns Missing ID", func(t *testing.T) {
		vRepo.ids = []string{"missing"}
		zRepo.results = []*domain.Zettel{}
		results, err := service.Search(ctx, "test", 5)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("expected 0 results for missing ID, got %d", len(results))
		}
	})

	t.Run("Hybrid RRF Sorting", func(t *testing.T) {
		zRepo.results = []*domain.Zettel{{ID: "z1", Body: "B1"}, {ID: "z2", Body: "B2"}}
		vRepo.ids = []string{"z2", "z1"}
		results, _ := service.Search(ctx, "test", 10)
		if len(results) < 2 {
			t.Fatalf("expected 2 results")
		}
	})

	t.Run("Both Fail", func(t *testing.T) {
		zRepo.fail = true
		embeds.fail = true
		defer func() { zRepo.fail = false; embeds.fail = false }()
		_, err := service.Search(ctx, "test", 5)
		if err == nil {
			t.Error("expected error when both fail")
		}
	})
}
