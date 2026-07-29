package todo

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/project-hub/internal/domain"
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

func TestScanProjectsDiscovery(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "todo-discover-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create subdirectories
	projAPath := filepath.Join(tempDir, "projA")
	projBPath := filepath.Join(tempDir, "projB")
	ignoredPath := filepath.Join(tempDir, "node_modules") // will be ignored
	nonGitPath := filepath.Join(tempDir, "nonGit") // should not be detected as a project (will fallback if no projects discovered, but here there are other projects discovered)

	for _, p := range []string{projAPath, projBPath, ignoredPath, nonGitPath} {
		if err := os.Mkdir(p, 0755); err != nil {
			t.Fatalf("failed to create dir: %v", err)
		}
	}

	// Create .git folders in projA, projB, and node_modules
	for _, p := range []string{projAPath, projBPath, ignoredPath} {
		gitPath := filepath.Join(p, ".git")
		if err := os.Mkdir(gitPath, 0755); err != nil {
			t.Fatalf("failed to create git folder: %v", err)
		}
	}

	cfg := domain.DefaultConfig()
	// IgnoreFolders defaults to containing "node_modules"
	cfg.IndexFolders = []string{filepath.Join(tempDir, "*")}

	scanner := NewScanner(cfg)
	projects, err := scanner.ScanProjects(tempDir)
	if err != nil {
		t.Fatalf("ScanProjects failed: %v", err)
	}

	// Should discover projA and projB, skip node_modules, and ignore nonGit (no .git)
	if len(projects) != 2 {
		t.Errorf("Expected 2 discovered projects, got %d", len(projects))
		for _, p := range projects {
			t.Logf("discovered: %s", p.Path)
		}
		return
	}

	discoveredA := false
	discoveredB := false
	for _, p := range projects {
		if p.Name == "projA" {
			discoveredA = true
		}
		if p.Name == "projB" {
			discoveredB = true
		}
	}

	if !discoveredA || !discoveredB {
		t.Errorf("Expected projA and projB to be discovered")
	}
}

func TestParseFTSDetailedAndFindTodos(t *testing.T) {
	cfg := domain.DefaultConfig()

	tempDir := t.TempDir()

	// Test FTS detailed parsing
	ftsContent := `
# Title

## Feature A
  - [x] Task 1
  - [ ] Task 2

### Feature B
  - [x] Task 3
`
	ftsPath := filepath.Join(tempDir, "tasks.md")
	_ = os.WriteFile(ftsPath, []byte(ftsContent), 0644)

	completed, total, features, err := ParseFTSDetailed(ftsPath, cfg)
	if err != nil {
		t.Fatalf("ParseFTSDetailed failed: %v", err)
	}
	if total != 3 || completed != 2 {
		t.Errorf("Expected 3 tasks (2 completed), got %d (%d completed)", total, completed)
	}
	if len(features) != 2 {
		t.Errorf("Expected 2 features, got %d", len(features))
	}
	if features[0].Name != "A" || len(features[0].Tasks) != 2 {
		t.Errorf("Expected Feature A (name 'A') with 2 tasks, got name=%q, tasks=%d", features[0].Name, len(features[0].Tasks))
	}
	if features[1].Name != "B" || len(features[1].Tasks) != 1 {
		t.Errorf("Expected Feature B (name 'B') with 1 task, got name=%q, tasks=%d", features[1].Name, len(features[1].Tasks))
	}

	// Test Todo scanning
	srcContent := `
// TODO: task 1
// FIXME: task 2
`
	srcPath := filepath.Join(tempDir, "main.go")
	_ = os.WriteFile(srcPath, []byte(srcContent), 0644)

	todos, err := FindTodosInFile(srcPath, cfg)
	if err != nil {
		t.Fatalf("FindTodosInFile failed: %v", err)
	}
	if len(todos) != 2 {
		t.Errorf("Expected 2 todos, got %d", len(todos))
	}
	if todos[0].Text != "// TODO: task 1" || todos[1].Text != "// FIXME: task 2" {
		t.Errorf("Unexpected todos text: %+v", todos)
	}
}

func TestRecurringTasks(t *testing.T) {
	cfg := domain.DefaultConfig()
	cfg.DailyReset = "04:20"
	cfg.WeeklyReset = "monday"

	tempDir := t.TempDir()
	dailyFile := filepath.Join(tempDir, "daily.md")

	content := `## Daily Habits
- [x] Exercise
- [ ] Read book
`
	if err := os.WriteFile(dailyFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write daily.md: %v", err)
	}

	// 1. First run should insert metadata and perform initial reset if checked
	if err := CheckAndResetRecurringTasks(dailyFile, "daily", cfg); err != nil {
		t.Fatalf("CheckAndResetRecurringTasks failed: %v", err)
	}

	data, _ := os.ReadFile(dailyFile)
	if !strings.Contains(string(data), "<!-- last_reset:") {
		t.Errorf("expected metadata comment in daily.md, got:\n%s", string(data))
	}

	// 2. Test toggling checkbox on line 2 (Exercise)
	if err := ToggleTaskInFile(dailyFile, 2); err != nil {
		t.Fatalf("ToggleTaskInFile failed: %v", err)
	}

	data, _ = os.ReadFile(dailyFile)
	if !strings.Contains(string(data), "- [x] Exercise") {
		t.Errorf("expected Exercise to be toggled to [x], got:\n%s", string(data))
	}

	// 3. Test ReadYAMLName
	yamlContent := `---
name: My Custom Daily Routine
---
## Tasks
- [ ] Task 1
`
	if name := ReadYAMLName(yamlContent); name != "My Custom Daily Routine" {
		t.Errorf("expected 'My Custom Daily Routine', got %q", name)
	}
}

