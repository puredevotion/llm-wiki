package services

import (
	"context"
	"fmt"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

type mockActorRepo struct {
	actors   map[string]*domain.Actor
	fail     bool
	findFail bool
}

func (m *mockActorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	if m.findFail {
		return nil, fmt.Errorf("find error")
	}
	for _, a := range m.actors {
		if a.DisplayName == name {
			return a, nil
		}
	}
	return nil, nil
}

func (m *mockActorRepo) Save(ctx context.Context, actor *domain.Actor) error {
	if actor.DisplayName == "ErrorActor" || m.fail {
		return fmt.Errorf("database error")
	}
	m.actors[actor.ID] = actor
	return nil
}

type mockSourceRepo struct {
	sources map[string]*domain.Source
	fail    bool
}

func (m *mockSourceRepo) Save(ctx context.Context, source *domain.Source) error {
	if m.fail {
		return fmt.Errorf("source save failed")
	}
	m.sources[source.ID] = source
	return nil
}

type mockZettelRepo struct {
	zettels map[string]*domain.Zettel
	fail    bool
}

func (m *mockZettelRepo) Save(ctx context.Context, zettel *domain.Zettel) error {
	if m.fail {
		return fmt.Errorf("zettel save failed")
	}
	m.zettels[zettel.ID] = zettel
	return nil
}

func (m *mockZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	return []*domain.Zettel{}, nil
}

type mockTopicRepo struct {
	topics   map[string]*domain.Topic
	fail     bool
	findFail bool
}

func (m *mockTopicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	if m.findFail {
		return nil, fmt.Errorf("find error")
	}
	for _, t := range m.topics {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, nil
}

func (m *mockTopicRepo) Save(ctx context.Context, topic *domain.Topic) error {
	if m.fail {
		return fmt.Errorf("topic save failed")
	}
	m.topics[topic.ID] = topic
	return nil
}

type mockGraphRepo struct {
	nodes     map[string]string // id -> label
	edges     []string          // from:fromLabel:to:toLabel:type
	fail      bool
	failRel   bool
	failLabel string
	failTo    string
}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	if m.fail || m.failLabel == label {
		return fmt.Errorf("graph upsert failed")
	}
	m.nodes[id] = label
	return nil
}

func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	if m.fail || m.failRel || (m.failTo != "" && m.failTo == toLabel) {
		return fmt.Errorf("graph relationship failed")
	}
	m.edges = append(m.edges, fmt.Sprintf("%s:%s:%s:%s:%s", fromID, fromLabel, toID, toLabel, relType))
	return nil
}

func TestIngestSummary(t *testing.T) {
	actors := &mockActorRepo{actors: make(map[string]*domain.Actor)}
	sources := &mockSourceRepo{sources: make(map[string]*domain.Source)}
	zettels := &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}
	topics := &mockTopicRepo{topics: make(map[string]*domain.Topic)}
	graph := &mockGraphRepo{nodes: make(map[string]string)}

	service := NewIngestionService(actors, sources, zettels, topics, graph)
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
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

	t.Run("Validation Errors", func(t *testing.T) {
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: ""})
		if err == nil {
			t.Error("expected error")
		}
		_, _, err = service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Actor Find Fail", func(t *testing.T) {
		actors.findFail = true
		defer func() { actors.findFail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Participants: []string{"A"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Actor Save Fail", func(t *testing.T) {
		actors.fail = true
		defer func() { actors.fail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Participants: []string{"A"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Actor Graph Upsert Fail", func(t *testing.T) {
		graph.failLabel = "Person"
		defer func() { graph.failLabel = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Participants: []string{"A"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Topic Find Fail", func(t *testing.T) {
		topics.findFail = true
		defer func() { topics.findFail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Topic Save Fail", func(t *testing.T) {
		topics.fail = true
		defer func() { topics.fail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Topic Graph Upsert Fail", func(t *testing.T) {
		graph.failLabel = "Topic"
		defer func() { graph.failLabel = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Project Graph Upsert Fail", func(t *testing.T) {
		graph.failLabel = "Project"
		defer func() { graph.failLabel = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Project: "P", Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Source Save Fail", func(t *testing.T) {
		sources.fail = true
		defer func() { sources.fail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Source Graph Upsert Fail", func(t *testing.T) {
		graph.failLabel = "Source"
		defer func() { graph.failLabel = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Zettel Save Fail", func(t *testing.T) {
		zettels.fail = true
		defer func() { zettels.fail = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Zettel Graph Upsert Fail", func(t *testing.T) {
		graph.failLabel = "Zettel"
		defer func() { graph.failLabel = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Rel to Actor Fail", func(t *testing.T) {
		graph.failTo = "Person"
		defer func() { graph.failTo = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Participants: []string{"A"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Rel to Topic Fail", func(t *testing.T) {
		graph.failTo = "Topic"
		defer func() { graph.failTo = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Rel to Project Fail", func(t *testing.T) {
		graph.failTo = "Project"
		defer func() { graph.failTo = "" }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Project: "P", Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Rel Zettel to Source Fail", func(t *testing.T) {
		graph.failRel = true
		defer func() { graph.failRel = false }()
		_, _, err := service.IngestSummary(ctx, domain.SummaryPayload{Summary: "s", Conclusions: []string{"c"}})
		if err == nil {
			t.Error("expected error")
		}
	})
}
