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
