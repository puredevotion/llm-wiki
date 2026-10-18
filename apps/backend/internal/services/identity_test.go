package services

import (
	"context"
	"fmt"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
)

type mockIdentityRepo struct {
	teams   map[string]*domain.Team
	orgs    map[string]*domain.Organization
	members []*domain.TeamMember
	fail    bool
}

func (m *mockIdentityRepo) SaveTeam(ctx context.Context, t *domain.Team) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	m.teams[t.ID] = t
	return nil
}
func (m *mockIdentityRepo) SaveOrganization(ctx context.Context, o *domain.Organization) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	m.orgs[o.ID] = o
	return nil
}
func (m *mockIdentityRepo) AddTeamMember(ctx context.Context, mem *domain.TeamMember) error {
	if m.fail {
		return fmt.Errorf("fail")
	}
	m.members = append(m.members, mem)
	return nil
}
func (m *mockIdentityRepo) FindTeamByName(ctx context.Context, name string) (*domain.Team, error) {
	if m.fail {
		return nil, fmt.Errorf("fail")
	}
	for _, t := range m.teams {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, nil
}

func TestIdentityService(t *testing.T) {
	actors := &mockActorRepo{actors: make(map[string]*domain.Actor)}
	identity := &mockIdentityRepo{teams: make(map[string]*domain.Team), orgs: make(map[string]*domain.Organization)}
	graph := &mockGraphRepo{nodes: make(map[string]string)}
	ops := &mockOpRepo{ops: make(map[string]*domain.Operation)}

	svc := NewIdentityService(actors, identity, graph, ops)
	ctx := context.Background()

	t.Run("Create Actor", func(t *testing.T) {
		a := &domain.Actor{DisplayName: "Alice", Kind: "person"}
		if err := svc.CreateActor(ctx, a); err != nil {
			t.Fatalf("CreateActor failed: %v", err)
		}
		if _, ok := actors.actors[a.ID]; !ok {
			t.Error("actor was not saved to repo")
		}
	})

	t.Run("Upsert Team", func(t *testing.T) {
		tm := &domain.Team{Name: "Core", OrgID: "o1"}
		if err := svc.UpsertTeam(ctx, tm); err != nil {
			t.Fatalf("UpsertTeam failed: %v", err)
		}
		if _, ok := identity.teams[tm.ID]; !ok {
			t.Error("team was not saved to repo")
		}
	})

	t.Run("Add Member", func(t *testing.T) {
		if err := svc.AddTeamMember(ctx, "t1", "a1", "lead"); err != nil {
			t.Fatalf("AddTeamMember failed: %v", err)
		}
		if len(identity.members) == 0 {
			t.Error("member was not added")
		}
	})

	t.Run("Save Organization", func(t *testing.T) {
		o := &domain.Organization{Name: "Acme"}
		if err := identity.SaveOrganization(ctx, o); err != nil {
			t.Error("failed to save org")
		}
	})

	t.Run("Set Manager", func(t *testing.T) {
		if err := svc.SetManager(ctx, "a1", "a2"); err != nil {
			t.Fatalf("SetManager failed: %v", err)
		}
	})

	t.Run("Set Manager Graph Fail", func(t *testing.T) {
		graph.fail = true
		defer func() { graph.fail = false }()
		if err := svc.SetManager(ctx, "a1", "a2"); err == nil {
			t.Error("expected error")
		}
	})

	t.Run("Failures", func(t *testing.T) {
		actors.fail = true
		if err := svc.CreateActor(ctx, &domain.Actor{}); err == nil {
			t.Error("expected actor save fail")
		}
		actors.fail = false

		identity.fail = true
		if err := svc.UpsertTeam(ctx, &domain.Team{}); err == nil {
			t.Error("expected team save fail")
		}
		if err := svc.AddTeamMember(ctx, "t", "a", "r"); err == nil {
			t.Error("expected member add fail")
		}
		identity.fail = false

		graph.fail = true
		if err := svc.UpsertTeam(ctx, &domain.Team{}); err == nil {
			t.Error("expected team graph fail")
		}
		if err := svc.AddTeamMember(ctx, "t", "a", "r"); err == nil {
			t.Error("expected member graph fail")
		}
		graph.fail = false
	})
}
