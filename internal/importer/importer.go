// Package importer clones or pulls the projects listed in a projects.lock file into a workspace.
package importer

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"

	"github.com/project-hub/internal/export"
)

// Action describes what happened to a project during import.
type Action string

const (
	Cloned   Action = "cloned"
	Pulled   Action = "pulled"
	UpToDate Action = "up to date"
	Ahead    Action = "ahead"
	Dirty    Action = "dirty"
	Skipped  Action = "skipped"
	Failed   Action = "failed"
)

// Actions lists every Action in display order.
var Actions = []Action{Cloned, Pulled, UpToDate, Ahead, Dirty, Skipped, Failed}

// Result is the outcome of importing a single project.
type Result struct {
	Project export.Project
	Dest    string
	Action  Action
	Detail  string
}

// FindLock looks for a lock file in the workspace, checking
// <workspace>/projects.lock then <workspace>/projects.lock/projects.lock.
func FindLock(workspace string) (string, error) {
	candidates := []string{
		filepath.Join(workspace, export.DefaultLockFile),
		filepath.Join(workspace, export.DefaultLockFile, export.DefaultLockFile),
	}
	for _, c := range candidates {
		if info, err := os.Stat(c); err == nil && info.Mode().IsRegular() {
			return c, nil
		}
	}
	return "", fmt.Errorf("no %s found (checked %s)", export.DefaultLockFile, strings.Join(candidates, ", "))
}

// Read decodes a lock file.
func Read(path string) (export.Lock, error) {
	var lock export.Lock
	data, err := os.ReadFile(path)
	if err != nil {
		return lock, err
	}
	if err := toml.Unmarshal(data, &lock); err != nil {
		return lock, fmt.Errorf("parse %s: %w", path, err)
	}
	return lock, nil
}

// Destinations maps each project in the lock to a directory inside workspace.
// Relative lock paths are joined to the workspace. Absolute (or ~) paths are made
// relative to the lock's recorded workspace when they live under it, otherwise to the
// common parent directory of all projects.
func Destinations(lock export.Lock, workspace string) []string {
	base := expand(lock.Workspace)
	var parents []string
	for _, p := range lock.Projects {
		if path := expand(p.Path); filepath.IsAbs(path) {
			parents = append(parents, filepath.Dir(path))
		}
	}
	common := commonDir(parents)

	dests := make([]string, len(lock.Projects))
	for i, p := range lock.Projects {
		path := expand(p.Path)
		if !filepath.IsAbs(path) {
			dests[i] = filepath.Join(workspace, path)
			continue
		}
		root := common
		if base != "" && within(path, base) {
			root = base
		}
		rel, err := filepath.Rel(root, path)
		if err != nil || rel == "." {
			rel = filepath.Base(path)
		}
		dests[i] = filepath.Join(workspace, rel)
	}
	return dests
}

// Import clones a project into dest, or pulls it if dest is already a git repository.
func Import(p export.Project, dest string) Result {
	res := Result{Project: p, Dest: dest}
	if info, err := os.Stat(dest); err == nil {
		if !info.IsDir() {
			return fail(res, "destination exists and is not a directory")
		}
		if _, err := os.Stat(filepath.Join(dest, ".git")); err != nil {
			return fail(res, "destination exists and is not a git repository")
		}
		pulled := Pull(dest)
		pulled.Project = p
		return pulled
	} else if !errors.Is(err, os.ErrNotExist) {
		return fail(res, err.Error())
	}

	primary, ok := primaryRemote(p.Remotes)
	if !ok {
		res.Action = Skipped
		res.Detail = "no remotes"
		return res
	}
	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return fail(res, err.Error())
	}
	if out, err := git("", "clone", "--origin", primary.Name, primary.URL, dest); err != nil {
		return fail(res, out)
	}
	res.Action = Cloned

	var warnings []string
	for _, r := range p.Remotes {
		if r.Name == primary.Name {
			continue
		}
		if out, err := git(dest, "remote", "add", r.Name, r.URL); err != nil {
			warnings = append(warnings, out)
		}
	}
	if p.Branch != "" && p.Branch != "HEAD" {
		current, _ := git(dest, "rev-parse", "--abbrev-ref", "HEAD")
		if current != p.Branch {
			if out, err := git(dest, "checkout", p.Branch); err != nil {
				warnings = append(warnings, "checkout "+p.Branch+": "+out)
			}
		}
	}
	res.Detail = strings.Join(warnings, "; ")
	return res
}

