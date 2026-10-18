package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
	"llm-wiki/apps/backend/internal/storage/graph"
	"llm-wiki/apps/backend/internal/storage/turso"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("bootstrap failed: %v", err)
	}
}

func run() error {
	cfg := config.FromEnv()
	ctx := context.Background()

	log.Printf("Bootstrapping knowledge base...")

	// 1. Initialize Turso
	sqlStore, err := turso.NewStore(cfg.TursoDSN)
	if err != nil {
		return err
	}
	defer sqlStore.Close()

	migrationsPath := os.Getenv("KBASE_MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}
	schema, err := os.ReadFile(filepath.Join(migrationsPath, "000001_initial.sql"))
	if err != nil {
		return err
	}
	if err := sqlStore.Migrate(ctx, string(schema)); err != nil {
		return err
	}

	// 2. Initialize Graph
	graphStore, err := graph.NewStore(cfg.GraphDBPath)
	if err != nil {
		return err
	}
	defer graphStore.Close()

	if err := graphStore.Migrate(ctx); err != nil {
		return err
	}

	// 3. Seed Data
	actors := sqlStore.Actors()
	topics := sqlStore.Topics()
	gRepo := graphStore.Graph()

	// Initialize ingestion with ops for consistency, though not used here
	ingestion := services.NewIngestionService(actors, sqlStore.Sources(), sqlStore.Zettels(), topics, gRepo, sqlStore.Operations(), sqlStore.Vectors(), nil)
	_ = ingestion // Placeholder if we need to call it

	system := &domain.Actor{
		ID:          "actor_system",
		Kind:        "service",
		DisplayName: "System",
		CreatedAt:   time.Now(),
	}
	if err := actors.Save(ctx, system); err != nil {
		return err
	}
	if err := gRepo.UpsertNode(ctx, system.ID, "Person", map[string]any{"name": system.DisplayName}); err != nil {
		return err
	}

	initialTopics := []string{"Architecture", "Development", "Research", "Meetings"}
	for _, tName := range initialTopics {
		topic := &domain.Topic{
			ID:        fmt.Sprintf("top_%s", strings.ToLower(tName)),
			Name:      tName,
			CreatedAt: time.Now(),
		}
		if err := topics.Save(ctx, topic); err != nil {
			return err
		}
		if err := gRepo.UpsertNode(ctx, topic.ID, "Topic", map[string]any{"name": topic.Name}); err != nil {
			return err
		}
	}

	log.Printf("🎉 Knowledge base bootstrap complete!")
	return nil
}
