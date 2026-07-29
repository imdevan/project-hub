package domain

import "testing"

func TestProject_TmuxWindowName(t *testing.T) {
	p := Project{
		Name: "my-project",
	}

	if got := p.TmuxWindowName(); got != "my-project" {
		t.Errorf("TmuxWindowName() = %q, want %q", got, "my-project")
	}

	p.TmuxAlias = "custom-alias"
	if got := p.TmuxWindowName(); got != "custom-alias" {
		t.Errorf("TmuxWindowName() = %q, want %q", got, "custom-alias")
	}
}
