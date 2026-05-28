package mcp

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"llm-wiki/apps/backend/internal/services"
)

func TestIngestToolHandler(t *testing.T) {
	agentToken := "valid-token"
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ingestion := services.NewIngestionService(
		&mockActorRepo{},
		&mockSourceRepo{},
		&mockZettelRepo{},
		&mockTopicRepo{},
		&mockGraphRepo{},
		&mockOpRepo{},
		&mockVectorRepo{},
		&mockEmbeddingsClient{},
	)
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
