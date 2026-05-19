package config

import (
	"os"
	"testing"
)

func TestFromEnv(t *testing.T) {
	t.Run("Default Values", func(t *testing.T) {
		os.Setenv("APP_ENV", "")
		cfg := FromEnv()
		if cfg.AppEnv != "local" {
			t.Errorf("expected local, got %s", cfg.AppEnv)
		}
	})

	t.Run("Custom Environment", func(t *testing.T) {
		os.Setenv("APP_ENV", "staging")
		cfg := FromEnv()
		if cfg.AppEnv != "staging" {
			t.Errorf("expected staging, got %s", cfg.AppEnv)
		}
	})

	t.Run("Load Dotenv", func(t *testing.T) {
		// Create a temporary .env.test file
		os.WriteFile(".env.test", []byte("KBASE_HTTP_ADDR=:9999\n"), 0644)
		defer os.Remove(".env.test")

		os.Setenv("APP_ENV", "test")
		cfg := FromEnv()
		if cfg.HTTPAddr != ":9999" {
			t.Errorf("expected :9999 from .env.test, got %s", cfg.HTTPAddr)
		}
	})

	t.Run("Fallback to .env", func(t *testing.T) {
		os.WriteFile(".env", []byte("KBASE_AGENT_TOKEN=secret-token\n"), 0644)
		defer os.Remove(".env")
		
		os.Setenv("APP_ENV", "missing")
		cfg := FromEnv()
		if cfg.AgentToken != "secret-token" {
			t.Errorf("expected secret-token from .env, got %s", cfg.AgentToken)
		}
	})
}
