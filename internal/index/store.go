package index

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"

	"github.com/super-todo/internal/domain"
	"github.com/super-todo/internal/utils"
)

// Store provides SQLite-backed project cache and pinning persistence.
type Store struct {
	db     *sql.DB
	dbPath string
}

// NewStore opens (or creates) the SQLite index at the default XDG data path.
func NewStore() (*Store, error) {
	dbPath := DefaultDBPath()
	return NewStoreAt(dbPath)
}

// NewStoreAt opens (or creates) the SQLite index at a specific path.
func NewStoreAt(dbPath string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("index: create data dir: %w", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("index: open db: %w", err)
	}

	// Enable WAL mode for concurrent reads during background refresh.
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("index: set WAL: %w", err)
	}

	s := &Store{db: db, dbPath: dbPath}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, err
	}
	return s, nil
}

// DefaultDBPath returns the standard location for the index database.
func DefaultDBPath() string {
	return filepath.Join(utils.XDGDataHome(), "super-todo", "index.db")
}

// Close releases the database connection.
func (s *Store) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate creates or updates the schema to the latest version.
func (s *Store) migrate() error {
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS projects_cache (
			id         INTEGER PRIMARY KEY CHECK (id = 1),
			data       TEXT    NOT NULL,
			scanned_at TEXT    NOT NULL
		);

		CREATE TABLE IF NOT EXISTS pinned_projects (
			path       TEXT PRIMARY KEY,
			pinned_at  TEXT NOT NULL DEFAULT ''
		);
	`)
	if err != nil {
		return fmt.Errorf("index: migrate: %w", err)
	}

	// Ensure pinned_at column exists for existing DBs
	_, _ = s.db.Exec("ALTER TABLE pinned_projects ADD COLUMN pinned_at TEXT NOT NULL DEFAULT ''")

	return nil
}

// SaveProjectsCache saves the scanned projects and current timestamp in SQLite database.
func (s *Store) SaveProjectsCache(projects []domain.Project) error {
	data, err := json.Marshal(projects)
	if err != nil {
		return fmt.Errorf("index: marshal projects: %w", err)
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`
		INSERT INTO projects_cache (id, data, scanned_at)
		VALUES (1, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			data       = excluded.data,
			scanned_at = excluded.scanned_at
	`, string(data), now)
	if err != nil {
		return fmt.Errorf("index: save projects cache: %w", err)
	}
	return nil
}

// LoadProjectsCache loads the cached projects and the last scanned timestamp.
func (s *Store) LoadProjectsCache() ([]domain.Project, time.Time, error) {
	var data string
	var scannedAt string
	err := s.db.QueryRow("SELECT data, scanned_at FROM projects_cache WHERE id = 1").Scan(&data, &scannedAt)
	if err != nil {
		return nil, time.Time{}, err
	}
	var projects []domain.Project
	if err := json.Unmarshal([]byte(data), &projects); err != nil {
		return nil, time.Time{}, fmt.Errorf("index: unmarshal projects: %w", err)
	}
	t, err := time.Parse(time.RFC3339, scannedAt)
	if err != nil {
		return projects, time.Time{}, nil
	}
	return projects, t, nil
}

// PinProject records a project path as pinned with current timestamp.
func (s *Store) PinProject(path string) error {
	now := time.Now().UTC().Format(time.RFC3339Nano)
	_, err := s.db.Exec("INSERT INTO pinned_projects (path, pinned_at) VALUES (?, ?) ON CONFLICT(path) DO UPDATE SET pinned_at = excluded.pinned_at", path, now)
	if err != nil {
		return fmt.Errorf("index: pin project: %w", err)
	}
	return nil
}

// UnpinProject removes a project path from pinned list.
func (s *Store) UnpinProject(path string) error {
	_, err := s.db.Exec("DELETE FROM pinned_projects WHERE path = ?", path)
	if err != nil {
		return fmt.Errorf("index: unpin project: %w", err)
	}
	return nil
}

// IsProjectPinned checks if a project path is pinned.
func (s *Store) IsProjectPinned(path string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM pinned_projects WHERE path = ?", path).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("index: check pinned project: %w", err)
	}
	return count > 0, nil
}

// GetPinnedProjects returns a map of project path to pin order (1-indexed based on pinned_at timestamp, most recent first).
func (s *Store) GetPinnedProjects() (map[string]int, error) {
	rows, err := s.db.Query("SELECT path FROM pinned_projects ORDER BY rowid DESC")
	if err != nil {
		return nil, fmt.Errorf("index: get pinned projects: %w", err)
	}
	defer rows.Close()

	pinned := make(map[string]int)
	order := 1
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err == nil {
			pinned[path] = order
			order++
		}
	}
	return pinned, nil
}
