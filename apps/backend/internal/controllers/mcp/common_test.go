package mcp

import (
	"context"
	"fmt"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

type mockActorRepo struct {
	fail bool
}

func (m *mockActorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	if name == "Alice" {
		return &domain.Actor{ID: "a1", DisplayName: "Alice"}, nil
	}
	return nil, nil
}
func (m *mockActorRepo) Save(ctx context.Context, actor *domain.Actor) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	return nil
}

type mockSourceRepo struct{}

func (m *mockSourceRepo) Save(ctx context.Context, source *domain.Source) error { return nil }

type mockZettelRepo struct{}

func (m *mockZettelRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockZettelRepo) FindByID(ctx context.Context, id string) (*domain.Zettel, error) {
	return nil, nil
}
func (m *mockZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	return []*domain.Zettel{}, nil
}

type mockTopicRepo struct{}

func (m *mockTopicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) Save(ctx context.Context, topic *domain.Topic) error { return nil }

type mockGraphRepo struct {
	failRel bool
}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	return nil
}
func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	if m.failRel {
		return fmt.Errorf("fail")
	}
	return nil
}
func (m *mockGraphRepo) FetchGraph(ctx context.Context) (*domain.GraphData, error) {
	return &domain.GraphData{}, nil
}

type mockOpRepo struct{}

func (m *mockOpRepo) Save(ctx context.Context, op *domain.Operation) error { return nil }
func (m *mockOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	return nil, nil
}
func (m *mockOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	return nil, nil
}

type mockVectorRepo struct{}

func (m *mockVectorRepo) Upsert(ctx context.Context, id, kind string, vec domain.Vector, model string) error {
	return nil
}
func (m *mockVectorRepo) Search(ctx context.Context, kind string, vector domain.Vector, limit int) ([]string, error) {
	return nil, nil
}

type mockEmbeddingsClient struct{}

func (m *mockEmbeddingsClient) Generate(ctx context.Context, text string) (domain.Vector, error) {
	return domain.Vector{0.1}, nil
}
func (m *mockEmbeddingsClient) BatchGenerate(ctx context.Context, texts []string) ([]domain.Vector, error) {
	return nil, nil
}

type mockTimelineRepo struct {
	events map[string]*domain.Event
	fail   bool
}

func (m *mockTimelineRepo) Save(ctx context.Context, e *domain.Event) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	if m.events == nil {
		m.events = make(map[string]*domain.Event)
	}
	m.events[e.ID] = e
	return nil
}
func (m *mockTimelineRepo) FindByID(ctx context.Context, id string) (*domain.Event, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return m.events[id], nil
}
func (m *mockTimelineRepo) Fetch(ctx context.Context, startsAt, endsAt *time.Time, limit int) ([]*domain.Event, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return nil, nil
}
