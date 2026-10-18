package services

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

type mockSyncOpRepo struct {
	ops  map[string]*domain.Operation
	fail bool
}

func (m *mockSyncOpRepo) Save(ctx context.Context, op *domain.Operation) error {
	if m.fail {
		return fmt.Errorf("op save failed")
	}
	m.ops[op.ID] = op
	return nil
}

func (m *mockSyncOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	if m.fail {
		return nil, fmt.Errorf("find error")
	}
	return m.ops[id], nil
}

func (m *mockSyncOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	if m.fail {
		return nil, fmt.Errorf("fetch error")
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

type mockSyncZettelRepo struct {
	zettels map[string]*domain.Zettel
}

func (m *mockSyncZettelRepo) Save(ctx context.Context, z *domain.Zettel) error {
	m.zettels[z.ID] = z
	return nil
}
func (m *mockSyncZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	return nil, nil
}

type mockSyncTopicRepo struct {
	topics map[string]*domain.Topic
}

func (m *mockSyncTopicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	for _, t := range m.topics {
		if t.Name == name {
			return t, nil
		}
	}
	return nil, nil
}
func (m *mockSyncTopicRepo) Save(ctx context.Context, t *domain.Topic) error {
	m.topics[t.ID] = t
	return nil
}

func TestSyncService(t *testing.T) {
	opRepo := &mockSyncOpRepo{ops: make(map[string]*domain.Operation)}
	zettelRepo := &mockSyncZettelRepo{zettels: make(map[string]*domain.Zettel)}
	topicRepo := &mockSyncTopicRepo{topics: make(map[string]*domain.Topic)}

	svc := NewSyncService(opRepo, zettelRepo, topicRepo)
	ctx := context.Background()

	t.Run("Apply Zettel Upsert", func(t *testing.T) {
		payload, _ := json.Marshal(domain.Zettel{Title: "New Zettel", Body: "Content", Lifecycle: "project", Status: "active"})
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:            "op1",
					EntityKind:    "zettel",
					EntityID:      "z1",
					OperationType: "upsert",
					Payload:       payload,
					CreatedAt:     time.Now(),
				},
			},
		}

		results, err := svc.ProcessBatch(ctx, batch)
		if err != nil {
			t.Fatalf("ProcessBatch failed: %v", err)
		}
		if results[0].Status != domain.OperationApplied {
			t.Errorf("expected applied, got %s", results[0].Status)
		}
	})

	t.Run("Apply Topic Upsert", func(t *testing.T) {
		payload, _ := json.Marshal(domain.Topic{Name: "New Topic"})
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:            "op_t1",
					EntityKind:    "topic",
					EntityID:      "t1",
					OperationType: "upsert",
					Payload:       payload,
					CreatedAt:     time.Now(),
				},
			},
		}
		results, _ := svc.ProcessBatch(ctx, batch)
		if results[0].Status != domain.OperationApplied {
			t.Error("topic op failed")
		}
	})

	t.Run("Unsupported Entity", func(t *testing.T) {
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:         "op_bad",
					EntityKind: "unknown",
					Payload:    json.RawMessage(`{}`),
				},
			},
		}
		results, _ := svc.ProcessBatch(ctx, batch)
		if results[0].Status != domain.OperationRejected {
			t.Error("expected rejection for unknown entity")
		}
	})

	t.Run("Existing Operation Idempotency", func(t *testing.T) {
		// Create an existing op
		op := &domain.Operation{ID: "existing1", Status: domain.OperationApplied}
		opRepo.Save(ctx, op)
		
		batch := domain.SyncBatch{
			Operations: []domain.Operation{{ID: "existing1"}},
		}
		results, err := svc.ProcessBatch(ctx, batch)
		if err != nil {
			t.Fatalf("ProcessBatch failed: %v", err)
		}
		if len(results) != 1 || results[0].ID != "existing1" {
			t.Error("expected existing op to be returned")
		}
	})

	t.Run("Fetch Changes", func(t *testing.T) {
		// Apply an op first to have something to fetch
		payload, _ := json.Marshal(domain.Zettel{Title: "Z"})
		svc.ProcessBatch(ctx, domain.SyncBatch{Operations: []domain.Operation{{ID: "f1", EntityKind: "zettel", Payload: payload}}})

		changes, err := svc.FetchChanges(ctx, "", 10)
		if err != nil {
			t.Fatalf("FetchChanges failed: %v", err)
		}
		if len(changes) == 0 {
			t.Error("expected changes, got 0")
		}
		
		// Test with cursor
		cursor := changes[0].Cursor
		_, err = svc.FetchChanges(ctx, cursor, 10)
		if err != nil {
			t.Errorf("FetchChanges with cursor failed: %v", err)
		}
	})

	t.Run("Op Find Error", func(t *testing.T) {
		opRepo.fail = true
		defer func() { opRepo.fail = false }()
		_, err := svc.ProcessBatch(ctx, domain.SyncBatch{Operations: []domain.Operation{{ID: "x"}}})
		if err == nil {
			t.Error("expected error on repo find failure")
		}
	})

	t.Run("Fetch Changes Error", func(t *testing.T) {
		opRepo.fail = true
		defer func() { opRepo.fail = false }()
		_, err := svc.FetchChanges(ctx, "", 10)
		if err == nil {
			t.Error("expected error on fetch failure")
		}
	})
}
