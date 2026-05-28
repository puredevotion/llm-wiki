package services

import (
	"context"
	"fmt"
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

func (m *mockActorRepo) List(ctx context.Context, limit int) ([]*domain.Actor, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	var results []*domain.Actor
	for _, a := range m.actors {
		results = append(results, a)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
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
	results []*domain.Zettel
	fail    bool
}

func (m *mockZettelRepo) Save(ctx context.Context, zettel *domain.Zettel) error {
	if m.fail {
		return fmt.Errorf("zettel save failed")
	}
	m.zettels[zettel.ID] = zettel
	return nil
}

func (m *mockZettelRepo) FindByID(ctx context.Context, id string) (*domain.Zettel, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return m.zettels[id], nil
}

func (m *mockZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	if m.fail {
		return nil, fmt.Errorf("search failed")
	}
	return m.results, nil
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

type mockIdentityRepo struct {
	teams map[string]*domain.Team
	fail  bool
}

func (m *mockIdentityRepo) SaveTeam(ctx context.Context, team *domain.Team) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	m.teams[team.ID] = team
	return nil
}
func (m *mockIdentityRepo) SaveOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}
func (m *mockIdentityRepo) AddTeamMember(ctx context.Context, member *domain.TeamMember) error {
	return nil
}
func (m *mockIdentityRepo) FindTeamByName(ctx context.Context, name string) (*domain.Team, error) {
	for _, t := range m.teams {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockIdentityRepo) ListTeams(ctx context.Context, limit int) ([]*domain.Team, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	var results []*domain.Team
	for _, t := range m.teams {
		results = append(results, t)
		if len(results) >= limit {
			break
		}
	}
	return results, nil
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
func (m *mockGraphRepo) FetchGraph(ctx context.Context) (*domain.GraphData, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return &domain.GraphData{}, nil
}

type mockOpRepo struct {
	ops  map[string]*domain.Operation
	fail bool
}

func (m *mockOpRepo) Save(ctx context.Context, op *domain.Operation) error {
	if m.fail {
		return fmt.Errorf("op save failed")
	}
	m.ops[op.ID] = op
	return nil
}
func (m *mockOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return m.ops[id], nil
}
func (m *mockOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	var results []*domain.Operation
	for _, op := range m.ops {
		if op.Status == domain.OperationApplied {
			if op.AppliedAt == nil {
				now := time.Now()
				op.AppliedAt = &now
			}
			results = append(results, op)
		}
		if len(results) >= limit {
			break
		}
	}
	return results, nil
}

type mockVectorRepo struct {
	ids  []string
	fail bool
}

func (m *mockVectorRepo) Upsert(ctx context.Context, id, kind string, vec domain.Vector, model string) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	return nil
}
func (m *mockVectorRepo) Search(ctx context.Context, kind string, query domain.Vector, limit int) ([]string, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return m.ids, nil
}

type mockEmbeddingsClient struct {
	fail bool
}

func (m *mockEmbeddingsClient) Generate(ctx context.Context, text string) (domain.Vector, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	return domain.Vector{0.1}, nil
}
func (m *mockEmbeddingsClient) BatchGenerate(ctx context.Context, texts []string) ([]domain.Vector, error) {
	return nil, nil
}
