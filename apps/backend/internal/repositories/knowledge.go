package repositories

import (
	"context"
	"time"
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
	FindByID(ctx context.Context, id string) (*domain.Zettel, error)
	SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error)
}

type TopicRepository interface {
	FindByName(ctx context.Context, name string) (*domain.Topic, error)
	Save(ctx context.Context, topic *domain.Topic) error
}

type IdentityRepository interface {
	SaveTeam(ctx context.Context, team *domain.Team) error
	SaveOrganization(ctx context.Context, org *domain.Organization) error
	AddTeamMember(ctx context.Context, member *domain.TeamMember) error
	FindTeamByName(ctx context.Context, name string) (*domain.Team, error)
}

type TimelineRepository interface {
	Save(ctx context.Context, event *domain.Event) error
	Fetch(ctx context.Context, startsAt, endsAt *time.Time, limit int) ([]*domain.Event, error)
}

type GraphRepository interface {
	// CreateRelationship creates a typed edge between two nodes.
	// Nodes are identified by their stable IDs from SQL stores.
	CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error
	// UpsertNode ensures a node exists with the given ID and label.
	UpsertNode(ctx context.Context, id, label string, properties map[string]any) error
	// FetchGraph returns the entire graph for visualization.
	FetchGraph(ctx context.Context) (*domain.GraphData, error)
}

type OperationRepository interface {
	Save(ctx context.Context, op *domain.Operation) error
	FindByID(ctx context.Context, id string) (*domain.Operation, error)
	FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error)
}

type VectorRepository interface {
	Upsert(ctx context.Context, entityID, entityKind string, vector domain.Vector, model string) error
	Search(ctx context.Context, entityKind string, vector domain.Vector, limit int) ([]string, error) // Returns entity IDs
}
