package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type IngestInput struct {
	Token        string   `json:"token" jsonschema:"agent api token for authentication"`
	Project      string   `json:"project" jsonschema:"name of the project this conversation belongs to"`
	Participants []string `json:"participants" jsonschema:"list of people involved in the conversation"`
	Topics       []string `json:"topics" jsonschema:"list of topics covered in the conversation"`
	Timestamp    string   `json:"timestamp" jsonschema:"RFC3339 timestamp of the conversation"`
	Summary      string   `json:"summary" jsonschema:"clear summary of the discussion"`
	Conclusions  []string `json:"conclusions" jsonschema:"list of key decisions or takeaways"`
}

type IngestOutput struct {
	SourceID string `json:"source_id"`
	ZettelID string `json:"zettel_id"`
	Status   string `json:"status"`
}

type GetTokenOutput struct {
	Token string `json:"token"`
}

func registerIngestTool(server *mcpsdk.Server, logger *slog.Logger, ingestion *services.IngestionService, agentToken string) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.ingest_summary",
		Description: "Ingest a summarized conversation into the knowledge base. Requires a valid agent token.",
	}, ingestToolHandler(logger, ingestion, agentToken))

	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.get_token",
		Description: "Retrieve the agent API token for the session.",
	}, func(ctx context.Context, req *mcpsdk.CallToolRequest, input struct{}) (*mcpsdk.CallToolResult, GetTokenOutput, error) {
		return &mcpsdk.CallToolResult{}, GetTokenOutput{Token: agentToken}, nil
	})
}

func ingestToolHandler(logger *slog.Logger, ingestion *services.IngestionService, agentToken string) func(context.Context, *mcpsdk.CallToolRequest, IngestInput) (*mcpsdk.CallToolResult, IngestOutput, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input IngestInput) (*mcpsdk.CallToolResult, IngestOutput, error) {
		// 1. Authenticate
		if input.Token != agentToken {
			logger.WarnContext(ctx, "unauthorized ingest attempt", "provided_token", input.Token)
			return nil, IngestOutput{}, fmt.Errorf("unauthorized: invalid agent token")
		}

		// 2. Parse Timestamp (Journey 1, Case 1)
		ts, err := time.Parse(time.RFC3339, input.Timestamp)
		if err != nil {
			// Fallback to now if invalid or missing
			logger.WarnContext(ctx, "invalid timestamp provided, falling back to current time", "error", err, "input_timestamp", input.Timestamp)
			ts = time.Now()
		}

		// 3. Call Ingestion Service
		payload := domain.SummaryPayload{
			Project:      input.Project,
			Participants: input.Participants,
			Topics:       input.Topics,
			Timestamp:    ts,
			Summary:      input.Summary,
			Conclusions:  input.Conclusions,
		}

		srcID, zetID, err := ingestion.IngestSummary(ctx, payload)
		if err != nil {
			logger.ErrorContext(ctx, "ingestion failed", "error", err)
			return nil, IngestOutput{}, err
		}

		logger.InfoContext(ctx, "ingestion successful", "source_id", srcID, "zettel_id", zetID)
		return &mcpsdk.CallToolResult{}, IngestOutput{
			SourceID: srcID,
			ZettelID: zetID,
			Status:   "created",
		}, nil
	}
}
