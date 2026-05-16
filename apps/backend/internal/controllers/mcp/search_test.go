package mcp

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewHandlerDocumentsTheRemoteMCPInitializeEndpoint(t *testing.T) {
	handler := NewHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))
	req := httptest.NewRequest(http.MethodPost, "/mcp", bytes.NewBufferString(`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	res := httptest.NewRecorder()

	handler.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected MCP initialize to return 200, got %d with body %q", res.Code, res.Body.String())
	}
	if contentType := res.Header().Get("Content-Type"); contentType == "" {
		t.Fatal("expected MCP initialize response to document its transport content type")
	}
}

func TestNormalizeSearchInputRejectsEmptyQueriesSoAgentsCannotTriggerUnboundedSearch(t *testing.T) {
	_, err := normalizeSearchInput(SearchInput{Query: "   ", Limit: 10})

	if err == nil {
		t.Fatal("expected blank query to be rejected before it reaches a service")
	}
}

func TestNormalizeSearchInputDefaultsLimitSoAgentSearchHasPredictablePagination(t *testing.T) {
	input, err := normalizeSearchInput(SearchInput{Query: "project memory"})
	if err != nil {
		t.Fatalf("expected valid query to normalize successfully: %v", err)
	}

	if input.Limit != defaultSearchLimit {
		t.Fatalf("expected default limit %d, got %d", defaultSearchLimit, input.Limit)
	}
}

func TestNormalizeSearchInputRejectsExcessiveLimitsToProtectTheMCPServer(t *testing.T) {
	_, err := normalizeSearchInput(SearchInput{Query: "project memory", Limit: maxSearchLimit + 1})

	if err == nil {
		t.Fatalf("expected limits above %d to be rejected", maxSearchLimit)
	}
}

func TestNormalizeSearchInputTrimsWhitespaceBeforeCallingSearchServices(t *testing.T) {
	input, err := normalizeSearchInput(SearchInput{Query: "  project memory  ", Limit: 5})
	if err != nil {
		t.Fatalf("expected valid query to normalize successfully: %v", err)
	}

	if input.Query != "project memory" {
		t.Fatalf("expected query whitespace to be trimmed, got %q", input.Query)
	}
}

func TestSearchToolHandlerDocumentsTheEmptyResultShapeBeforeSearchServicesExist(t *testing.T) {
	handler := searchToolHandler(slog.New(slog.NewTextHandler(io.Discard, nil)))

	result, output, err := handler(context.Background(), nil, SearchInput{Query: "project memory", Limit: 5})

	if err != nil {
		t.Fatalf("expected valid search tool input to succeed: %v", err)
	}
	if result == nil {
		t.Fatal("expected MCP search tool to return a call result object")
	}
	if output.Results == nil {
		t.Fatal("expected MCP search tool to document results as an empty list instead of null")
	}
	if len(output.Results) != 0 {
		t.Fatalf("expected scaffold search output to be empty, got %d results", len(output.Results))
	}
}
