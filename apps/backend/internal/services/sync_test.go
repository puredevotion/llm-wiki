package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

func TestSyncService(t *testing.T) {
	opRepo := &mockOpRepo{ops: make(map[string]*domain.Operation)}
	zettelRepo := &mockZettelRepo{zettels: make(map[string]*domain.Zettel)}
	topicRepo := &mockTopicRepo{topics: make(map[string]*domain.Topic)}
	vRepo := &mockVectorRepo{}
	embeds := &mockEmbeddingsClient{}

	svc := NewSyncService(opRepo, zettelRepo, topicRepo, vRepo, embeds)
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
		payload, _ := json.Marshal(domain.Topic{Name: "T"})
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:            "opt1",
					EntityKind:    "topic",
					EntityID:      "t1",
					OperationType: "upsert",
					Payload:       payload,
				},
			},
		}
		results, _ := svc.ProcessBatch(ctx, batch)
		if results[0].Status != domain.OperationApplied {
			t.Error("expected applied for topic")
		}
	})

	t.Run("Apply Zettel Marshal Fail", func(t *testing.T) {
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:         "opm1",
					EntityKind: "zettel",
					Payload:    []byte("{bad}"),
				},
			},
		}
		results, _ := svc.ProcessBatch(ctx, batch)
		if results[0].Status != domain.OperationRejected {
			t.Error("expected rejection")
		}
	})

	t.Run("Apply Topic Marshal Fail", func(t *testing.T) {
		batch := domain.SyncBatch{
			Operations: []domain.Operation{
				{
					ID:         "opm2",
					EntityKind: "topic",
					Payload:    []byte("{bad}"),
				},
			},
		}
		results, _ := svc.ProcessBatch(ctx, batch)
		if results[0].Status != domain.OperationRejected {
			t.Error("expected rejection")
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
	})
}
