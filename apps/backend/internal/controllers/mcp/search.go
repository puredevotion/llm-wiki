package mcp

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/services"
)

const (
	defaultSearchLimit = 10
	maxSearchLimit     = 50
)

func normalizeSearchInput(input SearchInput) (SearchInput, error) {
	query := strings.TrimSpace(input.Query)
	if query == "" {
		return SearchInput{}, errors.New("query is required")
	}

	limit := input.Limit
	if limit == 0 {
		limit = defaultSearchLimit
	}
	if limit < 0 {
		return SearchInput{}, errors.New("limit must be positive")
	}
	if limit > maxSearchLimit {
		return SearchInput{}, fmt.Errorf("limit must be less than or equal to %d", maxSearchLimit)
	}

	return SearchInput{Query: query, Limit: limit}, nil
}

func registerSearchTool(server *mcpsdk.Server, logger *slog.Logger, searchSvc *services.SearchService) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.search",
		Description: "Search the knowledge base across zettels, sources, topics, people, teams, and events.",
	}, searchToolHandler(logger, searchSvc))
}

func searchToolHandler(logger *slog.Logger, searchSvc *services.SearchService) func(context.Context, *mcpsdk.CallToolRequest, SearchInput) (*mcpsdk.CallToolResult, SearchOutput, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchInput) (*mcpsdk.CallToolResult, SearchOutput, error) {
		normalized, err := normalizeSearchInput(input)
		if err != nil {
			return nil, SearchOutput{}, err
		}
		logger.InfoContext(ctx, "mcp search requested", "query", normalized.Query, "limit", normalized.Limit)

		zettels, err := searchSvc.Search(ctx, normalized.Query, normalized.Limit)
		if err != nil {
			return nil, SearchOutput{}, err
		}

		results := make([]SearchResult, 0, len(zettels))
		for _, z := range zettels {
			results = append(results, SearchResult{
				ID:        z.ID,
				Kind:      "zettel",
				Title:     z.Title,
				Lifecycle: z.Lifecycle,
			})
		}

		return &mcpsdk.CallToolResult{}, SearchOutput{Results: results}, nil
	}
}
