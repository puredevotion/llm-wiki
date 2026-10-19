package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/controllers/rest"
	"llm-wiki/apps/backend/internal/services"
	"llm-wiki/apps/backend/internal/clients/embeddings"
	"llm-wiki/apps/backend/internal/storage/graph"
	"llm-wiki/apps/backend/internal/storage/turso"
)

func main() {
	if err := run(); err != nil {
		slog.Error("application failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.FromEnv()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	// Auto-migrate Turso
	schema, err := os.ReadFile(filepath.Join(migrationsPath, "000001_initial.sql"))
	if err == nil {
		if err := sqlStore.Migrate(ctx, string(schema)); err != nil {
			logger.Warn("turso migration failed", "error", err)
		} else {
			logger.Info("turso migrated")
		}
	}

	// 2. Initialize Graph
	graphStore, err := graph.NewStore(cfg.GraphDBPath)
	if err != nil {
		return err
	}
	defer graphStore.Close()

	// Auto-migrate Graph
	if err := graphStore.Migrate(ctx); err != nil {
		logger.Warn("graph migration failed", "error", err)
	} else {
		logger.Info("graph migrated")
	}

	// 3. Initialize Services
	embeds := embeddings.NewOpenAIClient(cfg.OpenAIAPIKey)
	
	ingestion := services.NewIngestionService(
		sqlStore.Actors(),
		sqlStore.Sources(),
		sqlStore.Zettels(),
		sqlStore.Topics(),
		graphStore.Graph(),
		sqlStore.Operations(),
		sqlStore.Vectors(),
		embeds,
	)
	searchSvc := services.NewSearchService(sqlStore.Zettels(), sqlStore.Vectors(), embeds)
	syncSvc := services.NewSyncService(sqlStore.Operations(), sqlStore.Zettels(), sqlStore.Topics(), sqlStore.Vectors(), embeds)
	idSvc := services.NewIdentityService(sqlStore.Actors(), sqlStore.Identity(), graphStore.Graph(), sqlStore.Operations())
	timeSvc := services.NewTimelineService(sqlStore.Timeline(), graphStore.Graph(), sqlStore.Operations())
	graphSvc := services.NewGraphService(graphStore.Graph())

	server := rest.NewServer(cfg, logger, ingestion, searchSvc, syncSvc, idSvc, timeSvc, graphSvc)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("backend listening", "addr", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("backend stopped unexpectedly", "error", err)
			stop()
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	return httpServer.Shutdown(shutdownCtx)
}
