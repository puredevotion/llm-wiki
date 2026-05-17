package rest

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"llm-wiki/apps/backend/internal/config"
)

func TestHealthEndpointDocumentsTheStandardSuccessEnvelope(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	server.Handler().ServeHTTP(res, req)

	assertStatus(t, res, http.StatusOK)
	assertJSONContentType(t, res)
	body := decodeJSON(t, res)
	data := body["data"].(map[string]any)
	if data["status"] != "ok" {
		t.Fatalf("expected health status to document service readiness, got %#v", data["status"])
	}
}

func TestUnknownVersionedAPIRouteDocumentsTheStandardNotFoundError(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/does-not-exist", nil)
	res := httptest.NewRecorder()

	server.Handler().ServeHTTP(res, req)

	assertStatus(t, res, http.StatusNotFound)
	assertJSONContentType(t, res)
	errorBody := decodeError(t, res)
	if errorBody.Code != "not_found" {
		t.Fatalf("expected stable not_found error code, got %q", errorBody.Code)
	}
	if errorBody.Message == "" {
		t.Fatal("expected not_found error to include a user-safe message")
	}
}

func TestUnsupportedMethodDocumentsTheStandardMethodNotAllowedError(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(http.MethodPost, "/healthz", nil)
	res := httptest.NewRecorder()

	server.Handler().ServeHTTP(res, req)

	assertStatus(t, res, http.StatusMethodNotAllowed)
	assertJSONContentType(t, res)
	if allow := res.Header().Get("Allow"); allow != http.MethodGet {
		t.Fatalf("expected Allow header to document supported method, got %q", allow)
	}
	errorBody := decodeError(t, res)
	if errorBody.Code != "method_not_allowed" {
		t.Fatalf("expected stable method_not_allowed error code, got %q", errorBody.Code)
	}
}

func TestVersionedSyncEndpointsAreReservedWithStandardNotImplementedErrors(t *testing.T) {
	server := newTestServer()
	tests := []struct {
		name   string
		method string
		path   string
		body   io.Reader
	}{
		{
			name:   "offline clients push operation batches through the versioned operations collection",
			method: http.MethodPost,
			path:   "/api/v1/sync/operations",
			body:   strings.NewReader(`{"operations":[]}`),
		},
		{
			name:   "offline clients pull ordered changes with a cursor and bounded limit",
			method: http.MethodGet,
			path:   "/api/v1/sync/changes?cursor=start&limit=50",
		},
		{
			name:   "offline clients bootstrap from a compact versioned snapshot endpoint",
			method: http.MethodGet,
			path:   "/api/v1/sync/bootstrap",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, tt.body)
			res := httptest.NewRecorder()

			server.Handler().ServeHTTP(res, req)

			assertStatus(t, res, http.StatusNotImplemented)
			assertJSONContentType(t, res)
			errorBody := decodeError(t, res)
			if errorBody.Code != "not_implemented" {
				t.Fatalf("expected sync placeholder to use not_implemented code, got %q", errorBody.Code)
			}
		})
	}
}

func TestVersionedSyncEndpointsDocumentAllowedMethodsForOfflineClients(t *testing.T) {
	server := newTestServer()
	tests := []struct {
		name          string
		method        string
		path          string
		allowedMethod string
	}{
		{
			name:          "operation batches are only accepted with POST",
			method:        http.MethodGet,
			path:          "/api/v1/sync/operations",
			allowedMethod: http.MethodPost,
		},
		{
			name:          "change pulls are read-only GET requests",
			method:        http.MethodPost,
			path:          "/api/v1/sync/changes",
			allowedMethod: http.MethodGet,
		},
		{
			name:          "bootstrap snapshots are read-only GET requests",
			method:        http.MethodPost,
			path:          "/api/v1/sync/bootstrap",
			allowedMethod: http.MethodGet,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			res := httptest.NewRecorder()

			server.Handler().ServeHTTP(res, req)

			assertStatus(t, res, http.StatusMethodNotAllowed)
			if allow := res.Header().Get("Allow"); allow != tt.allowedMethod {
				t.Fatalf("expected Allow header %q, got %q", tt.allowedMethod, allow)
			}
			errorBody := decodeError(t, res)
			if errorBody.Code != "method_not_allowed" {
				t.Fatalf("expected stable method_not_allowed error code, got %q", errorBody.Code)
			}
		})
	}
}

func TestUnversionedSyncEndpointsStayUnavailableToProtectMobileContractMigrations(t *testing.T) {
	server := newTestServer()
	req := httptest.NewRequest(http.MethodGet, "/api/sync/pull", nil)
	res := httptest.NewRecorder()

	server.Handler().ServeHTTP(res, req)

	assertStatus(t, res, http.StatusNotFound)
	errorBody := decodeError(t, res)
	if errorBody.Code != "not_found" {
		t.Fatalf("expected unversioned sync endpoint to stay unavailable, got %q", errorBody.Code)
	}
}

func newTestServer() *Server {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(config.Config{HTTPAddr: ":0"}, logger, nil)
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

func decodeJSON(t *testing.T, res *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var body map[string]any
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON body, got decode error: %v", err)
	}
	return body
}

type apiErrorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func decodeError(t *testing.T, res *httptest.ResponseRecorder) struct {
	Code    string
	Message string
} {
	t.Helper()
	var body apiErrorEnvelope
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("expected JSON error body, got decode error: %v", err)
	}
	return struct {
		Code    string
		Message string
	}{Code: body.Error.Code, Message: body.Error.Message}
}
