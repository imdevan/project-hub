package shell

import (
	"testing"
)

func TestFormatExitCommand(t *testing.T) {
	sh := New("bash")

	ctx := ExitContext{
		Project:    "super-todo",
		ProjectDir: "/home/devy/Projects/clis/super-todo",
		TaskFile:   "/home/devy/Projects/clis/super-todo/plan/tasks.md",
	}

	t.Run("template execution", func(t *testing.T) {
		tmpl := "tmux neww -n {{.Project}} && cd {{.ProjectDir}}"
		got := sh.FormatExitCommand(tmpl, ctx)
		want := "tmux neww -n super-todo && cd /home/devy/Projects/clis/super-todo"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("boolean fallback true", func(t *testing.T) {
		got := sh.FormatExitCommand("true", ctx)
		want := "cd /home/devy/Projects/clis/super-todo"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("empty template", func(t *testing.T) {
		got := sh.FormatExitCommand("", ctx)
		if got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})
}
