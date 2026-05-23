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
	"llm-wiki/apps/backend/internal/services"
)

type mockZettelSearchRepo struct{}

func (m *mockZettelSearchRepo) Save(_ context.Context, _ *domain.Zettel) error { return nil }
func (m *mockZettelSearchRepo) SearchZettels(_ context.Context, _ string, _ int) ([]*domain.Zettel, error) {
	return []*domain.Zettel{}, nil
}

type mockOpRepo struct {
	ops map[string]*domain.Operation
}

func (m *mockOpRepo) Save(ctx context.Context, op *domain.Operation) error {
	m.ops[op.ID] = op
	return nil
}
func (m *mockOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	return m.ops[id], nil
}
func (m *mockOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	var results []*domain.Operation
	for _, op := range m.ops {
		results = append(results, op)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

type mockTopicRepo struct{}

func (m *mockTopicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) Save(ctx context.Context, topic *domain.Topic) error { return nil }

type mockActorRepo struct{}

func (m *mockActorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	return nil, nil
}
func (m *mockActorRepo) Save(ctx context.Context, actor *domain.Actor) error { return nil }

type mockIdentityRepo struct{}

func (m *mockIdentityRepo) SaveTeam(ctx context.Context, team *domain.Team) error { return nil }
func (m *mockIdentityRepo) SaveOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}
func (m *mockIdentityRepo) AddTeamMember(ctx context.Context, member *domain.TeamMember) error {
	return nil
}
func (m *mockIdentityRepo) FindTeamByName(ctx context.Context, name string) (*domain.Team, error) {
	return nil, nil
}

type mockGraphRepo struct{}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	return nil
}
func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	return nil
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
		payload := `{"operations": [{"id": "op1", "entity_kind": "zettel", "entity_id": "z1", "payload": {}}]}`
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/operations", strings.NewReader(payload))
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusOK)
		assertJSONContentType(t, res)
	})

	t.Run("Sync Operations Invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/api/v1/sync/operations", strings.NewReader("{bad}"))
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusBadRequest)
	})

	t.Run("Sync Changes GET", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/changes?limit=10", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusOK)
		assertJSONContentType(t, res)
	})

	t.Run("Sync Changes Invalid Limit", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/v1/sync/changes?limit=bad", nil)
		res := httptest.NewRecorder()
		server.Handler().ServeHTTP(res, req)
		assertStatus(t, res, http.StatusOK)
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

func TestRequestLogger(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	
	logged := requestLogger(logger, handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	res := httptest.NewRecorder()
	
	logged.ServeHTTP(res, req)
	if res.Code != http.StatusOK {
		t.Error("request logger failed to pass through request")
	}
}

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	opRepo := &mockOpRepo{ops: make(map[string]*domain.Operation)}
	zettelRepo := &mockZettelSearchRepo{}
	topicRepo := &mockTopicRepo{}
	actorRepo := &mockActorRepo{}
	identityRepo := &mockIdentityRepo{}
	graphRepo := &mockGraphRepo{}
	
	syncSvc := services.NewSyncService(opRepo, zettelRepo, topicRepo)
	idSvc := services.NewIdentityService(actorRepo, identityRepo, graphRepo, opRepo)
	
	return NewServer(config.Config{HTTPAddr: ":0"}, logger, nil, nil, syncSvc, idSvc)
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
