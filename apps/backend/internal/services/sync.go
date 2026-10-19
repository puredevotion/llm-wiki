package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"llm-wiki/apps/backend/internal/clients/embeddings"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/repositories"
)

type SyncService struct {
	ops      repositories.OperationRepository
	zettels  repositories.ZettelRepository
	topics   repositories.TopicRepository
	vectors  repositories.VectorRepository
	timeline repositories.TimelineRepository
	actors   repositories.ActorRepository
	identity repositories.IdentityRepository
	embeds   embeddings.Client
}

func NewSyncService(
	ops repositories.OperationRepository,
	zettels repositories.ZettelRepository,
	topics repositories.TopicRepository,
	vectors repositories.VectorRepository,
	timeline repositories.TimelineRepository,
	actors repositories.ActorRepository,
	identity repositories.IdentityRepository,
	embeds embeddings.Client,
) *SyncService {
	return &SyncService{
		ops:      ops,
		zettels:  zettels,
		topics:   topics,
		vectors:  vectors,
		timeline: timeline,
		actors:   actors,
		identity: identity,
		embeds:   embeds,
	}
}

func (s *SyncService) ProcessBatch(ctx context.Context, batch domain.SyncBatch) ([]domain.Operation, error) {
	results := make([]domain.Operation, 0, len(batch.Operations))

	for _, op := range batch.Operations {
		// 1. Idempotency check
		existing, err := s.ops.FindByID(ctx, op.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to check idempotency for %s: %w", op.ID, err)
		}
		if existing != nil {
			results = append(results, *existing)
			continue
		}

		// 2. Reconciliation & Application
		appliedOp := op
		appliedOp.Status = domain.OperationApplied
		now := time.Now()
		appliedOp.AppliedAt = &now

		if err := s.applyOperation(ctx, &appliedOp); err != nil {
			appliedOp.Status = domain.OperationRejected
		}

		if err := s.ops.Save(ctx, &appliedOp); err != nil {
			return nil, fmt.Errorf("failed to save operation %s: %w", op.ID, err)
		}

		results = append(results, appliedOp)
	}

	return results, nil
}

func (s *SyncService) FetchChanges(ctx context.Context, cursor string, limit int) ([]domain.SyncChange, error) {
	if limit <= 0 {
		limit = 50
	}
	ops, err := s.ops.FetchChanges(ctx, cursor, limit)
	if err != nil {
		return nil, err
	}

	changes := make([]domain.SyncChange, 0, len(ops))
	for _, op := range ops {
		changes = append(changes, domain.SyncChange{
			Operation: *op,
			Cursor:    op.AppliedAt.Format(time.RFC3339),
		})
	}
	return changes, nil
}

func (s *SyncService) GetZettel(ctx context.Context, id string) (*domain.Zettel, error) {
	return s.zettels.FindByID(ctx, id)
}

func (s *SyncService) applyOperation(ctx context.Context, op *domain.Operation) error {
	switch op.EntityKind {
	case "zettel":
		return s.applyZettelOp(ctx, op)
	case "topic":
		return s.applyTopicOp(ctx, op)
	case "event":
		return s.applyEventOp(ctx, op)
	case "actor":
		return s.applyActorOp(ctx, op)
	case "team":
		return s.applyTeamOp(ctx, op)
	default:
		return fmt.Errorf("unsupported entity kind: %s", op.EntityKind)
	}
}

func (s *SyncService) applyZettelOp(ctx context.Context, op *domain.Operation) error {
	var z domain.Zettel
	if err := json.Unmarshal(op.Payload, &z); err != nil {
		return fmt.Errorf("invalid zettel payload: %w", err)
	}
	
	z.ID = op.EntityID
	z.UpdatedAt = time.Now()
	if z.CreatedAt.IsZero() {
		z.CreatedAt = z.UpdatedAt
	}

	if err := s.zettels.Save(ctx, &z); err != nil {
		return err
	}

	// Semantic Projection
	if s.embeds != nil && s.vectors != nil {
		text := fmt.Sprintf("%s\n\n%s", z.Title, z.Body)
		vec, err := s.embeds.Generate(ctx, text)
		if err == nil {
			_ = s.vectors.Upsert(ctx, z.ID, "zettel", vec, "text-embedding-3-small")
		}
	}
	return nil
}

func (s *SyncService) applyTopicOp(ctx context.Context, op *domain.Operation) error {
	var t domain.Topic
	if err := json.Unmarshal(op.Payload, &t); err != nil {
		return fmt.Errorf("invalid topic payload: %w", err)
	}

	t.ID = op.EntityID
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}

	if err := s.topics.Save(ctx, &t); err != nil {
		return err
	}

	// Semantic Projection
	if s.embeds != nil && s.vectors != nil {
		vec, err := s.embeds.Generate(ctx, t.Name)
		if err == nil {
			_ = s.vectors.Upsert(ctx, t.ID, "topic", vec, "text-embedding-3-small")
		}
	}
	return nil
}

func (s *SyncService) applyEventOp(ctx context.Context, op *domain.Operation) error {
	var e domain.Event
	if err := json.Unmarshal(op.Payload, &e); err != nil {
		return fmt.Errorf("invalid event payload: %w", err)
	}
	e.ID = op.EntityID
	if e.RecordedAt.IsZero() {
		e.RecordedAt = time.Now()
	}
	return s.timeline.Save(ctx, &e)
}

func (s *SyncService) applyActorOp(ctx context.Context, op *domain.Operation) error {
	var a domain.Actor
	if err := json.Unmarshal(op.Payload, &a); err != nil {
		return fmt.Errorf("invalid actor payload: %w", err)
	}
	a.ID = op.EntityID
	if a.CreatedAt.IsZero() {
		a.CreatedAt = time.Now()
	}
	return s.actors.Save(ctx, &a)
}

func (s *SyncService) applyTeamOp(ctx context.Context, op *domain.Operation) error {
	var t domain.Team
	if err := json.Unmarshal(op.Payload, &t); err != nil {
		return fmt.Errorf("invalid team payload: %w", err)
	}
	t.ID = op.EntityID
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	return s.identity.SaveTeam(ctx, &t)
}
