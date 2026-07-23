package todo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/super-todo/internal/domain"
)

// Scanner handles finding projects and scanning their tasks and todos.
type Scanner struct {
	config domain.Config
}

// NewScanner creates a new Scanner instance.
func NewScanner(cfg domain.Config) *Scanner {
	return &Scanner{config: cfg}
}

// ScanProjects finds and scans all projects according to config.
func (s *Scanner) ScanProjects(cwd string) ([]Project, error) {
	var projectDirs []string

	// 1. Resolve target project directories from IndexFolders.
	if len(s.config.IndexFolders) == 0 {
		// Default to scanning the current directory as a single project
		projectDirs = append(projectDirs, cwd)
	} else {
		visited := make(map[string]bool)
		for _, pattern := range s.config.IndexFolders {
			expanded := expandPath(pattern)
			matches, err := filepath.Glob(expanded)
			if err != nil {
				continue
			}
			for _, match := range matches {
				info, err := os.Stat(match)
				if err != nil {
					continue
				}
				if info.IsDir() {
					absPath, err := filepath.Abs(match)
					if err == nil && !visited[absPath] {
						visited[absPath] = true
						projectDirs = append(projectDirs, absPath)
					}
				}
			}
		}
		// If nothing matched, fallback to current directory
		if len(projectDirs) == 0 {
			projectDirs = append(projectDirs, cwd)
		}
	}

	// 2. Scan each resolved project directory.
	var projects []Project
	for _, dir := range projectDirs {
		proj, err := s.ScanProject(dir)
		if err == nil {
			projects = append(projects, proj)
		}
	}

	return projects, nil
}

// ScanProject scans a single project directory recursively.
func (s *Scanner) ScanProject(dir string) (Project, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		absDir = dir
	}

	name := filepath.Base(absDir)

	proj := Project{
		Name: name,
		Path: absDir,
	}

	// Map of ignore folder names for fast lookup
	ignoreMap := make(map[string]bool)
	for _, folder := range s.config.IgnoreFolders {
		ignoreMap[strings.ToLower(folder)] = true
	}

	// Map of FTS files to check
	ftsMap := make(map[string]bool)
	for _, ftsFile := range s.config.FtsFiles {
		ftsMap[strings.ToLower(ftsFile)] = true
	}

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // skip errors
		}

		// Calculate relative path to project root
		relPath, err := filepath.Rel(absDir, path)
		if err != nil {
			return nil
		}

		// Skip ignored folders
		if info.IsDir() {
			dirName := strings.ToLower(info.Name())
			if ignoreMap[dirName] {
				return filepath.SkipDir
			}
			return nil
		}

		relPathLower := strings.ToLower(relPath)

		// Check if it is an FTS task file
		// Check direct name match (e.g. plan/tasks.md or tasks.md)
		isFTS := false
		if ftsMap[relPathLower] {
			isFTS = true
		} else {
			// Also check filename base match in case FtsFiles contains relative files
			baseLower := strings.ToLower(filepath.Base(path))
			if ftsMap[baseLower] {
				isFTS = true
			}
		}

		if isFTS {
			completed, total, err := ParseFTSFile(path, s.config)
			if err == nil {
				proj.CompletedTasks += completed
				proj.TotalTasks += total
			}
		} else {
			// Scan code file for TODOs
			todos, err := CountTodosInFile(path, s.config)
			if err == nil {
				proj.TodoCount += todos
			}
		}

		return nil
	})

	return proj, err
}

func expandPath(value string) string {
	expanded := os.ExpandEnv(value)
	if expanded == "" {
		return expanded
	}
	if expanded == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			return home
		}
		return expanded
	}
	if strings.HasPrefix(expanded, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(expanded, "~/"))
		}
	}
	return expanded
}
