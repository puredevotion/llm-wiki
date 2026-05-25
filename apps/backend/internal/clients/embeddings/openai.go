package embeddings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"llm-wiki/apps/backend/internal/domain"
)

type OpenAIClient struct {
	apiKey string
	model  string
	url    string
	client *http.Client
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		apiKey: apiKey,
		model:  "text-embedding-3-small",
		url:    "https://api.openai.com/v1/embeddings",
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *OpenAIClient) Generate(ctx context.Context, text string) (domain.Vector, error) {
	vecs, err := c.BatchGenerate(ctx, []string{text})
	if err != nil {
		return nil, err
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}
	return vecs[0], nil
}

func (c *OpenAIClient) BatchGenerate(ctx context.Context, texts []string) ([]domain.Vector, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("openai api key is required")
	}

	reqBody, _ := json.Marshal(map[string]any{
		"input": texts,
		"model": c.model,
	})

	req, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return nil, fmt.Errorf("openai error (%d): %s", resp.StatusCode, errResp.Error.Message)
	}

	var result struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	vecs := make([]domain.Vector, 0, len(result.Data))
	for _, d := range result.Data {
		vecs = append(vecs, domain.Vector(d.Embedding))
	}
	return vecs, nil
}
