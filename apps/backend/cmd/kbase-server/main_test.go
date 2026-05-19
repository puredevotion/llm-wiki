package main

import (
	"os"
	"syscall"
	"testing"
	"time"
)

func TestRun(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "server-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	os.Setenv("KBASE_TURSO_DSN", "file:"+tmpDir+"/t.db")
	os.Setenv("KBASE_GRAPH_DB_PATH", tmpDir+"/g")
	os.Setenv("KBASE_HTTP_ADDR", ":0")
	os.Setenv("KBASE_MIGRATIONS_PATH", "../../migrations")

	// We need to signal the process to stop after some time
	go func() {
		time.Sleep(1 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGTERM)
	}()

	err = run()
	if err != nil {
		t.Fatalf("run() failed: %v", err)
	}
}
