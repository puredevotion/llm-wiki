package mcp

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type mockZettelSearchSvcRepo struct {
	results []*domain.Zettel
}

func (m *mockZettelSearchSvcRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockZettelSearchSvcRepo) FindByID(ctx context.Context, id string) (*domain.Zettel, error) {
	return nil, nil
}
func (m *mockZettelSearchSvcRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if query == "fail" {
		return nil, fmt.Errorf("keyword fail")
	}
	return m.results, nil
}

type mockSearchEmbeddingsClient struct{}

func (m *mockSearchEmbeddingsClient) Generate(ctx context.Context, text string) (domain.Vector, error) {
	if text == "fail" {
		return nil, fmt.Errorf("semantic fail")
	}
	return domain.Vector{0.1}, nil
}
func (m *mockSearchEmbeddingsClient) BatchGenerate(ctx context.Context, texts []string) ([]domain.Vector, error) {
	return nil, nil
}

func TestNewHandler(t *testing.T) {
	cfg := config.Config{AgentToken: "test-token"}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	searchSvc := services.NewSearchService(&mockZettelSearchSvcRepo{}, &mockVectorRepo{}, &mockSearchEmbeddingsClient{})
	handler := NewHandler(cfg, logger, nil, searchSvc, nil, nil)
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
	repo := &mockZettelSearchSvcRepo{
		results: []*domain.Zettel{
			{ID: "z1", Title: "Title 1", Lifecycle: "evergreen"},
		},
	}
	searchSvc := services.NewSearchService(repo, &mockVectorRepo{}, &mockSearchEmbeddingsClient{})
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
