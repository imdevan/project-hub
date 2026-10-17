package export

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCollectTreeWrite(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "proj")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	run := func(args ...string) {
		c := exec.Command("git", append([]string{"-C", dir}, args...)...)
		c.Env = append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s", args, out)
		}
	}
	run("init", "-b", "main")
	run("commit", "--allow-empty", "-m", "x")
	run("remote", "add", "origin", "https://example.com/p.git")

	lock := Collect([]string{dir})
	p := lock.Projects[0]
	if p.Branch != "main" || p.Commit == "" || len(p.Remotes) != 1 || p.Remotes[0].URL != "https://example.com/p.git" {
		t.Fatalf("unexpected project: %+v", p)
	}
	if tree := Tree(lock); !strings.Contains(tree, "origin https://example.com/p.git") {
		t.Fatalf("tree missing remote:\n%s", tree)
	}
	out := filepath.Join(t.TempDir(), DefaultLockFile)
	if err := Write(lock, out); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(out)
	if !strings.Contains(string(data), "[[projects]]") {
		t.Fatalf("bad lock:\n%s", data)
	}
}
