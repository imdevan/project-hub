package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// DefaultLockFile is the default name of the generated lock file.
const DefaultLockFile = "projects.lock"

// Remote is a named git remote.
type Remote struct {
	Name string `toml:"name"`
	URL  string `toml:"url"`
}

// Project holds the exported git info for a single project.
type Project struct {
	Path     string   `toml:"path"`
	Branch   string   `toml:"branch,omitempty"`
	Commit   string   `toml:"commit,omitempty"`
	Branches []string `toml:"branches,omitempty"`
	Remotes  []Remote `toml:"remotes,omitempty"`
}

// Lock is the on-disk representation of projects.lock.
type Lock struct {
	Projects []Project `toml:"projects"`
}

// Collect gathers git info for each project directory.
// Paths under the home directory are stored with a leading "~".
func Collect(dirs []string) Lock {
	home, _ := os.UserHomeDir()
	lock := Lock{}
	for _, dir := range dirs {
		lock.Projects = append(lock.Projects, Project{
			Path:     tildePath(dir, home),
			Branch:   git(dir, "rev-parse", "--abbrev-ref", "HEAD"),
			Commit:   git(dir, "rev-parse", "HEAD"),
			Branches: lines(git(dir, "for-each-ref", "--format=%(refname:short)", "refs/heads")),
			Remotes:  remotes(dir),
		})
	}
	sort.Slice(lock.Projects, func(i, j int) bool { return lock.Projects[i].Path < lock.Projects[j].Path })
	return lock
}

// Write encodes the lock as TOML to path.
func Write(lock Lock, path string) error {
	data, err := toml.Marshal(lock)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

// Tree renders the projects as a directory tree, with remotes under each project.
func Tree(lock Lock) string {
	root := &node{children: map[string]*node{}}
	for i := range lock.Projects {
		p := lock.Projects[i]
		n := root
		for _, part := range strings.Split(strings.Trim(p.Path, "/"), "/") {
			if n.children[part] == nil {
				n.children[part] = &node{children: map[string]*node{}}
			}
			n = n.children[part]
		}
		n.project = &lock.Projects[i]
	}
	var sb strings.Builder
	root.render(&sb, "")
	return sb.String()
}

type node struct {
	children map[string]*node
	project  *Project
}

func (n *node) render(sb *strings.Builder, prefix string) {
	names := make([]string, 0, len(n.children))
	for name := range n.children {
		names = append(names, name)
	}
	sort.Strings(names)
	for i, name := range names {
		last := i == len(names)-1
		branch, next := "├── ", "│   "
		if last {
			branch, next = "└── ", "    "
		}
		child := n.children[name]
		sb.WriteString(prefix + branch + name + "\n")
		if child.project != nil {
			for _, r := range child.project.Remotes {
				sb.WriteString(prefix + next + "  ↳ " + r.Name + " " + r.URL + "\n")
			}
		}
		child.render(sb, prefix+next)
	}
}

func remotes(dir string) []Remote {
	var out []Remote
	for _, name := range lines(git(dir, "remote")) {
		out = append(out, Remote{Name: name, URL: git(dir, "remote", "get-url", name)})
	}
	return out
}

func git(dir string, args ...string) string {
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func lines(s string) []string {
	if s == "" {
		return nil
	}
	return strings.Split(s, "\n")
}

func tildePath(path, home string) string {
	if home != "" && (path == home || strings.HasPrefix(path, home+string(filepath.Separator))) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
