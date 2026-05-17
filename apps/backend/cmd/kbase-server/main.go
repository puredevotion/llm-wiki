package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-wiki/apps/backend/internal/config"
	"llm-wiki/apps/backend/internal/controllers/rest"
	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

// In-memory repositories for bootstrapping until Turso/Limbo is ready
type memActorRepo struct {
	actors map[string]*domain.Actor
}

func (m *memActorRepo) FindByName(_ context.Context, name string) (*domain.Actor, error) {
	for _, a := range m.actors {
		if a.DisplayName == name {
			return a, nil
		}
	}
	return nil, nil
}
func (m *memActorRepo) Save(_ context.Context, actor *domain.Actor) error {
	m.actors[actor.ID] = actor
	return nil
}

type memSourceRepo struct{}

func (m *memSourceRepo) Save(_ context.Context, _ *domain.Source) error { return nil }

type memZettelRepo struct{}

func (m *memZettelRepo) Save(_ context.Context, _ *domain.Zettel) error { return nil }

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	cfg := config.FromEnv()

	// Bootstrap services with memory repos
	actorRepo := &memActorRepo{actors: make(map[string]*domain.Actor)}
	sourceRepo := &memSourceRepo{}
	zettelRepo := &memZettelRepo{}
	ingestion := services.NewIngestionService(actorRepo, sourceRepo, zettelRepo)

	server := rest.NewServer(cfg, logger, ingestion)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		logger.Error("backend shutdown failed", "error", err)
		os.Exit(1)
	}
}
