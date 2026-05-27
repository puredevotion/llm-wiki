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
	
	t.Run("Search Error", func(t *testing.T) {
		s2, c2 := setupTestDB(t)
		r2 := s2.Zettels()
		c2() // close
		_, err := r2.SearchZettels(ctx, "test", 5)
		if err == nil {
			t.Error("expected search error on closed db")
		}
	})
}

func TestZettelRepo_FindByID(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Zettels()
	ctx := context.Background()

	// Seed actor for FK
	actor := &domain.Actor{ID: "a1", Kind: "person", DisplayName: "Alice", CreatedAt: time.Now()}
	store.Actors().Save(ctx, actor)

	t.Run("Find Success", func(t *testing.T) {
		z := &domain.Zettel{ID: "zf1", Title: "Find Me", Lifecycle: "project", Status: "active", CreatedBy: "a1", CreatedAt: time.Now(), UpdatedAt: time.Now()}
		if err := repo.Save(ctx, z); err != nil {
			t.Fatalf("save failed: %v", err)
		}
		
		found, err := repo.FindByID(ctx, "zf1")
		if err != nil || found == nil {
			t.Errorf("failed to find zettel: %v, found: %v", err, found)
		}
	})

	t.Run("Not Found", func(t *testing.T) {
		found, err := repo.FindByID(ctx, "missing")
		if err != nil || found != nil {
			t.Errorf("expected nil for missing, got %v, %v", found, err)
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
		s2, c2 := setupTestDB(t)
		r2 := s2.Topics()
		c2()
		_, err := r2.FindByName(ctx, "Tech")
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("Save Error", func(t *testing.T) {
		s2, c2 := setupTestDB(t)
		r2 := s2.Topics()
		c2()
		err := r2.Save(ctx, &domain.Topic{ID: "t2", Name: "Fail", CreatedAt: time.Now()})
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
	})
	
	t.Run("Find Error", func(t *testing.T) {
		s2, c2 := setupTestDB(t)
		r2 := s2.Operations()
		c2()
		_, err := r2.FindByID(ctx, "op1")
		if err == nil {
			t.Error("expected error on closed db")
		}
	})

	t.Run("Fetch Error", func(t *testing.T) {
		s2, c2 := setupTestDB(t)
		r2 := s2.Operations()
		c2()
		_, err := r2.FetchChanges(ctx, "", 10)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})
}

func TestIdentityRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Identity()
	ctx := context.Background()

	t.Run("Save Team and Organization", func(t *testing.T) {
		org := &domain.Organization{ID: "o1", Name: "Acme"}
		if err := repo.SaveOrganization(ctx, org); err != nil {
			t.Fatalf("failed to save org: %v", err)
		}

		team := &domain.Team{ID: "t1", OrgID: "o1", Name: "Core"}
		if err := repo.SaveTeam(ctx, team); err != nil {
			t.Fatalf("failed to save team: %v", err)
		}

		found, err := repo.FindTeamByName(ctx, "Core")
		if err != nil {
			t.Fatalf("failed to find team: %v", err)
		}
		if found == nil || found.ID != "t1" {
			t.Errorf("expected team t1, got %v", found)
		}
	})

	t.Run("Team Membership", func(t *testing.T) {
		actor := &domain.Actor{ID: "a1", DisplayName: "Alice", Kind: "person"}
		store.Actors().Save(ctx, actor)

		member := &domain.TeamMember{TeamID: "t1", ActorID: "a1", Role: "lead", CreatedAt: time.Now()}
		if err := repo.AddTeamMember(ctx, member); err != nil {
			t.Fatalf("failed to add member: %v", err)
		}
	})
}

func TestVectorRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Vectors()
	ctx := context.Background()

	t.Run("Upsert and Search", func(t *testing.T) {
		v1 := domain.Vector{1.0, 0.0, 0.0}
		if err := repo.Upsert(ctx, "z1", "zettel", v1, "model1"); err != nil {
			t.Fatalf("upsert failed: %v", err)
		}
		
		v2 := domain.Vector{0.0, 1.0, 0.0}
		repo.Upsert(ctx, "z2", "zettel", v2, "model1")

		// Search with v1
		ids, err := repo.Search(ctx, "zettel", v1, 10)
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(ids) == 0 || ids[0] != "z1" {
			t.Errorf("expected z1 as top match, got %v", ids)
		}
	})

	t.Run("Search Different Lengths", func(t *testing.T) {
		repo.Upsert(ctx, "vlen", "zettel", domain.Vector{1, 2}, "m1")
		_, err := repo.Search(ctx, "zettel", domain.Vector{1}, 5)
		if err != nil {
			t.Fatalf("search should not fail on length mismatch: %v", err)
		}
	})

	t.Run("Search Zero Norm", func(t *testing.T) {
		repo.Upsert(ctx, "vzero", "zettel", domain.Vector{0, 0}, "m1")
		repo.Search(ctx, "zettel", domain.Vector{1, 1}, 5)
	})

	t.Run("Search Empty Vectors", func(t *testing.T) {
		repo.Search(ctx, "zettel", domain.Vector{}, 5)
	})

	t.Run("Search Error", func(t *testing.T) {
		s2, c2 := setupTestDB(t)
		r2 := s2.Vectors()
		c2()
		_, err := r2.Search(ctx, "zettel", domain.Vector{1, 1}, 5)
		if err == nil {
			t.Error("expected error on closed db")
		}
	})
}

func TestTimelineRepo(t *testing.T) {
	store, cleanup := setupTestDB(t)
	defer cleanup()

	repo := store.Timeline()
	ctx := context.Background()

	// FK actor
	store.Actors().Save(ctx, &domain.Actor{ID: "a1", Kind: "person", DisplayName: "Alice", CreatedAt: time.Now()})

	t.Run("Save and Fetch", func(t *testing.T) {
		now := time.Now().Truncate(time.Second)
		e := &domain.Event{
			ID:         "e1",
			Kind:       domain.EventMeeting,
			Title:      "Meeting 1",
			OccurredAt: &now,
			RecordedAt: now,
			CreatedBy:  "a1",
		}
		if err := repo.Save(ctx, e); err != nil {
			t.Fatalf("failed to save event: %v", err)
		}

		events, err := repo.Fetch(ctx, &now, nil, 10)
		if err != nil {
			t.Fatalf("failed to fetch events: %v", err)
		}
		if len(events) == 0 || events[0].ID != "e1" {
			t.Errorf("expected event e1, got %v", events)
		}
	})

	t.Run("Fetch Error", func(t *testing.T) {
		store.Close()
		_, err := repo.Fetch(ctx, nil, nil, 10)
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
