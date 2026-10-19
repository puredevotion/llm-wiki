package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type TimelineService struct {
	events repositories.TimelineRepository
	graph  repositories.GraphRepository
	ops    repositories.OperationRepository
}

func NewTimelineService(
	events repositories.TimelineRepository,
	graph repositories.GraphRepository,
	ops repositories.OperationRepository,
) *TimelineService {
	return &TimelineService{
		events: events,
		graph:  graph,
		ops:    ops,
	}
}

func (s *TimelineService) RecordEvent(ctx context.Context, e *domain.Event) error {
	if e.ID == "" {
		e.ID = fmt.Sprintf("evt_%d", time.Now().UnixNano())
	}
	if e.RecordedAt.IsZero() {
		e.RecordedAt = time.Now()
	}

	// 1. SQL Save
	if err := s.events.Save(ctx, e); err != nil {
		return fmt.Errorf("failed to save event to sql: %w", err)
	}

	// 2. Graph Save
	props := map[string]any{
		"title": e.Title,
		"kind":  string(e.Kind),
	}
	if err := s.graph.UpsertNode(ctx, e.ID, "Event", props); err != nil {
		return fmt.Errorf("failed to upsert event node: %w", err)
	}

	// 3. Sync Log
	pJSON, _ := json.Marshal(e)
	op := &domain.Operation{
		ID:            fmt.Sprintf("op_evt_%d", time.Now().UnixNano()),
		EntityKind:    "event",
		EntityID:      e.ID,
		OperationType: "upsert",
		Payload:       pJSON,
		Status:        domain.OperationApplied,
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	op.AppliedAt = &now
	return s.ops.Save(ctx, op)
}

func (s *TimelineService) FetchTimeline(ctx context.Context, startsAt, endsAt *time.Time, limit int) ([]*domain.Event, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.events.Fetch(ctx, startsAt, endsAt, limit)
}

func (s *TimelineService) GetEvent(ctx context.Context, id string) (*domain.Event, error) {
	return s.events.FindByID(ctx, id)
}

func (s *TimelineService) RelateEvent(ctx context.Context, eventID, targetID, targetLabel, relType string) error {
	// Identity relationship in Graph
	// Valid rel types: HAPPENED_DURING, INVOLVES, PRECEDES, FOLLOWS, MENTIONED_IN
	if err := s.graph.CreateRelationship(ctx, eventID, "Event", targetID, targetLabel, relType); err != nil {
		return fmt.Errorf("failed to relate event: %w", err)
	}

	// Sync Log
	return s.logRelOperation(ctx, eventID, targetID, targetLabel, relType)
}

func (s *TimelineService) logRelOperation(ctx context.Context, eventID, targetID, targetLabel, relType string) error {
	payload := map[string]string{
		"event_id":     eventID,
		"target_id":    targetID,
		"target_label": targetLabel,
		"rel_type":     relType,
	}
	pJSON, _ := json.Marshal(payload)
	op := &domain.Operation{
		ID:            fmt.Sprintf("op_evr_%d", time.Now().UnixNano()),
		EntityKind:    "event_relationship",
		EntityID:      fmt.Sprintf("%s:%s", eventID, targetID),
		OperationType: "create",
		Payload:       pJSON,
		Status:        domain.OperationApplied,
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	op.AppliedAt = &now
	return s.ops.Save(ctx, op)
}
