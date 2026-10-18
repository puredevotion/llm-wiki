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
	ID        string  `json:"id" jsonschema:"entity identifier"`
	Kind      string  `json:"kind" jsonschema:"source, zettel, topic, person, team, or event"`
	Title     string  `json:"title" jsonschema:"human-readable result title"`
	Snippet   string  `json:"snippet" jsonschema:"short citation-friendly excerpt"`
	Lifecycle string  `json:"lifecycle,omitempty" jsonschema:"project, evergreen, or ephemeral"`
	Score     float64 `json:"score" jsonschema:"relevance score from 0 to 1"`
}

func NewHandler(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, idSvc *services.IdentityService) http.Handler {
	return mcpsdk.NewStreamableHTTPHandler(mcpServerFactory(cfg, logger, ingestion, searchSvc, idSvc), nil)
}

func mcpServerFactory(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, idSvc *services.IdentityService) func(r *http.Request) *mcpsdk.Server {
	return func(r *http.Request) *mcpsdk.Server {
		server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)
		registerSearchTool(server, logger, searchSvc)
		registerIngestTool(server, logger, ingestion, cfg.AgentToken)
		registerIdentityTools(server, logger, idSvc, cfg.AgentToken)
		return server
	}
}
