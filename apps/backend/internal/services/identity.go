package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type IdentityService struct {
	actors   repositories.ActorRepository
	identity repositories.IdentityRepository
	graph    repositories.GraphRepository
	ops      repositories.OperationRepository
}

func NewIdentityService(
	actors repositories.ActorRepository,
	identity repositories.IdentityRepository,
	graph repositories.GraphRepository,
	ops repositories.OperationRepository,
) *IdentityService {
	return &IdentityService{
		actors:   actors,
		identity: identity,
		graph:    graph,
		ops:      ops,
	}
}

func (s *IdentityService) CreateActor(ctx context.Context, a *domain.Actor) error {
	if a.ID == "" {
		a.ID = fmt.Sprintf("actor_%d", time.Now().UnixNano())
	}
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}

	// 1. SQL Save
	if err := s.actors.Save(ctx, a); err != nil {
		return fmt.Errorf("failed to save actor: %w", err)
	}

	// 2. Graph Save
	if err := s.graph.UpsertNode(ctx, a.ID, "Person", map[string]any{"name": a.DisplayName}); err != nil {
		return fmt.Errorf("failed to upsert graph node: %w", err)
	}

	// 3. Sync Log
	return s.logOperation(ctx, "actor", a.ID, "upsert", a)
}

func (s *IdentityService) UpsertTeam(ctx context.Context, t *domain.Team) error {
	if t.ID == "" {
		t.ID = fmt.Sprintf("team_%d", time.Now().UnixNano())
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	t.UpdatedAt = time.Now()

	// 1. SQL Save
	if err := s.identity.SaveTeam(ctx, t); err != nil {
		return fmt.Errorf("failed to save team: %w", err)
	}

	// 2. Graph Save
	if err := s.graph.UpsertNode(ctx, t.ID, "Team", map[string]any{"name": t.Name}); err != nil {
		return fmt.Errorf("failed to upsert team node: %w", err)
	}
	if t.OrgID != "" {
		if err := s.graph.CreateRelationship(ctx, t.ID, "Team", t.OrgID, "Organization", "PART_OF"); err != nil {
			return fmt.Errorf("failed to link team to org: %w", err)
		}
	}

	// 3. Sync Log
	return s.logOperation(ctx, "team", t.ID, "upsert", t)
}

func (s *IdentityService) AddTeamMember(ctx context.Context, teamID, actorID, role string) error {
	m := &domain.TeamMember{
		TeamID:    teamID,
		ActorID:   actorID,
		Role:      role,
		CreatedAt: time.Now(),
	}

	// 1. SQL Save
	if err := s.identity.AddTeamMember(ctx, m); err != nil {
		return fmt.Errorf("failed to add team member: %w", err)
	}

	// 2. Graph Save
	if err := s.graph.CreateRelationship(ctx, actorID, "Person", teamID, "Team", "MEMBER_OF"); err != nil {
		return fmt.Errorf("failed to link member to team: %w", err)
	}

	// 3. Sync Log
	return s.logOperation(ctx, "team_member", fmt.Sprintf("%s:%s", teamID, actorID), "add", m)
}

func (s *IdentityService) SetManager(ctx context.Context, actorID, managerID string) error {
	// 1. Graph edge is the primary store for management reporting
	if err := s.graph.CreateRelationship(ctx, actorID, "Person", managerID, "Person", "REPORTS_TO"); err != nil {
		return fmt.Errorf("failed to link manager: %w", err)
	}

	// 2. Sync Log
	return s.logOperation(ctx, "management", actorID, "set_manager", map[string]string{"manager_id": managerID})
}

func (s *IdentityService) logOperation(ctx context.Context, kind, entityID, opType string, payload any) error {
	pJSON, _ := json.Marshal(payload)
	op := &domain.Operation{
		ID:            fmt.Sprintf("op_id_%d", time.Now().UnixNano()),
		EntityKind:    kind,
		EntityID:      entityID,
		OperationType: opType,
		Payload:       pJSON,
		Status:        domain.OperationApplied,
		CreatedAt:     time.Now(),
	}
	now := time.Now()
	op.AppliedAt = &now
	return s.ops.Save(ctx, op)
}

func (s *IdentityService) ListActors(ctx context.Context, limit int) ([]*domain.Actor, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.actors.List(ctx, limit)
}

func (s *IdentityService) ListTeams(ctx context.Context, limit int) ([]*domain.Team, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.identity.ListTeams(ctx, limit)
}

func (s *IdentityService) ListOrganizations(ctx context.Context, limit int) ([]*domain.Organization, error) {
	if limit <= 0 {
		limit = 100
	}
	return s.identity.ListOrganizations(ctx, limit)
}
