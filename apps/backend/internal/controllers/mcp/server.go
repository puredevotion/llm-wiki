package mcp

import (
	"log/slog"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/services"
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
func NewHandler(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService) http.Handler {
	return mcpsdk.NewStreamableHTTPHandler(func(r *http.Request) *mcpsdk.Server {
		server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)
		registerSearchTool(server, logger)
		registerIngestTool(server, logger, ingestion, cfg.AgentToken)
		return server
	}, nil)
}
