package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/services"
)

type SearchInput struct {
	Query string `json:"query" jsonschema:"the search query, conceptual or keyword"`
	Limit int    `json:"limit,omitempty" jsonschema:"maximum results, defaults to 10"`
}

type SearchResult struct {
	ID        string  `json:"id"`
	Kind      string  `json:"kind"`
	Title     string  `json:"title"`
	Snippet   string  `json:"snippet"`
	Lifecycle string  `json:"lifecycle"`
	Score     float64 `json:"score"`
}

type SearchOutput struct {
	Results []SearchResult `json:"results"`
}

func normalizeSearchInput(input SearchInput) (SearchInput, error) {
	input.Query = strings.TrimSpace(input.Query)
	if input.Query == "" {
		return input, fmt.Errorf("query is required")
	}
	if input.Limit < 0 {
		return input, fmt.Errorf("limit must be non-negative")
	}
	if input.Limit == 0 {
		input.Limit = 10
	}
	if input.Limit > 50 {
		return input, fmt.Errorf("limit exceeds maximum of 50")
	}
	return input, nil
}

func registerSearchTool(server *mcpsdk.Server, logger *slog.Logger, searchSvc *services.SearchService) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.search",
		Description: "Conceptual and keyword search over the knowledge base.",
	}, searchToolHandler(logger, searchSvc))
}

func searchToolHandler(logger *slog.Logger, searchSvc *services.SearchService) func(context.Context, *mcpsdk.CallToolRequest, SearchInput) (*mcpsdk.CallToolResult, SearchOutput, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input SearchInput) (*mcpsdk.CallToolResult, SearchOutput, error) {
		normalized, err := normalizeSearchInput(input)
		if err != nil {
			return nil, SearchOutput{}, err
		}
		logger.InfoContext(ctx, "mcp search requested", "query", normalized.Query, "limit", normalized.Limit)

		results, err := searchSvc.Search(ctx, normalized.Query, normalized.Limit)
		if err != nil {
			return nil, SearchOutput{}, err
		}

		output := SearchOutput{
			Results: make([]SearchResult, 0, len(results)),
		}
		for _, r := range results {
			output.Results = append(output.Results, SearchResult{
				ID:        r.ID,
				Kind:      "zettel",
				Title:     r.Title,
				Snippet:   r.Snippet,
				Lifecycle: r.Lifecycle,
				Score:     r.Score,
			})
		}

		return &mcpsdk.CallToolResult{}, output, nil
	}
}
