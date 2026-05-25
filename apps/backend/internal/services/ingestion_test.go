package services

import (
	"context"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

func TestIngestSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		service := NewIngestionService(
			&mockActorRepo{actors: make(map[string]*domain.Actor)},
			&mockSourceRepo{sources: make(map[string]*domain.Source)},
			&mockZettelRepo{zettels: make(map[string]*domain.Zettel)},
			&mockTopicRepo{topics: make(map[string]*domain.Topic)},
			&mockGraphRepo{nodes: make(map[string]string)},
			&mockOpRepo{ops: make(map[string]*domain.Operation)},
			&mockVectorRepo{},
			&mockEmbeddingsClient{},
		)
		payload := domain.SummaryPayload{
			Project:      "Test Project",
			Participants: []string{"Alice", "Bob"},
			Topics:       []string{"Go", "Architecture"},
			Timestamp:    time.Now(),
			Summary:      "We discussed the core architecture.",
			Conclusions:  []string{"Use Go for backend", "Use Turso for SQL"},
		}
		_, _, err := service.IngestSummary(ctx, payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("Graph Failures", func(t *testing.T) {
		graph := &mockGraphRepo{nodes: make(map[string]string)}
		service := NewIngestionService(
			&mockActorRepo{actors: make(map[string]*domain.Actor)},
			&mockSourceRepo{sources: make(map[string]*domain.Source)},
			&mockZettelRepo{zettels: make(map[string]*domain.Zettel)},
			&mockTopicRepo{topics: make(map[string]*domain.Topic)},
			graph,
			&mockOpRepo{ops: make(map[string]*domain.Operation)},
			&mockVectorRepo{},
			&mockEmbeddingsClient{},
		)

		labels := []string{"Person", "Topic", "Project", "Source", "Zettel"}
		for _, label := range labels {
			graph.failLabel = label
			payload := domain.SummaryPayload{Project: "P", Participants: []string{"A"}, Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}}
			_, _, err := service.IngestSummary(ctx, payload)
			if err == nil {
				t.Errorf("expected error for label %s", label)
			}
		}
	})

	t.Run("Repo Failures", func(t *testing.T) {
		actors := &mockActorRepo{actors: make(map[string]*domain.Actor)}
		topics := &mockTopicRepo{topics: make(map[string]*domain.Topic)}
		sources := &mockSourceRepo{sources: make(map[string]*domain.Source)}
		zettels := &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}

		service := NewIngestionService(actors, sources, zettels, topics, &mockGraphRepo{nodes: make(map[string]string)}, &mockOpRepo{ops: make(map[string]*domain.Operation)}, &mockVectorRepo{}, &mockEmbeddingsClient{})
		payload := domain.SummaryPayload{Project: "P", Participants: []string{"A"}, Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}}

		actors.findFail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on actor find")
		}
		actors.findFail = false

		actors.fail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on actor save")
		}
		actors.fail = false

		topics.findFail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on topic find")
		}
		topics.findFail = false

		topics.fail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on topic save")
		}
		topics.fail = false

		sources.fail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on source save")
		}
		sources.fail = false

		zettels.fail = true
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on zettel save")
		}
		zettels.fail = false
	})
}
