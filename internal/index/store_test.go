package index

import (
	"path/filepath"
	"testing"
)

func tempDB(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test-index.db")
	s, err := NewStoreAt(dbPath)
	if err != nil {
		t.Fatalf("NewStoreAt: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestStore_ProjectPinning(t *testing.T) {
	s := tempDB(t)

	pinned, err := s.IsProjectPinned("/path/to/project")
	if err != nil {
		t.Fatalf("IsProjectPinned error: %v", err)
	}
	if pinned {
		t.Error("expected project to not be pinned initially")
	}

	if err := s.PinProject("/path/to/project"); err != nil {
		t.Fatalf("PinProject error: %v", err)
	}

	pinned, err = s.IsProjectPinned("/path/to/project")
	if err != nil {
		t.Fatalf("IsProjectPinned error: %v", err)
	}
	if !pinned {
		t.Error("expected project to be pinned")
	}

	all, err := s.GetPinnedProjects()
	if err != nil {
		t.Fatalf("GetPinnedProjects error: %v", err)
	}
	if all["/path/to/project"] == 0 {
		t.Error("expected path to be in pinned projects map")
	}

	if err := s.PinProject("/path/to/project1"); err != nil {
		t.Fatalf("PinProject error: %v", err)
	}
	if err := s.PinProject("/path/to/project2"); err != nil {
		t.Fatalf("PinProject error: %v", err)
	}

	all, err = s.GetPinnedProjects()
	if err != nil {
		t.Fatalf("GetPinnedProjects error: %v", err)
	}
	// Most recently pinned project (/path/to/project2) should have order 1
	if all["/path/to/project2"] >= all["/path/to/project1"] {
		t.Errorf("expected project2 rank (%d) to be lower (higher priority) than project1 rank (%d)", all["/path/to/project2"], all["/path/to/project1"])
	}

	if err := s.UnpinProject("/path/to/project"); err != nil {
		t.Fatalf("UnpinProject error: %v", err)
	}

	pinned, err = s.IsProjectPinned("/path/to/project")
	if err != nil {
		t.Fatalf("IsProjectPinned error: %v", err)
	}
	if pinned {
		t.Error("expected project to be unpinned")
	}
}

func TestStore_TmuxAliases(t *testing.T) {
	s := tempDB(t)

	aliases, err := s.GetTmuxAliases()
	if err != nil {
		t.Fatalf("GetTmuxAliases error: %v", err)
	}
	if len(aliases) != 0 {
		t.Errorf("expected 0 aliases initially, got %d", len(aliases))
	}

	path := "/path/to/myproj"
	if err := s.SetTmuxAlias(path, "my-alias"); err != nil {
		t.Fatalf("SetTmuxAlias error: %v", err)
	}

	aliases, err = s.GetTmuxAliases()
	if err != nil {
		t.Fatalf("GetTmuxAliases error: %v", err)
	}
	if got := aliases[path]; got != "my-alias" {
		t.Errorf("expected alias %q, got %q", "my-alias", got)
	}

	// Update alias
	if err := s.SetTmuxAlias(path, "updated-alias"); err != nil {
		t.Fatalf("SetTmuxAlias update error: %v", err)
	}
	aliases, _ = s.GetTmuxAliases()
	if got := aliases[path]; got != "updated-alias" {
		t.Errorf("expected updated alias %q, got %q", "updated-alias", got)
	}

	// Clear alias
	if err := s.SetTmuxAlias(path, ""); err != nil {
		t.Fatalf("SetTmuxAlias clear error: %v", err)
	}
	aliases, _ = s.GetTmuxAliases()
	if _, exists := aliases[path]; exists {
		t.Error("expected alias to be removed when cleared")
	}
}
