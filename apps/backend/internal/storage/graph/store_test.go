package graph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func setupTestGraph(t *testing.T) (*Store, func()) {
	tmpDir, err := os.MkdirTemp("", "graph-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	graphPath := filepath.Join(tmpDir, "test_graph")
	store, err := NewStore(graphPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create store: %v", err)
	}

	if err := store.Migrate(context.Background()); err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	return store, func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}
}

func TestGraphRepo(t *testing.T) {
	store, cleanup := setupTestGraph(t)
	defer cleanup()

	repo := store.Graph()
	ctx := context.Background()

	t.Run("Upsert and Update Node", func(t *testing.T) {
		id := "p1"
		// Initial Create
		if err := repo.UpsertNode(ctx, id, "Person", map[string]any{"name": "Alice"}); err != nil {
			t.Fatalf("failed to create node: %v", err)
		}

		// Update (triggers SET logic in UpsertNode)
		if err := repo.UpsertNode(ctx, id, "Person", map[string]any{"name": "Alice Updated"}); err != nil {
			t.Fatalf("failed to update node: %v", err)
		}
	})

	t.Run("Create Relationship", func(t *testing.T) {
		// Create nodes first
		repo.UpsertNode(ctx, "s1", "Source", map[string]any{"title": "Src", "kind": "kb"})
		repo.UpsertNode(ctx, "t1", "Topic", map[string]any{"name": "Topic"})

		if err := repo.CreateRelationship(ctx, "s1", "Source", "t1", "Topic", "RELATED_TO"); err != nil {
			t.Fatalf("failed to create relationship: %v", err)
		}
	})

	t.Run("Upsert Invalid Label", func(t *testing.T) {
		if err := repo.UpsertNode(ctx, "bad1", "NonExistentTable", map[string]any{"x": 1}); err == nil {
			t.Errorf("expected error for invalid label, got nil")
		}
	})

	t.Run("Upsert Invalid Property", func(t *testing.T) {
		if err := repo.UpsertNode(ctx, "p2", "Person", map[string]any{"invalid_prop": 1}); err == nil {
			t.Errorf("expected error for invalid property, got nil")
		}
	})

	t.Run("Create Relationship Error", func(t *testing.T) {
		if err := repo.CreateRelationship(ctx, "s1", "Source", "t1", "Topic", "NON_EXISTENT_REL"); err == nil {
			t.Errorf("expected error for invalid rel type, got nil")
		}
	})

	t.Run("Migrate Idempotency", func(t *testing.T) {
		if err := store.Migrate(ctx); err != nil {
			t.Errorf("expected migrate to be idempotent, got %v", err)
		}
	})
}
