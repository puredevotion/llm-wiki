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

type TimelineInput struct {
	Token      string           `json:"token" jsonschema:"agent api token"`
	Action     string           `json:"action" jsonschema:"record_event, get_timeline, relate_event"`
	Event      *domain.Event    `json:"event,omitempty" jsonschema:"event data for record_event"`
	StartsAt   string           `json:"starts_at,omitempty" jsonschema:"RFC3339 start of range for get_timeline"`
	EndsAt     string           `json:"ends_at,omitempty" jsonschema:"RFC3339 end of range for get_timeline"`
	Limit      int              `json:"limit,omitempty" jsonschema:"max events to return"`
	EventID    string           `json:"event_id,omitempty" jsonschema:"ID of the event to relate"`
	TargetID   string           `json:"target_id,omitempty" jsonschema:"ID of the target entity to relate to"`
	TargetKind string           `json:"target_kind,omitempty" jsonschema:"kind of the target entity"`
	RelType    string           `json:"rel_type,omitempty" jsonschema:"HAPPENED_DURING, INVOLVES, PRECEDES, FOLLOWS, MENTIONED_IN"`
}

type TimelineOutput struct {
	Events []*domain.Event `json:"events,omitempty"`
	Status string          `json:"status"`
}

func registerTimelineTools(server *mcpsdk.Server, logger *slog.Logger, timeSvc *services.TimelineService, agentToken string) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.manage_timeline",
		Description: "Manage and query the temporal history of events, meetings, and decisions.",
	}, timelineToolHandler(logger, timeSvc, agentToken))
}

func timelineToolHandler(logger *slog.Logger, timeSvc *services.TimelineService, agentToken string) func(context.Context, *mcpsdk.CallToolRequest, TimelineInput) (*mcpsdk.CallToolResult, TimelineOutput, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input TimelineInput) (*mcpsdk.CallToolResult, TimelineOutput, error) {
		if input.Token != agentToken {
			return nil, TimelineOutput{}, fmt.Errorf("unauthorized: invalid agent token")
		}

		switch input.Action {
		case "record_event":
			if input.Event == nil {
				return nil, TimelineOutput{}, fmt.Errorf("missing event data")
			}
			if err := timeSvc.RecordEvent(ctx, input.Event); err != nil {
				return nil, TimelineOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, TimelineOutput{Status: "event_recorded"}, nil

		case "get_timeline":
			var sAt, eAt *time.Time
			if input.StartsAt != "" {
				t, _ := time.Parse(time.RFC3339, input.StartsAt)
				sAt = &t
			}
			if input.EndsAt != "" {
				t, _ := time.Parse(time.RFC3339, input.EndsAt)
				eAt = &t
			}
			events, err := timeSvc.FetchTimeline(ctx, sAt, eAt, input.Limit)
			if err != nil {
				return nil, TimelineOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, TimelineOutput{Events: events, Status: "success"}, nil

		case "relate_event":
			if err := timeSvc.RelateEvent(ctx, input.EventID, input.TargetID, input.TargetKind, input.RelType); err != nil {
				return nil, TimelineOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, TimelineOutput{Status: "relationship_created"}, nil

		default:
			return nil, TimelineOutput{}, fmt.Errorf("unsupported action: %s", input.Action)
		}
	}
}
