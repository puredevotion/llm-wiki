package mcp

import (
	"log/slog"
	"net/http"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/services"
)

func NewHandler(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, idSvc *services.IdentityService, timeSvc *services.TimelineService) http.Handler {
	return mcpsdk.NewStreamableHTTPHandler(mcpServerFactory(cfg, logger, ingestion, searchSvc, idSvc, timeSvc), nil)
}

func mcpServerFactory(cfg config.Config, logger *slog.Logger, ingestion *services.IngestionService, searchSvc *services.SearchService, idSvc *services.IdentityService, timeSvc *services.TimelineService) func(r *http.Request) *mcpsdk.Server {
	return func(r *http.Request) *mcpsdk.Server {
		server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "llm-wiki", Version: "0.1.0"}, nil)
		registerSearchTool(server, logger, searchSvc)
		registerIngestTool(server, logger, ingestion, cfg.AgentToken)
		registerIdentityTools(server, logger, idSvc, cfg.AgentToken)
		registerTimelineTools(server, logger, timeSvc, cfg.AgentToken)
		return server
	}
}
