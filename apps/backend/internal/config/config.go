package config

import "os"

// Config contains process-level settings. Secrets should come from the environment.
type Config struct {
	HTTPAddr string
}

func FromEnv() Config {
	return Config{
		HTTPAddr: envOrDefault("KBASE_HTTP_ADDR", ":8080"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
