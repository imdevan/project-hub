package todo

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/super-todo/internal/domain"
)

func TestParseFTSContent(t *testing.T) {
	cfg := domain.DefaultConfig()

	content := `
# Context

# Definitions

# v0.1.0

## Feature 1: CLI
  - [x] 1.1 First task
    - notes: some notes
    - [ ] 1.1.1 Sub task 1
    - [x] 1.1.2 Sub task 2
`

	completed, total := ParseFTSContent(content, cfg)
	if total != 3 {
		t.Errorf("Expected 3 tasks, got %d", total)
	}
	if completed != 2 {
		t.Errorf("Expected 2 completed tasks, got %d", completed)
	}
}

func TestScanner(t *testing.T) {
	cfg := domain.DefaultConfig()
	cfg.FtsFiles = []string{"tasks.md"}

	// Create temp dir structure for testing
	tempDir, err := os.MkdirTemp("", "todo-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a mock FTS file
	ftsContent := `
## Feature 1
  - [x] 1.1 Task A
  - [ ] 1.2 Task B
`
	ftsPath := filepath.Join(tempDir, "tasks.md")
	if err := os.WriteFile(ftsPath, []byte(ftsContent), 0644); err != nil {
		t.Fatalf("Failed to write mock FTS: %v", err)
	}

	// Create a mock source code file with TODOs
	srcContent := `
// TODO: clean up code
func main() {
    // FIXME: fix error
    println("hello")
}
`
	srcPath := filepath.Join(tempDir, "main.go")
	if err := os.WriteFile(srcPath, []byte(srcContent), 0644); err != nil {
		t.Fatalf("Failed to write mock source: %v", err)
	}

	// Create an ignored dir with TODOs (should be ignored)
	ignoredDir := filepath.Join(tempDir, "node_modules")
	if err := os.Mkdir(ignoredDir, 0755); err != nil {
		t.Fatalf("Failed to create ignored dir: %v", err)
	}
	ignoredSrc := filepath.Join(ignoredDir, "library.js")
	if err := os.WriteFile(ignoredSrc, []byte("// TODO: ignore me"), 0644); err != nil {
		t.Fatalf("Failed to write ignored source: %v", err)
	}

	scanner := NewScanner(cfg)
	proj, err := scanner.ScanProject(tempDir)
	if err != nil {
		t.Fatalf("Failed to scan project: %v", err)
	}

	if proj.TotalTasks != 2 {
		t.Errorf("Expected 2 total tasks, got %d", proj.TotalTasks)
	}
	if proj.CompletedTasks != 1 {
		t.Errorf("Expected 1 completed task, got %d", proj.CompletedTasks)
	}
	if proj.TodoCount != 2 {
		t.Errorf("Expected 2 code TODOs, got %d", proj.TodoCount)
	}
}
