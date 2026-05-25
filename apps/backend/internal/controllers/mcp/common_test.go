package mcp

import (
	"context"
	"llm-wiki/apps/backend/internal/domain"
)

type mockActorRepo struct{}

func (m *mockActorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	if name == "Alice" {
		return &domain.Actor{ID: "a1", DisplayName: "Alice"}, nil
	}
	return nil, nil
}
func (m *mockActorRepo) Save(ctx context.Context, actor *domain.Actor) error { return nil }

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

type mockGraphRepo struct{}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	return nil
}
func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	return nil
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
