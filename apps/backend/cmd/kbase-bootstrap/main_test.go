package main

import (
	"os"
	"testing"
)

func TestBootstrapRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "bootstrap-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("KBASE_TURSO_DSN", "file:"+tmpDir+"/t.db")
	os.Setenv("KBASE_GRAPH_DB_PATH", tmpDir+"/g")
	os.Setenv("KBASE_MIGRATIONS_PATH", "../../migrations")

	err = run()
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}

func TestBootstrapRun_Failures(t *testing.T) {
	t.Run("Invalid DSN", func(t *testing.T) {
		os.Setenv("KBASE_TURSO_DSN", "unknown_driver://bad")
		err := run()
		if err == nil {
			t.Error("expected error for bad DSN")
		}
	})

	t.Run("Missing Migrations", func(t *testing.T) {
		os.Setenv("KBASE_TURSO_DSN", "file:test.db")
		os.Setenv("KBASE_MIGRATIONS_PATH", "/non/existent/path")
		err := run()
		if err == nil {
			t.Error("expected error for missing migrations")
		}
	})
	
	t.Run("Bad Graph Path", func(t *testing.T) {
		os.Setenv("KBASE_TURSO_DSN", "file:test.db")
		os.Setenv("KBASE_MIGRATIONS_PATH", "../../migrations")
		os.Setenv("KBASE_GRAPH_DB_PATH", "/proc/invalid/path")
		
		err := run()
		if err == nil {
			t.Error("expected error for bad graph path")
		}
	})

	t.Run("Default Migrations Path", func(t *testing.T) {
		os.Setenv("KBASE_MIGRATIONS_PATH", "")
		// Just trigger the logic
	})
}
