package services

import (
	"context"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
)

func TestIdentityService(t *testing.T) {
	actors := &mockActorRepo{actors: make(map[string]*domain.Actor)}
	identity := &mockIdentityRepo{teams: make(map[string]*domain.Team)}
	graph := &mockGraphRepo{nodes: make(map[string]string)}
	ops := &mockOpRepo{ops: make(map[string]*domain.Operation)}

	svc := NewIdentityService(actors, identity, graph, ops)
	ctx := context.Background()

	t.Run("Create Actor", func(t *testing.T) {
		a := &domain.Actor{DisplayName: "Alice", Kind: "person"}
		if err := svc.CreateActor(ctx, a); err != nil {
			t.Fatalf("CreateActor failed: %v", err)
		}
		if actors.actors[a.ID] == nil {
			t.Error("actor was not saved to repo")
		}
		if graph.nodes[a.ID] != "Person" {
			t.Error("actor was not saved to graph")
		}
	})

	t.Run("Upsert Team", func(t *testing.T) {
		team := &domain.Team{Name: "Core"}
		if err := svc.UpsertTeam(ctx, team); err != nil {
			t.Fatalf("UpsertTeam failed: %v", err)
		}
		if identity.teams[team.ID] == nil {
			t.Error("team was not saved to repo")
		}
		if graph.nodes[team.ID] != "Team" {
			t.Error("team was not saved to graph")
		}
	})

	t.Run("Add Member", func(t *testing.T) {
		if err := svc.AddTeamMember(ctx, "t1", "a1", "lead"); err != nil {
			t.Fatalf("AddTeamMember failed: %v", err)
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
		graph.failRel = true
		defer func() { graph.failRel = false }()
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
		identity.fail = false
	})
}
