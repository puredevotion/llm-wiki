package mcp

import (
	"context"
	"log/slog"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

type SearchInput struct {
	Query string `json:"query" jsonschema:"natural language or keyword query"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum results to return"`
}

type SearchOutput struct {
	Results []SearchResult `json:"results" jsonschema:"ranked knowledge base results"`
}

type SearchResult struct {
	ID      string  `json:"id" jsonschema:"entity identifier"`
	Kind    string  `json:"kind" jsonschema:"source, zettel, topic, person, team, or event"`
	Title   string  `json:"title" jsonschema:"human-readable result title"`
	Snippet string  `json:"snippet" jsonschema:"short citation-friendly excerpt"`
	Score   float64 `json:"score" jsonschema:"relevance score from 0 to 1"`
}

// NewHandler exposes the remote MCP endpoint. Tool handlers should call domain services,
// not storage adapters; the current implementation is a compile-time scaffold.
func NewHandler(logger *slog.Logger) http.Handler {
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.search",
		Description: "Search the knowledge base across zettels, sources, topics, people, teams, and events.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchInput) (*mcpsdk.CallToolResult, SearchOutput, error) {
		logger.InfoContext(ctx, "mcp search requested", "query", input.Query, "limit", input.Limit)
		return &mcpsdk.CallToolResult{}, SearchOutput{Results: []SearchResult{}}, nil
	})

	return mcpsdk.NewStreamableHTTPHandler(func(_ *http.Request) *mcpsdk.Server {
		return server
	}, nil)
}
