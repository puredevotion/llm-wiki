package embeddings

import (
	"context"
	"llm-wiki/apps/backend/internal/domain"
)

// Client defines the interface for generating text embeddings.
type Client interface {
	Generate(ctx context.Context, text string) (domain.Vector, error)
	BatchGenerate(ctx context.Context, texts []string) ([]domain.Vector, error)
}
