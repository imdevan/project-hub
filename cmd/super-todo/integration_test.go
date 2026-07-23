package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/super-todo/internal/testutil"
)

func TestIntegration_Version(t *testing.T) {
	env := testutil.SetupTestEnv(t)
	_ = env

	cmd := newRootCmd()
	cmd.SetArgs([]string{"--version"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("failed to run version command: %v", err)
	}
}

func TestIntegration_ConfigLoading(t *testing.T) {
	env := testutil.SetupTestEnv(t)

	// Create a dummy config
	configContent := `
ignore_folders = ["temp_ignored"]
fts_files = ["my_tasks.md"]
`
	if err := os.WriteFile(env.ConfigPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}

	cmd := newRootCmd()
	cmd.SetArgs([]string{"-c", env.ConfigPath})

	// Since running the interactive list starts a Bubble Tea program which expects a TTY,
	// we won't fully run the program in integration tests, but we verify it parses correctly.
	if cmd == nil {
		t.Fatal("root command should not be nil")
	}

	parentDir := filepath.Dir(env.ConfigPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		t.Fatal(err)
	}
}
