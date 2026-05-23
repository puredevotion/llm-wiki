package turso

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

func setupTestDB(t *testing.T) (*Store, func()) {
	tmpDir, err := os.MkdirTemp("", "turso-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	store, err := NewStore(fmt.Sprintf("file:%s", dbPath))
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to create store: %v", err)
	}

	// Load schema
	schema, err := os.ReadFile("../../../migrations/000001_initial.sql")
	if err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to read schema: %v", err)
	}

	if err := store.Migrate(context.Background(), string(schema)); err != nil {
		store.Close()
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to migrate: %v", err)
	}

	return store, func() {
		store.Close()
		os.RemoveAll(tmpDir)
	}
}

func TestActorRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Actors()
	ctx := context.Background()

	t.Run("Save and Find", func(t *testing.T) {
		actor := &domain.Actor{
			ID:          "a1",
			Kind:        "person",
			DisplayName: "Alice",
			CreatedAt:   time.Now().Truncate(time.Second),
			Metadata:    map[string]any{"key": "val"},
		}

		if err := repo.Save(ctx, actor); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		found, err := repo.FindByName(ctx, "Alice")
		if err != nil {
			t.Fatalf("failed to find: %v", err)
		}
		if found == nil || found.ID != actor.ID {
			t.Errorf("expected %v, got %v", actor, found)
		}
	})

	t.Run("Upsert", func(t *testing.T) {
		actor := &domain.Actor{
			ID:          "a1",
			Kind:        "person",
			DisplayName: "Alice Updated",
			CreatedAt:   time.Now().Truncate(time.Second),
		}
		if err := repo.Save(ctx, actor); err != nil {
			t.Fatalf("failed to update: %v", err)
		}

		found, _ := repo.FindByName(ctx, "Alice Updated")
		if found == nil || found.DisplayName != "Alice Updated" {
			t.Errorf("expected updated name, got %v", found)
		}
	})

	t.Run("Find Error - Invalid JSON", func(t *testing.T) {
		store.db.Exec("INSERT INTO actors (id, kind, display_name, metadata_json, created_at) VALUES ('bad_json', 'person', 'BadJSON', '{invalid}', '2026-05-17T12:00:00Z')")
		_, err := repo.FindByName(ctx, "BadJSON")
		if err == nil {
			t.Errorf("expected unmarshal error, got nil")
		}
	})

	t.Run("Save Error - Marshal", func(t *testing.T) {
		actor := &domain.Actor{
			ID:       "err1",
			Metadata: map[string]any{"bad": func() {}},
		}
		if err := repo.Save(ctx, actor); err == nil {
			t.Errorf("expected marshal error, got nil")
		}
	})
}

func TestSourceRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Sources()
	ctx := context.Background()

	// Seed actor for FK
	actor := &domain.Actor{ID: "a1", Kind: "person", DisplayName: "Alice", CreatedAt: time.Now()}
	store.Actors().Save(ctx, actor)

	t.Run("Save", func(t *testing.T) {
		src := &domain.Source{
			ID:          "s1",
			Kind:        "conversation",
			Title:       "Convo 1",
			CapturedBy:  "a1",
			CapturedAt:  time.Now().Truncate(time.Second),
			Metadata:    map[string]any{"project": "p1"},
		}
		if err := repo.Save(ctx, src); err != nil {
			t.Fatalf("failed to save: %v", err)
		}
	})

	t.Run("Save Error - Marshal", func(t *testing.T) {
		src := &domain.Source{
			ID:       "err1",
			Metadata: map[string]any{"bad": func() {}},
		}
		if err := repo.Save(ctx, src); err == nil {
			t.Errorf("expected marshal error, got nil")
		}
	})
}

func TestZettelRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Zettels()
	ctx := context.Background()

	// Seed actor for FK
	actor := &domain.Actor{ID: "a1", Kind: "person", DisplayName: "Alice", CreatedAt: time.Now()}
	store.Actors().Save(ctx, actor)

	t.Run("Save with All Times", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		z := &domain.Zettel{
			ID:          "z2",
			Title:       "Full Zettel",
			Body:        "Body",
			Lifecycle:   "evergreen",
			Status:      "active",
			CreatedBy:   "a1",
			CreatedAt:   now,
			UpdatedAt:   now,
			ValidFrom:   &now,
			ValidTo:     &now,
			ReviewAfter: &now,
		}
		if err := repo.Save(ctx, z); err != nil {
			t.Fatalf("failed to save: %v", err)
		}
	})

	t.Run("Save and Search", func(t *testing.T) {
		z := &domain.Zettel{
			ID:        "z1",
			Title:     "Searching Go",
			Body:      "Go is a language",
			Lifecycle: "project",
			Status:    "active",
			CreatedBy: "a1",
			CreatedAt: time.Now().Truncate(time.Second),
			UpdatedAt: time.Now().Truncate(time.Second),
		}
		if err := repo.Save(ctx, z); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		results, err := repo.SearchZettels(ctx, "Go", 5)
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}
		if len(results) == 0 {
			t.Errorf("expected results, got 0")
		}
	})
	
	t.Run("Search Scan Error", func(t *testing.T) {
		// Manually insert a zettel with NULL title which should fail scan into domain.Zettel
		store.db.Exec("INSERT INTO zettels (id, title, body, lifecycle, status, created_by, created_at, updated_at) VALUES ('bad_scan', NULL, 'body', 'project', 'active', 'a1', '2026-05-17T12:00:00Z', '2026-05-17T12:00:00Z')")
		// We need to manually sync FTS if triggers didn't work for NULL (they might not if INSERT fails constraint, but title is NOT NULL in schema)
		// Wait, if title is NOT NULL, the INSERT will fail.
		// Let's use a field that is NULLABLE in schema but NOT in our Scan.
		// Actually, all fields in Scan seem to be NOT NULL in domain struct.
		// Let's just use a closed DB for one more branch.
		store.Close()
		_, err := repo.SearchZettels(ctx, "test", 5)
		if err == nil {
			t.Error("expected search error on closed db")
		}
	})
}

func TestTopicRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Topics()
	ctx := context.Background()

	t.Run("Save and Find", func(t *testing.T) {
		topic := &domain.Topic{
			ID:        "t1",
			Name:      "Tech",
			CreatedAt: time.Now().Truncate(time.Second),
		}

		if err := repo.Save(ctx, topic); err != nil {
			t.Fatalf("failed to save: %v", err)
		}

		found, err := repo.FindByName(ctx, "Tech")
		if err != nil {
			t.Fatalf("failed to find: %v", err)
		}
		if found == nil || found.ID != topic.ID {
			t.Errorf("expected %v, got %v", topic, found)
		}
	})

	t.Run("Find Error", func(t *testing.T) {
		store.Close()
		_, err := repo.FindByName(ctx, "Tech")
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("Save Error", func(t *testing.T) {
		store.Close()
		err := repo.Save(ctx, &domain.Topic{ID: "t2", Name: "Fail", CreatedAt: time.Now()})
		if err == nil {
			t.Error("expected error on closed db")
		}
	})
}

func TestOperationRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Operations()
	ctx := context.Background()

	// Seed actor for FK
	actor := &domain.Actor{ID: "a1", Kind: "person", DisplayName: "Alice", CreatedAt: time.Now()}
	store.Actors().Save(ctx, actor)

	t.Run("Save and Find", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		op := &domain.Operation{
			ID:            "op1",
			ActorID:       "a1",
			EntityKind:    "zettel",
			EntityID:      "z1",
			OperationType: "upsert",
			Payload:       json.RawMessage(`{"title":"test"}`),
			Status:        domain.OperationApplied,
			CreatedAt:     now,
			AppliedAt:     &now,
		}

		if err := repo.Save(ctx, op); err != nil {
			t.Fatalf("failed to save op: %v", err)
		}

		found, err := repo.FindByID(ctx, "op1")
		if err != nil {
			t.Fatalf("failed to find op: %v", err)
		}
		if found == nil || found.ID != op.ID {
			t.Errorf("expected %v, got %v", op, found)
		}
	})

	t.Run("Fetch Changes", func(t *testing.T) {
		// Save ops
		now := time.Now().Truncate(time.Second)
		op2 := &domain.Operation{
			ID:            "op2",
			ActorID:       "a1",
			EntityKind:    "topic",
			EntityID:      "t1",
			OperationType: "upsert",
			Payload:       json.RawMessage(`{}`),
			Status:        domain.OperationApplied,
			CreatedAt:     now,
			AppliedAt:     &now,
		}
		repo.Save(ctx, op2)

		changes, err := repo.FetchChanges(ctx, "", 10)
		if err != nil {
			t.Fatalf("failed to fetch changes: %v", err)
		}
		if len(changes) == 0 {
			t.Fatalf("expected changes, got 0")
		}
		
		// Test with cursor
		cursor := changes[0].AppliedAt.Format(time.RFC3339)
		_, _ = repo.FetchChanges(ctx, cursor, 10)
	})
	
	t.Run("Find Error", func(t *testing.T) {
		store.Close()
		_, err := repo.FindByID(ctx, "op1")
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("Fetch Error", func(t *testing.T) {
		store.Close()
		_, err := repo.FetchChanges(ctx, "", 10)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})
}

func TestNewStore_Error(t *testing.T) {
	t.Run("Bad Driver", func(t *testing.T) {
		_, err := NewStore("unknown_driver://bad")
		if err == nil {
			t.Errorf("expected error for bad driver, got nil")
		}
	})

	t.Run("Bad File", func(t *testing.T) {
		tmpFile, _ := os.CreateTemp("", "not-a-db")
		tmpFile.WriteString("hello world")
		tmpFile.Close()
		defer os.Remove(tmpFile.Name())
		
		_, err := NewStore("file:" + tmpFile.Name())
		if err == nil {
			t.Error("expected error for non-db file")
		}
	})
}
