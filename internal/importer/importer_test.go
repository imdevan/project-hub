package importer

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/project-hub/internal/export"
)

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	c := exec.Command("git", append([]string{"-C", dir}, args...)...)
	c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	if out, err := c.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %s", args, out)
	}
}

func TestDestinations(t *testing.T) {
	ws := "/ws"
	lock := export.Lock{Projects: []export.Project{
		{Path: "/old/Projects/a"},
		{Path: "/old/Projects/clis/b"},
		{Path: "rel/c"},
	}}
	got := Destinations(lock, ws)
	want := []string{"/ws/a", "/ws/clis/b", "/ws/rel/c"}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("dest[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	lock = export.Lock{Workspace: "/old", Projects: []export.Project{{Path: "/old/Projects/clis/b"}}}
	if got := Destinations(lock, ws); got[0] != "/ws/Projects/clis/b" {
		t.Errorf("with lock workspace: got %q", got[0])
	}
}

func TestFindLock(t *testing.T) {
	ws := t.TempDir()
	if _, err := FindLock(ws); err == nil {
		t.Fatal("expected error when no lock exists")
	}
	nested := filepath.Join(ws, export.DefaultLockFile, export.DefaultLockFile)
	if err := os.MkdirAll(filepath.Dir(nested), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(nested, nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if got, err := FindLock(ws); err != nil || got != nested {
		t.Fatalf("FindLock = %q, %v; want %q", got, err, nested)
	}
}

func TestImportCloneThenPull(t *testing.T) {
	src := filepath.Join(t.TempDir(), "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, src, "init", "-b", "main")
	gitRun(t, src, "commit", "--allow-empty", "-m", "x")
	gitRun(t, src, "branch", "feature")

	p := export.Project{
		Path:    src,
		Branch:  "feature",
		Remotes: []export.Remote{{Name: "origin", URL: src}, {Name: "upstream", URL: "https://example.com/u.git"}},
	}
	dest := filepath.Join(t.TempDir(), "ws", "proj")

	res := Import(p, dest)
	if res.Action != Cloned || res.Detail != "" {
		t.Fatalf("first import: %+v", res)
	}
	if branch, _ := git(dest, "rev-parse", "--abbrev-ref", "HEAD"); branch != "feature" {
		t.Errorf("branch = %q, want feature", branch)
	}
	if url, _ := git(dest, "remote", "get-url", "upstream"); url != "https://example.com/u.git" {
		t.Errorf("upstream remote = %q", url)
	}

	if res := Import(p, dest); res.Action != UpToDate {
		t.Fatalf("second import: %+v", res)
	}

	if res := Import(export.Project{Path: "x"}, filepath.Join(t.TempDir(), "none")); res.Action != Skipped {
		t.Fatalf("no remotes: %+v", res)
	}
}

func TestPull(t *testing.T) {
	// Pull's rebase creates commits, so it needs an identity independent of the user's global config.
	for k, v := range map[string]string{"GIT_AUTHOR_NAME": "t", "GIT_AUTHOR_EMAIL": "t@t", "GIT_COMMITTER_NAME": "t", "GIT_COMMITTER_EMAIL": "t@t"} {
		t.Setenv(k, v)
	}
	src := filepath.Join(t.TempDir(), "src")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, src, "init", "-b", "main")
	gitRun(t, src, "commit", "--allow-empty", "-m", "x")

	if res := Pull(src); res.Action != Skipped {
		t.Fatalf("no upstream: %+v", res)
	}

	clone := filepath.Join(t.TempDir(), "clone")
	if out, err := exec.Command("git", "clone", src, clone).CombinedOutput(); err != nil {
		t.Fatalf("clone: %s", out)
	}
	gitRun(t, clone, "config", "pull.rebase", "true")

	if res := Pull(clone); res.Action != UpToDate {
		t.Fatalf("up to date: %+v", res)
	}

	// behind → pulled
	gitRun(t, src, "commit", "--allow-empty", "-m", "y")
	if res := Pull(clone); res.Action != Pulled || res.Detail != "behind 1" {
		t.Fatalf("behind: %+v", res)
	}
	head, _ := git(clone, "rev-parse", "HEAD")
	want, _ := git(src, "rev-parse", "HEAD")
	if head != want {
		t.Errorf("clone HEAD %s, want %s", head, want)
	}

	// ahead only → not pulled
	gitRun(t, clone, "commit", "--allow-empty", "-m", "local")
	if res := Pull(clone); res.Action != Ahead || res.Detail != "ahead 1, nothing to pull" {
		t.Fatalf("ahead: %+v", res)
	}

	// diverged → pulled using the repo's pull config (rebase here)
	gitRun(t, src, "commit", "--allow-empty", "-m", "z")
	if res := Pull(clone); res.Action != Pulled || res.Detail != "diverged: ahead 1, behind 1" {
		t.Fatalf("diverged: %+v", res)
	}
	if merges, _ := git(clone, "rev-list", "--merges", "HEAD"); merges != "" {
		t.Errorf("expected rebase (no merge commits), got merges %q", merges)
	}

	// uncommitted tracked change → dirty, not pulled; untracked files don't count
	if err := os.WriteFile(filepath.Join(clone, "untracked"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if res := Pull(clone); res.Action != Ahead {
		t.Fatalf("untracked only: %+v", res)
	}
	gitRun(t, clone, "add", "untracked")
	if res := Pull(clone); res.Action != Dirty {
		t.Fatalf("dirty: %+v", res)
	}
}
