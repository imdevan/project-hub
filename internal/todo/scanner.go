package todo

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/project-hub/internal/domain"
	"github.com/project-hub/internal/index"
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
	projectDirs, err := s.ResolveProjectDirs(cwd)
	if err != nil {
		return nil, err
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


// ResolveProjectDirs resolves the project root directories according to config without scanning them.
func (s *Scanner) ResolveProjectDirs(cwd string) ([]string, error) {
	var projectDirs []string
	visited := make(map[string]bool)

	ignoreMap := make(map[string]bool)
	for _, folder := range s.config.IgnoreFolders {
		ignoreMap[strings.ToLower(folder)] = true
	}

	if len(s.config.IndexFolders) == 0 {
		discovered := s.discoverProjects(cwd)
		for _, dir := range discovered {
			absPath, err := filepath.Abs(dir)
			if err == nil && !visited[absPath] {
				visited[absPath] = true
				projectDirs = append(projectDirs, absPath)
			}
		}
	} else {
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
					dirName := strings.ToLower(filepath.Base(match))
					if ignoreMap[dirName] {
						continue
					}

					discovered := s.discoverProjects(match)
					for _, dir := range discovered {
						absPath, err := filepath.Abs(dir)
						if err == nil && !visited[absPath] {
							visited[absPath] = true
							projectDirs = append(projectDirs, absPath)
						}
					}
				}
			}
		}
	}

	// Fallback to cwd if no projects discovered
	if len(projectDirs) == 0 {
		absPath, err := filepath.Abs(cwd)
		if err == nil {
			projectDirs = append(projectDirs, absPath)
		}
	}

	return projectDirs, nil
}

// discoverProjects traverses rootDir recursively to find directories containing a .git subdirectory.
// It skips any directories named in config.IgnoreFolders.
func (s *Scanner) discoverProjects(rootDir string) []string {
	var projectDirs []string
	ignoreMap := make(map[string]bool)
	for _, folder := range s.config.IgnoreFolders {
		ignoreMap[strings.ToLower(folder)] = true
	}

	_ = filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			return nil
		}

		dirName := strings.ToLower(info.Name())
		if ignoreMap[dirName] {
			return filepath.SkipDir
		}

		// Check if this directory contains a .git folder
		gitPath := filepath.Join(path, ".git")
		gitInfo, err := os.Stat(gitPath)
		if err == nil && gitInfo.IsDir() {
			projectDirs = append(projectDirs, path)
			return filepath.SkipDir // Found project root, do not descend deeper
		}

		return nil
	})

	return projectDirs
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

	ignoredFilesMap := make(map[string]bool)
	if st, err := index.NewStore(); err == nil {
		ignoredFilesMap, _ = st.GetIgnoredTaskFiles()
		st.Close()
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
			completed, total, features, err := ParseFTSDetailed(path, s.config)
			if err == nil {
				isIgnored := ignoredFilesMap[path]
				if !isIgnored {
					proj.CompletedTasks += completed
					proj.TotalTasks += total
				}
				proj.FtsFiles = append(proj.FtsFiles, FtsFileDetail{
					Path:           relPath,
					CompletedTasks: completed,
					TotalTasks:     total,
					Features:       features,
					Ignored:        isIgnored,
				})
			}
		} else {
			// Scan code file for TODOs
			todos, err := FindTodosInFile(path, s.config)
			if err == nil {
				proj.TodoCount += len(todos)
				for i := range todos {
					todos[i].File = relPath
				}
				proj.Todos = append(proj.Todos, todos...)
			}
		}

		return nil
	})

	proj.Scanned = true
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
