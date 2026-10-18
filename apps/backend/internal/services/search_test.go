package services

import (
	"context"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
)

func TestSearchService(t *testing.T) {
	zRepo := &mockZettelRepo{
		zettels: map[string]*domain.Zettel{
			"z1": {ID: "z1", Title: "Z1"},
			"z2": {ID: "z2", Title: "Z2"},
		},
		results: []*domain.Zettel{{ID: "z1", Title: "Keyword Match"}},
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
			{ID: "1"}, {ID: "2"}, {ID: "3"},
		}
		s2 := NewSearchService(zRepo, nil, nil)
		res, _ := s2.Search(ctx, "test", 2)
		if len(res) != 2 {
			t.Errorf("expected 2 results, got %d", len(res))
		}
	})

	t.Run("Semantic Returns No Results", func(t *testing.T) {
		zRepo.results = []*domain.Zettel{{ID: "z1"}}
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
		zRepo.results = []*domain.Zettel{{ID: "z1"}, {ID: "z2"}}
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
