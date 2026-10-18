package mcp

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

func TestIdentityToolHandler(t *testing.T) {
	agentToken := "test-token"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	
	actors := &mockActorRepo{}
	identity := &mockIdentityRepo{}
	graph := &mockGraphRepo{}
	ops := &mockOpRepo{}
	
	idSvc := services.NewIdentityService(actors, identity, graph, ops)
	handler := identityToolHandler(logger, idSvc, agentToken)

	t.Run("Create Actor", func(t *testing.T) {
		input := IdentityInput{
			Token:            "test-token",
			Action:           ActionCreateActor,
			ActorDisplayName: "Bob",
			ActorKind:        "person",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "actor_created" {
			t.Errorf("expected actor_created, got %s", output.Status)
		}
	})

	t.Run("Create Team", func(t *testing.T) {
		input := IdentityInput{
			Token:    "test-token",
			Action:   ActionCreateTeam,
			TeamName: "Engineering",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "team_created" {
			t.Errorf("expected team_created, got %s", output.Status)
		}
	})

	t.Run("Add Member", func(t *testing.T) {
		input := IdentityInput{
			Token:   "test-token",
			Action:  ActionAddMember,
			TeamID:  "t1",
			ActorID: "a1",
			Role:    "member",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "member_added" {
			t.Errorf("expected member_added, got %s", output.Status)
		}
	})

	t.Run("Set Manager", func(t *testing.T) {
		input := IdentityInput{
			Token:     "test-token",
			Action:    ActionSetManager,
			ActorID:   "a1",
			ManagerID: "a2",
		}
		_, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "manager_set" {
			t.Errorf("expected manager_set, got %s", output.Status)
		}
	})

	t.Run("Unauthorized", func(t *testing.T) {
		input := IdentityInput{Token: "bad"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error for bad token")
		}
	})

	t.Run("Unsupported Action", func(t *testing.T) {
		input := IdentityInput{Token: "test-token", Action: "bad"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected error for unsupported action")
		}
	})
}

type mockIdentityRepo struct{}

func (m *mockIdentityRepo) SaveTeam(ctx context.Context, team *domain.Team) error { return nil }
func (m *mockIdentityRepo) SaveOrganization(ctx context.Context, org *domain.Organization) error {
	return nil
}
func (m *mockIdentityRepo) AddTeamMember(ctx context.Context, member *domain.TeamMember) error {
	return nil
}
func (m *mockIdentityRepo) FindTeamByName(ctx context.Context, name string) (*domain.Team, error) {
	return nil, nil
}
