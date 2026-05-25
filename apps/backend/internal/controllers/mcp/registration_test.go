package mcp

import (
	"log/slog"
	"io"
	"testing"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/config"
)

func TestToolRegistration(t *testing.T) {
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	server := mcpsdk.NewServer(&mcpsdk.Implementation{Name: "test", Version: "1.0"}, nil)
	
	t.Run("Register Ingest", func(t *testing.T) {
		registerIngestTool(server, logger, nil, "token")
	})

	t.Run("Register Search", func(t *testing.T) {
		registerSearchTool(server, logger, nil)
	})

	t.Run("Register Identity", func(t *testing.T) {
		registerIdentityTools(server, logger, nil, "token")
	})

	t.Run("Server Factory", func(t *testing.T) {
		factory := mcpServerFactory(config.Config{AgentToken: "t"}, logger, nil, nil, nil, nil)
		server := factory(nil)
		if server == nil {
			t.Error("expected non-nil server from factory")
		}
	})
}
