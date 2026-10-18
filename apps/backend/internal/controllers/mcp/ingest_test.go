package mcp

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/domain"
	"llm-wiki/apps/backend/internal/services"
)

type mockActorRepo struct{}

func (m *mockActorRepo) FindByName(ctx context.Context, name string) (*domain.Actor, error) {
	if name == "Alice" {
		return &domain.Actor{ID: "a1", DisplayName: "Alice"}, nil
	}
	return nil, nil
}
func (m *mockActorRepo) Save(ctx context.Context, actor *domain.Actor) error { return nil }

type mockSourceRepo struct{}

func (m *mockSourceRepo) Save(ctx context.Context, source *domain.Source) error { return nil }

type mockZettelRepo struct{}

func (m *mockZettelRepo) Save(ctx context.Context, zettel *domain.Zettel) error { return nil }
func (m *mockZettelRepo) SearchZettels(ctx context.Context, query string, limit int) ([]*domain.Zettel, error) {
	return []*domain.Zettel{}, nil
}

type mockTopicRepo struct{}

func (m *mockTopicRepo) FindByName(ctx context.Context, name string) (*domain.Topic, error) {
	return nil, nil
}
func (m *mockTopicRepo) Save(ctx context.Context, topic *domain.Topic) error { return nil }

type mockGraphRepo struct{}

func (m *mockGraphRepo) UpsertNode(ctx context.Context, id, label string, properties map[string]any) error {
	return nil
}
func (m *mockGraphRepo) CreateRelationship(ctx context.Context, fromID, fromLabel, toID, toLabel, relType string) error {
	return nil
}

type mockOpRepo struct{}

func (m *mockOpRepo) Save(ctx context.Context, op *domain.Operation) error { return nil }
func (m *mockOpRepo) FindByID(ctx context.Context, id string) (*domain.Operation, error) {
	return nil, nil
}
func (m *mockOpRepo) FetchChanges(ctx context.Context, cursor string, limit int) ([]*domain.Operation, error) {
	return nil, nil
}

func TestIngestToolHandler(t *testing.T) {
	agentToken := "valid-token"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ingestion := services.NewIngestionService(&mockActorRepo{}, &mockSourceRepo{}, &mockZettelRepo{}, &mockTopicRepo{}, &mockGraphRepo{}, &mockOpRepo{})
	handler := ingestToolHandler(logger, ingestion, agentToken)

	t.Run("Valid Full Ingestion", func(t *testing.T) {
		input := IngestInput{
			Token:        "valid-token",
			Project:      "Test Project",
			Summary:      "This is a test summary.",
			Conclusions:  []string{"Conclusion 1"},
			Participants: []string{"Alice", "Bob"},
			Topics:       []string{"Go", "Testing"},
			Timestamp:    time.Now().Format(time.RFC3339),
		}
		res, output, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.Status != "created" {
			t.Errorf("expected status created, got %s", output.Status)
		}
		if res == nil {
			t.Error("expected non-nil result")
		}
	})

	t.Run("Invalid Token", func(t *testing.T) {
		input := IngestInput{Token: "wrong-token"}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil || err.Error() != "unauthorized: invalid agent token" {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("Service Error", func(t *testing.T) {
		input := IngestInput{
			Token:       "valid-token",
			Summary:     "", // Service requires non-empty summary
			Conclusions: []string{"c"},
		}
		_, _, err := handler(context.Background(), nil, input)
		if err == nil {
			t.Error("expected ingestion error, got nil")
		}
	})

	t.Run("Invalid Timestamp Fallback", func(t *testing.T) {
		input := IngestInput{
			Token:       "valid-token",
			Summary:     "s",
			Conclusions: []string{"c"},
			Timestamp:   "invalid",
		}
		_, _, err := handler(context.Background(), nil, input)
		if err != nil {
			t.Errorf("expected success with fallback timestamp, got %v", err)
		}
	})
}
