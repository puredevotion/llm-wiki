package mcp

import (
	"context"
	"fmt"
	"log/slog"

	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type IdentityAction string

const (
	ActionCreateActor IdentityAction = "create_actor"
	ActionCreateTeam  IdentityAction = "create_team"
	ActionAddMember   IdentityAction = "add_member"
	ActionSetManager  IdentityAction = "set_manager"
)

type IdentityInput struct {
	Token  string         `json:"token" jsonschema:"agent api token"`
	Action IdentityAction `json:"action" jsonschema:"the identity management action to perform"`

	// Full entity objects
	Actor *domain.Actor `json:"actor,omitempty" jsonschema:"full actor data for create_actor"`
	Team  *domain.Team  `json:"team,omitempty" jsonschema:"full team data for create_team"`

	// Flat fields for convenience / other actions
	TeamID    string `json:"team_id,omitempty" jsonschema:"ID of the team"`
	ActorID   string `json:"actor_id,omitempty" jsonschema:"ID of the actor"`
	TargetID  string `json:"target_id,omitempty" jsonschema:"ID of the manager/target"`
	Role      string `json:"role,omitempty" jsonschema:"member, lead, etc."`
}

type IdentityOutput struct {
	ID     string `json:"id,omitempty"`
	Status string `json:"status"`
}

func registerIdentityTools(server *mcpsdk.Server, logger *slog.Logger, idSvc *services.IdentityService, agentToken string) {
	mcpsdk.AddTool(server, &mcpsdk.Tool{
		Name:        "kb.manage_identity",
		Description: "Manage people, agents, teams, and organizations in the knowledge base.",
	}, identityToolHandler(logger, idSvc, agentToken))
}

func identityToolHandler(logger *slog.Logger, idSvc *services.IdentityService, agentToken string) func(context.Context, *mcpsdk.CallToolRequest, IdentityInput) (*mcpsdk.CallToolResult, IdentityOutput, error) {
	return func(ctx context.Context, req *mcpsdk.CallToolRequest, input IdentityInput) (*mcpsdk.CallToolResult, IdentityOutput, error) {
		if input.Token != agentToken {
			return nil, IdentityOutput{}, fmt.Errorf("unauthorized: invalid agent token")
		}

		switch input.Action {
		case ActionCreateActor:
			if input.Actor == nil {
				return nil, IdentityOutput{}, fmt.Errorf("missing actor data")
			}
			if err := idSvc.CreateActor(ctx, input.Actor); err != nil {
				return nil, IdentityOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, IdentityOutput{ID: input.Actor.ID, Status: "actor_created"}, nil

		case ActionCreateTeam:
			if input.Team == nil {
				return nil, IdentityOutput{}, fmt.Errorf("missing team data")
			}
			if err := idSvc.UpsertTeam(ctx, input.Team); err != nil {
				return nil, IdentityOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, IdentityOutput{ID: input.Team.ID, Status: "team_created"}, nil

		case ActionAddMember:
			if err := idSvc.AddTeamMember(ctx, input.TeamID, input.ActorID, input.Role); err != nil {
				return nil, IdentityOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, IdentityOutput{Status: "member_added"}, nil

		case ActionSetManager:
			if err := idSvc.SetManager(ctx, input.ActorID, input.TargetID); err != nil {
				return nil, IdentityOutput{}, err
			}
			return &mcpsdk.CallToolResult{}, IdentityOutput{Status: "manager_set"}, nil

		default:
			return nil, IdentityOutput{}, fmt.Errorf("unsupported action: %s", input.Action)
		}
	}
}
