package main

import (
	"strings"
	"testing"

	"github.com/super-todo/internal/domain"
)

func TestRenderConfigTemplate(t *testing.T) {
	cfg := domain.DefaultConfig()
	cfg.IgnoreFolders = []string{"node_modules", "dist", ".git"}

	res := renderConfigTemplate(cfg)

	expected := `# ignore_folders = ["node_modules", "dist", ".git"]`
	if !strings.Contains(res, expected) {
		t.Errorf("expected config template to contain %q, but got:\n%s", expected, res)
	}
}