// Pull brings the git repository at dir up to date with its upstream, pulling only when needed:
//   - uncommitted changes to tracked files → Dirty, not pulled
//   - no upstream branch → Skipped
//   - after `git fetch`: same as upstream → UpToDate; only ahead → Ahead (nothing to pull)
//   - behind, or split from upstream (diverged) → plain `git pull`, so the user's git config
//     (pull.rebase, pull.ff, ...) decides how changes are integrated
func Pull(dir string) Result {
	res := Result{Project: export.Project{Path: dir}, Dest: dir}
	if out, err := git(dir, "status", "--porcelain", "--untracked-files=no"); err != nil {
		return fail(res, out)
	} else if out != "" {
		res.Action = Dirty
		res.Detail = "uncommitted changes, not pulled"
		return res
	}
	if _, err := git(dir, "rev-parse", "--abbrev-ref", "@{upstream}"); err != nil {
		res.Action = Skipped
		res.Detail = "no upstream branch"
		return res
	}
	if out, err := git(dir, "fetch"); err != nil {
		return fail(res, out)
	}
	ahead, behind, err := aheadBehind(dir)
	if err != nil {
		return fail(res, err.Error())
	}

	switch {
	case ahead == 0 && behind == 0:
		res.Action = UpToDate
		return res
	case behind == 0:
		res.Action = Ahead
		res.Detail = fmt.Sprintf("ahead %d, nothing to pull", ahead)
		return res
	case ahead == 0:
		res.Detail = fmt.Sprintf("behind %d", behind)
	default:
		res.Detail = fmt.Sprintf("diverged: ahead %d, behind %d", ahead, behind)
	}
	if out, err := git(dir, "pull"); err != nil {
		return fail(res, res.Detail+"\n"+out)
	}
	res.Action = Pulled
	return res
}

// aheadBehind counts commits on HEAD not in upstream (ahead) and in upstream not in HEAD (behind).
func aheadBehind(dir string) (ahead, behind int, err error) {
	out, err := git(dir, "rev-list", "--left-right", "--count", "HEAD...@{upstream}")
	if err != nil {
		return 0, 0, fmt.Errorf("compare with upstream: %s", out)
	}
	if _, err := fmt.Sscanf(out, "%d %d", &ahead, &behind); err != nil {
		return 0, 0, fmt.Errorf("parse ahead/behind %q: %w", out, err)
	}
	return ahead, behind, nil
}

func fail(res Result, detail string) Result {
	res.Action = Failed
	res.Detail = detail
	return res
}

func primaryRemote(remotes []export.Remote) (export.Remote, bool) {
	for _, r := range remotes {
		if r.Name == "origin" && r.URL != "" {
			return r, true
		}
	}
	for _, r := range remotes {
		if r.URL != "" {
			return r, true
		}
	}
	return export.Remote{}, false
}

func git(dir string, args ...string) (string, error) {
	if dir != "" {
		args = append([]string{"-C", dir}, args...)
	}
	cmd := exec.Command("git", args...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.CombinedOutput()
	return strings.TrimSpace(string(out)), err
}

func expand(path string) string {
	if path == "~" || strings.HasPrefix(path, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			return filepath.Join(home, strings.TrimPrefix(path, "~"))
		}
	}
	return path
}

func within(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != "." && !strings.HasPrefix(rel, "..")
}

func commonDir(dirs []string) string {
	if len(dirs) == 0 {
		return ""
	}
	common := filepath.Clean(dirs[0])
	for _, d := range dirs[1:] {
		d = filepath.Clean(d)
		for common != d && !within(d, common) {
			parent := filepath.Dir(common)
			if parent == common {
				break
			}
			common = parent
		}
	}
	return common
}
