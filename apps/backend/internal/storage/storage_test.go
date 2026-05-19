package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/storage/graph"
	"llm-wiki/apps/backend/internal/storage/turso"
)

func TestStorageIntegration(t *testing.T) {
	ctx := context.Background()
	tmpDir, err := os.MkdirTemp("", "kbase-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// 1. Setup Turso
	dbPath := filepath.Join(tmpDir, "test.db")
	sqlStore, err := turso.NewStore(fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		t.Fatalf("failed to create turso store: %v", err)
	}
	defer sqlStore.Close()

	// Run migration
	schema, err := os.ReadFile("../../migrations/000001_initial.sql")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}
	if err := sqlStore.Migrate(ctx, string(schema)); err != nil {
		t.Fatalf("failed to migrate turso: %v", err)
	}

	// 2. Setup Graph
	graphPath := filepath.Join(tmpDir, "test_graph")
	graphStore, err := graph.NewStore(graphPath)
	if err != nil {
		t.Fatalf("failed to create graph store: %v", err)
	}
	defer graphStore.Close()

	if err := graphStore.Migrate(ctx); err != nil {
		t.Fatalf("failed to migrate graph: %v", err)
	}

	// 3. Test Repositories
	actors := sqlStore.Actors()
	topics := sqlStore.Topics()
	gRepo := graphStore.Graph()

	// Save Actor
	actor := &domain.Actor{
		ID:          "actor_1",
		Kind:        "person",
		DisplayName: "Alice",
		CreatedAt:   time.Now().Round(time.Second),
		Metadata:    map[string]any{"role": "admin"},
	}
	if err := actors.Save(ctx, actor); err != nil {
		t.Errorf("failed to save actor: %v", err)
	}

	// Find Actor
	foundActor, err := actors.FindByName(ctx, "Alice")
	if err != nil {
		t.Errorf("failed to find actor: %v", err)
	}
	if foundActor == nil || foundActor.ID != actor.ID {
		t.Errorf("expected actor %v, got %v", actor, foundActor)
	}

	// Save Topic
	topic := &domain.Topic{
		ID:        "topic_1",
		Name:      "Testing",
		CreatedAt: time.Now().Round(time.Second),
	}
	if err := topics.Save(ctx, topic); err != nil {
		t.Errorf("failed to save topic: %v", err)
	}

	// Test Graph Upsert
	if err := gRepo.UpsertNode(ctx, actor.ID, "Person", map[string]any{"name": actor.DisplayName}); err != nil {
		t.Errorf("failed to upsert graph node: %v", err)
	}
	if err := gRepo.UpsertNode(ctx, topic.ID, "Topic", map[string]any{"name": topic.Name}); err != nil {
		t.Errorf("failed to upsert topic node: %v", err)
	}

	// Test Graph Relationship
	// Create a dummy Source node first.
	if err := gRepo.UpsertNode(ctx, "src_1", "Source", map[string]any{"title": "Test Source", "kind": "paste"}); err != nil {
		t.Errorf("failed to upsert source node: %v", err)
	}
	if err := gRepo.CreateRelationship(ctx, "src_1", "Source", topic.ID, "Topic", "RELATED_TO"); err != nil {
		t.Errorf("failed to create relationship: %v", err)
	}
}
