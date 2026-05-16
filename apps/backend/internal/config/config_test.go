package config

import "testing"

func TestFromEnvUsesDefaultHTTPAddressWhenEnvironmentIsUnset(t *testing.T) {
	t.Setenv("KBASE_HTTP_ADDR", "")

	cfg := FromEnv()

	if cfg.HTTPAddr != ":8080" {
		t.Fatalf("expected default HTTP address :8080, got %q", cfg.HTTPAddr)
	}
}

func TestFromEnvUsesConfiguredHTTPAddressForDeployments(t *testing.T) {
	t.Setenv("KBASE_HTTP_ADDR", "127.0.0.1:9000")

	cfg := FromEnv()

	if cfg.HTTPAddr != "127.0.0.1:9000" {
		t.Fatalf("expected configured HTTP address, got %q", cfg.HTTPAddr)
	}
}
