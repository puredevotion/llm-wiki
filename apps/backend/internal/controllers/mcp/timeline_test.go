package mcp

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

func TestTimelineToolHandler(t *testing.T) {
	agentToken := "test-token"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	events := &mockTimelineRepo{events: make(map[string]*domain.Event)}
	graph := &mockGraphRepo{}
	ops := &mockOpRepo{}
	
	timeSvc := services.NewTimelineService(events, graph, ops)
	handler := timelineToolHandler(logger, timeSvc, agentToken)

	t.Run("Record Event", func(t *testing.T) {
		input := TimelineInput{
			Token:  "test-token",
			Action: "record_event",
			Event: &domain.Event{
				Kind:  domain.EventMilestone,
				Title: "Beta Launch",
			},
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "event_recorded" {
			t.Errorf("expected event_recorded, got %s", output.Status)
		}
	})

	t.Run("Get Timeline", func(t *testing.T) {
		input := TimelineInput{
			Token:  "test-token",
			Action: "get_timeline",
			Limit:  10,
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "success" {
			t.Errorf("expected success, got %s", output.Status)
		}
	})

	t.Run("Get Timeline with Dates", func(t *testing.T) {
		input := TimelineInput{
			Token:    "test-token",
			Action:   "get_timeline",
			StartsAt: "2026-05-01T00:00:00Z",
			EndsAt:   "2026-05-31T23:59:59Z",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "success" {
			t.Errorf("expected success, got %s", output.Status)
		}
	})

	t.Run("Get Timeline with Invalid Dates", func(t *testing.T) {
		input := TimelineInput{
			Token:    "test-token",
			Action:   "get_timeline",
			StartsAt: "invalid",
			EndsAt:   "invalid",
		}
		handler(context.Background(), nil, input)
	})

	t.Run("Relate Event", func(t *testing.T) {
		input := TimelineInput{
			Token:      "test-token",
			Action:     "relate_event",
			EventID:    "e1",
			TargetID:   "p1",
			TargetKind: "Project",
			RelType:    "HAPPENED_DURING",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "relationship_created" {
			t.Errorf("expected relationship_created, got %s", output.Status)
		}
	})

	t.Run("Record Event Fail", func(t *testing.T) {
		events.fail = true
		defer func() { events.fail = false }()
		input := TimelineInput{Token: "test-token", Action: "record_event", Event: &domain.Event{}}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Get Timeline Fail", func(t *testing.T) {
		events.fail = true
		defer func() { events.fail = false }()
		input := TimelineInput{Token: "test-token", Action: "get_timeline"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Relate Event Fail", func(t *testing.T) {
		graph.failRel = true
		defer func() { graph.failRel = false }()
		input := TimelineInput{Token: "test-token", Action: "relate_event", EventID: "e1", TargetID: "p1"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		input := TimelineInput{Token: "bad"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected unauthorized error")
		}
	})

	t.Run("Missing Event Data", func(t *testing.T) {
		input := TimelineInput{Token: "test-token", Action: "record_event"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error for missing event data")
		}
	})

	t.Run("Unsupported Action", func(t *testing.T) {
		input := TimelineInput{Token: "test-token", Action: "bad"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected unsupported action error")
		}
	})
}
