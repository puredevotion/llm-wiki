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
	failFrom  string
}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	if m.fail || (m.failLabel != "" && m.failLabel == label) {
		return fmt.Errorf("graph upsert failed")
	}
	m.nodes[id] = label
	return nil
}

func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	if m.fail || m.failRel || (m.failTo != "" && m.failTo == toLabel) || (m.failFrom != "" && m.failFrom == fromLabel) {
		return fmt.Errorf("graph relationship failed")
	}
	m.edges = append(m.edges, fmt.Sprintf("%s:%s:%s:%s:%s", fromID, fromLabel, toID, toLabel, relType))
	return nil
}

type mockOpRepo struct {
	ops map[string]*domain.Operation
	fail bool
}

func (m *mockOpRepo) Save(ctx context.Context, op *domain.Operation) error {
	if m.fail {
		return fmt.Errorf("op save failed")
	}
	m.ops[op.ID] = op
	return nil
}
func (m *mockOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) { return nil, nil }
func (m *mockOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) { return nil, nil }

func TestIngestSummary(t *testing.T) {
	ctx := context.Background()

	t.Run("Happy Path", func(t *testing.T) {
		service := NewIngestionService(&mockActorRepo{actors: make(map[string]*domain.Actor)}, &mockSourceRepo{sources: make(map[string]*domain.Source)}, &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}, &mockTopicRepo{topics: make(map[string]*domain.Topic)}, &mockGraphRepo{nodes: make(map[string]string)}, &mockOpRepo{ops: make(map[string]*domain.Operation)})
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
		service := NewIngestionService(&mockActorRepo{actors: make(map[string]*domain.Actor)}, &mockSourceRepo{sources: make(map[string]*domain.Source)}, &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}, &mockTopicRepo{topics: make(map[string]*domain.Topic)}, graph, &mockOpRepo{ops: make(map[string]*domain.Operation)})
		
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

	t.Run("Relationship Failures", func(t *testing.T) {
		graph := &mockGraphRepo{nodes: make(map[string]string)}
		service := NewIngestionService(&mockActorRepo{actors: make(map[string]*domain.Actor)}, &mockSourceRepo{sources: make(map[string]*domain.Source)}, &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}, &mockTopicRepo{topics: make(map[string]*domain.Topic)}, graph, &mockOpRepo{ops: make(map[string]*domain.Operation)})
		payload := domain.SummaryPayload{Project: "P", Participants: []string{"A"}, Topics: []string{"T"}, Summary: "s", Conclusions: []string{"c"}}

		// Fail Source -> Actor
		graph.failTo = "Person"
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on Source->Actor rel")
		}
		graph.failTo = ""

		// Fail Source -> Topic
		graph.failTo = "Topic"
		graph.failFrom = "Source"
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on Source->Topic rel")
		}
		graph.failTo = ""
		graph.failFrom = ""

		// Fail Source -> Project
		graph.failTo = "Project"
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on Source->Project rel")
		}
		graph.failTo = ""

		// Fail Zettel -> Source
		graph.failFrom = "Zettel"
		graph.failTo = "Source"
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on Zettel->Source rel")
		}
		graph.failFrom = ""
		graph.failTo = ""

		// Fail Zettel -> Topic
		graph.failFrom = "Zettel"
		graph.failTo = "Topic"
		if _, _, err := service.IngestSummary(ctx, payload); err == nil {
			t.Error("expected error on Zettel->Topic rel")
		}
	})

	t.Run("Repo Failures", func(t *testing.T) {
		actors := &mockActorRepo{actors: make(map[string]*domain.Actor)}
		topics := &mockTopicRepo{topics: make(map[string]*domain.Topic)}
		sources := &mockSourceRepo{sources: make(map[string]*domain.Source)}
		zettels := &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}
		
		service := NewIngestionService(actors, sources, zettels, topics, &mockGraphRepo{nodes: make(map[string]string)}, &mockOpRepo{ops: make(map[string]*domain.Operation)})
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
