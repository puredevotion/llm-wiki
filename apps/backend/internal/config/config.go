package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// Config contains process-level settings.
type Config struct {
	AppEnv       string
	HTTPAddr     string
	AgentToken   string
	TursoDSN     string
	GraphDBPath  string
	OpenAIAPIKey string
}

func FromEnv() Config {
	// 1. Determine environment
	appEnv := os.Getenv("APP_ENV")
	if appEnv == "" {
		appEnv = "local"
	}

	// 2. Load corresponding .env file
	// We try to load .env.{env} first, then fallback to .env
	envFile := ".env." + appEnv
	if err := godotenv.Load(envFile); err != nil {
		// It's okay if the specific env file is missing, we'll try the default .env
		if err := godotenv.Load(); err != nil {
			log.Printf("No .env file found, relying on environment variables")
		}
	} else {
		log.Printf("Loaded config from %s", envFile)
	}

	return Config{
		AppEnv:       appEnv,
		HTTPAddr:     envOrDefault("KBASE_HTTP_ADDR", ":8080"),
		AgentToken:   envOrDefault("KBASE_AGENT_TOKEN", "default-agent-token"),
		TursoDSN:     envOrDefault("KBASE_TURSO_DSN", "file:kbase.db"),
		GraphDBPath:  envOrDefault("KBASE_GRAPH_DB_PATH", "kbase_graph"),
		OpenAIAPIKey: os.Getenv("OPENAI_API_KEY"),
	}
}

func envOrDefault(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
