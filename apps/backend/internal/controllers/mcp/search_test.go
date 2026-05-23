package mcp

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type mockZettelSearchRepo struct {
	results []*domain.Zettel
}

func (m *mockZettelSearchRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockZettelSearchRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if query == "fail" {
		return nil, http.ErrHandlerTimeout
	}
	return m.results, nil
}

func TestNewHandler(t *testing.T) {
	cfg := config.Config{AgentToken: "test-token"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	searchSvc := services.NewSearchService(&mockZettelSearchRepo{})
	handler := NewHandler(cfg, logger, nil, searchSvc, nil)
	if handler == nil {
		t.Fatal("expected non-nil handler")
	}

	t.Run("Trigger Factory", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/mcp", nil)
		res := httptest.NewRecorder()
		handler.ServeHTTP(res, req)
	})
}

func TestSearchToolHandler(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	repo := &mockZettelSearchRepo{
		results: []*domain.Zettel{
			{ID: "z1", Title: "Title 1", Lifecycle: "evergreen"},
		},
	}
	searchSvc := services.NewSearchService(repo)
	handler := searchToolHandler(logger, searchSvc)

	t.Run("Successful Search", func(t *testing.T) {
		_, output, err := handler(context.Background(), nil, SearchInput{Query: "test", Limit: 5})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(output.Results) != 1 {
			t.Errorf("expected 1 result, got %d", len(output.Results))
		}
	})

	t.Run("Normalization Error", func(t *testing.T) {
		_, _, err := handler(context.Background(), nil, SearchInput{Query: ""})
		if err == nil {
			t.Error("expected normalization error")
		}
	})

	t.Run("Service Error", func(t *testing.T) {
		_, _, err := handler(context.Background(), nil, SearchInput{Query: "fail"})
		if err == nil {
			t.Error("expected service error")
		}
	})
}

func TestNormalizeSearchInput(t *testing.T) {
	t.Run("Empty Query", func(t *testing.T) {
		_, err := normalizeSearchInput(SearchInput{Query: " "})
		if err == nil {
			t.Error("expected error for empty query")
		}
	})

	t.Run("Negative Limit", func(t *testing.T) {
		_, err := normalizeSearchInput(SearchInput{Query: "q", Limit: -1})
		if err == nil {
			t.Error("expected error for negative limit")
		}
	})

	t.Run("Excessive Limit", func(t *testing.T) {
		_, err := normalizeSearchInput(SearchInput{Query: "q", Limit: 100})
		if err == nil {
			t.Error("expected error for excessive limit")
		}
	})
}
