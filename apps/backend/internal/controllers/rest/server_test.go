package rest

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/domain"
)

type mockZettelSearchRepo struct{}

func (m *mockZettelSearchRepo) Save(_ context.Context, _ *domain.Zettel) error { return nil }
func (m *mockZettelSearchRepo) SearchZettels(_ context.Context, _ string, _ int) ([]*domain.Zettel, error) {
	return []*domain.Zettel{}, nil
}

func TestHealthEndpoint(t *testing.T) {
	server := newTestServer()
	
	t.Run("Valid GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusOK)
		assertJSONContentType(t, res)
	})

	t.Run("Invalid Method", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusMethodNotAllowed)
	})
}

func TestSyncEndpoints(t *testing.T) {
	server := newTestServer()

	t.Run("Sync Operations POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/operations", strings.NewReader("{}"))
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusNotImplemented)
	})

	t.Run("Sync Changes GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/changes", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusNotImplemented)
	})

	t.Run("Sync Operations GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/operations", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusMethodNotAllowed)
	})

	t.Run("Sync Changes POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/changes", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusMethodNotAllowed)
	})

	t.Run("Sync Bootstrap GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/bootstrap", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusNotImplemented)
	})

	t.Run("Sync Bootstrap POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/bootstrap", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusMethodNotAllowed)
	})
}

func TestNotFound(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
	res := httptest.NewRecorder()
	server.Handler().ServeHTTP(res, req)
	assertStatus(t, res, http.StatusNotFound)
}

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(config.Config{HTTPAddr: ":0"}, logger, nil, nil)
}

func assertStatus(t *testing.T, res *httptest.ResponseRecorder, expected int) {
	t.Helper()
	if res.Code != expected {
		t.Fatalf("expected HTTP status %d, got %d with body %q", expected, res.Code, res.Body.String())
	}
}

func assertJSONContentType(t *testing.T, res *httptest.ResponseRecorder) {
	t.Helper()
	if contentType := res.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "application/json") {
		t.Fatalf("expected JSON content type, got %q", contentType)
	}
}
