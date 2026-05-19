package repositories

import (
	"context"
	"llm-wiki/apps/backend/internal/domain"
)

type ActorRepository interface {
	FindByName(ctx context.Context, name string) (*domain.Actor, error)
	Save(ctx context.Context, actor *domain.Actor) error
}

type SourceRepository interface {
	Save(ctx context.Context, source *domain.Source) error
}

type ZettelRepository interface {
	Save(ctx context.Context, zettel *domain.Zettel) error
	SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error)
}

type TopicRepository interface {
	FindByName(ctx context.Context, name string) (*domain.Topic, error)
	Save(ctx context.Context, topic *domain.Topic) error
}

type GraphRepository interface {
	// CreateRelationship creates a typed edge between two nodes.
	// Nodes are identified by their stable IDs from SQL stores.
	CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error
	// UpsertNode ensures a node exists with the given ID and label.
	UpsertNode(ctx context.Context, id, label string, properties map[string]any) error
}
