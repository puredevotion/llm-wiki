package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type IngestionService struct {
	actors  repositories.ActorRepository
	sources repositories.SourceRepository
	zettels repositories.ZettelRepository
	topics  repositories.TopicRepository
	graph   repositories.GraphRepository
	ops     repositories.OperationRepository
}

func NewIngestionService(
	actors repositories.ActorRepository,
	sources repositories.SourceRepository,
	zettels repositories.ZettelRepository,
	topics repositories.TopicRepository,
	graph repositories.GraphRepository,
	ops repositories.OperationRepository,
) *IngestionService {
	return &IngestionService{
		actors:  actors,
		sources: sources,
		zettels: zettels,
		topics:  topics,
		graph:   graph,
		ops:     ops,
	}
}

// IngestSummary takes a conversation summary and persists it as a Source and a Zettel.
// It also resolves or creates Actors for the participants, Topics, and Graph relationships.
func (s *IngestionService) IngestSummary(ctx context.Context, payload domain.SummaryPayload) (sourceID string, zettelID string, err error) {
	if payload.Summary == "" {
		return "", "", fmt.Errorf("summary body is required")
	}
	if len(payload.Conclusions) == 0 {
		return "", "", fmt.Errorf("at least one conclusion is required")
	}

	// 1. Resolve participants to Actors
	actorIDs := make([]string, 0, len(payload.Participants))
	for _, p := range payload.Participants {
		actor, err := s.actors.FindByName(ctx, p)
		if err != nil {
			return "", "", fmt.Errorf("failed to find actor %s: %w", p, err)
		}
		if actor == nil {
			actor = &domain.Actor{
				ID:          fmt.Sprintf("actor_%d", time.Now().UnixNano()),
				Kind:        "person",
				DisplayName: p,
				CreatedAt:   time.Now(),
			}
			if err := s.actors.Save(ctx, actor); err != nil {
				return "", "", fmt.Errorf("failed to save actor %s: %w", p, err)
			}
		}
		actorIDs = append(actorIDs, actor.ID)
		// Ensure actor node in graph
		if err := s.graph.UpsertNode(ctx, actor.ID, "Person", map[string]any{"name": actor.DisplayName}); err != nil {
			return "", "", fmt.Errorf("failed to upsert actor node %s: %w", actor.ID, err)
		}
	}

	// 2. Resolve Topics
	topicIDs := make([]string, 0, len(payload.Topics))
	for _, t := range payload.Topics {
		topic, err := s.topics.FindByName(ctx, t)
		if err != nil {
			return "", "", fmt.Errorf("failed to find topic %s: %w", t, err)
		}
		if topic == nil {
			topic = &domain.Topic{
				ID:        fmt.Sprintf("top_%d", time.Now().UnixNano()),
				Name:      t,
				CreatedAt: time.Now(),
			}
			if err := s.topics.Save(ctx, topic); err != nil {
				return "", "", fmt.Errorf("failed to save topic %s: %w", t, err)
			}

			// Record Sync Operation for Topic
			tPayload, _ := json.Marshal(topic)
			tOp := &domain.Operation{
				ID:            fmt.Sprintf("op_top_%d", time.Now().UnixNano()),
				EntityKind:    "topic",
				EntityID:      topic.ID,
				OperationType: "upsert",
				Payload:       tPayload,
				Status:        domain.OperationApplied,
				CreatedAt:     time.Now(),
			}
			now := time.Now()
			tOp.AppliedAt = &now
			_ = s.ops.Save(ctx, tOp)
		}
		topicIDs = append(topicIDs, topic.ID)
		// Ensure topic node in graph
		if err := s.graph.UpsertNode(ctx, topic.ID, "Topic", map[string]any{"name": topic.Name}); err != nil {
			return "", "", fmt.Errorf("failed to upsert topic node %s: %w", topic.ID, err)
		}
	}

	// 3. Resolve Project Node
	projectID := fmt.Sprintf("prj_%s", payload.Project) // Stable-ish ID for project
	if err := s.graph.UpsertNode(ctx, projectID, "Project", map[string]any{"name": payload.Project}); err != nil {
		return "", "", fmt.Errorf("failed to upsert project node: %w", err)
	}

	// 4. Create Source
	source := &domain.Source{
		ID:          fmt.Sprintf("src_%d", time.Now().UnixNano()),
		Kind:        "conversation",
		Title:       fmt.Sprintf("Conversation Summary: %s", payload.Project),
		CapturedAt:  time.Now(),
		Metadata: map[string]any{
			"project":      payload.Project,
			"participants": payload.Participants,
			"topics":       payload.Topics,
			"timestamp":    payload.Timestamp,
		},
	}
	if err := s.sources.Save(ctx, source); err != nil {
		return "", "", fmt.Errorf("failed to save source: %w", err)
	}
	// Source node in graph
	if err := s.graph.UpsertNode(ctx, source.ID, "Source", map[string]any{"title": source.Title, "kind": source.Kind}); err != nil {
		return "", "", fmt.Errorf("failed to upsert source node: %w", err)
	}

	// 5. Create Zettel
	body := fmt.Sprintf("%s\n\n### Conclusions\n- %s", payload.Summary, strings.Join(payload.Conclusions, "\n- "))
	zettel := &domain.Zettel{
		ID:        fmt.Sprintf("zet_%d", time.Now().UnixNano()),
		Title:     fmt.Sprintf("Summary: %s", payload.Project),
		Body:      body,
		Lifecycle: "project",
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := s.zettels.Save(ctx, zettel); err != nil {
		return "", "", fmt.Errorf("failed to save zettel: %w", err)
	}

	// 5.1 Record Sync Operation for Zettel
	zPayload, _ := json.Marshal(zettel)
	zOp := &domain.Operation{
		ID:            fmt.Sprintf("op_zet_%d", time.Now().UnixNano()),
		EntityKind:    "zettel",
		EntityID:      zettel.ID,
		OperationType: "upsert",
		Payload:       zPayload,
		Status:        domain.OperationApplied,
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	zOp.AppliedAt = &now
	_ = s.ops.Save(ctx, zOp)

	// Zettel node in graph
	if err := s.graph.UpsertNode(ctx, zettel.ID, "Zettel", map[string]any{"title": zettel.Title, "lifecycle": zettel.Lifecycle}); err != nil {
		return "", "", fmt.Errorf("failed to upsert zettel node: %w", err)
	}

	// 6. Create Graph Edges
	// Source -> Actors (AUTHORED_BY)
	for _, aid := range actorIDs {
		if err := s.graph.CreateRelationship(ctx, source.ID, "Source", aid, "Person", "AUTHORED_BY"); err != nil {
			return "", "", fmt.Errorf("failed to link source to actor: %w", err)
		}
	}
	// Source -> Topics (RELATED_TO)
	for _, tid := range topicIDs {
		if err := s.graph.CreateRelationship(ctx, source.ID, "Source", tid, "Topic", "RELATED_TO"); err != nil {
			return "", "", fmt.Errorf("failed to link source to topic: %w", err)
		}
	}
	// Source -> Project (RELATED_TO)
	if err := s.graph.CreateRelationship(ctx, source.ID, "Source", projectID, "Project", "RELATED_TO"); err != nil {
		return "", "", fmt.Errorf("failed to link source to project: %w", err)
	}
	// Zettel -> Source (DERIVED_FROM)
	if err := s.graph.CreateRelationship(ctx, zettel.ID, "Zettel", source.ID, "Source", "DERIVED_FROM"); err != nil {
		return "", "", fmt.Errorf("failed to link zettel to source: %w", err)
	}
	// Zettel -> Topics (RELATED_TO)
	for _, tid := range topicIDs {
		if err := s.graph.CreateRelationship(ctx, zettel.ID, "Zettel", tid, "Topic", "RELATED_TO"); err != nil {
			return "", "", fmt.Errorf("failed to link zettel to topic: %w", err)
		}
	}

	return source.ID, zettel.ID, nil
}
