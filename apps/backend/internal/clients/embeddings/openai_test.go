package embeddings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAIClient(t *testing.T) {
	t.Run("Generate Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data": [{"embedding": [0.1, 0.2]}]}`))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-key")
		client.url = ts.URL
		
		vec, err := client.Generate(context.Background(), "hello")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(vec) != 2 || vec[0] != 0.1 {
			t.Errorf("expected [0.1, 0.2], got %v", vec)
		}
	})

	t.Run("Generate Empty Data", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"data": []}`))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-key")
		client.url = ts.URL
		
		_, err := client.Generate(context.Background(), "hello")
		if err == nil {
			t.Error("expected error for empty data")
		}
	})

	t.Run("Error Response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(`{"error": {"message": "invalid request"}}`))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-key")
		client.url = ts.URL
		
		_, err := client.Generate(context.Background(), "hello")
		if err == nil {
			t.Error("expected error, got nil")
		}
	})
	
	t.Run("Invalid JSON Response", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{invalid}`))
		}))
		defer ts.Close()

		client := NewOpenAIClient("test-key")
		client.url = ts.URL
		
		_, err := client.Generate(context.Background(), "hello")
		if err == nil {
			t.Error("expected decode error")
		}
	})

	t.Run("Invalid Request URL", func(t *testing.T) {
		client := NewOpenAIClient("test-key")
		client.url = ":" // Invalid URL to trigger NewRequest error
		_, err := client.Generate(context.Background(), "hello")
		if err == nil {
			t.Error("expected request error")
		}
	})

	t.Run("Missing API Key", func(t *testing.T) {
		client := NewOpenAIClient("")
		_, err := client.Generate(context.Background(), "hello")
		if err == nil {
			t.Error("expected error for missing API key")
		}
	})
}
