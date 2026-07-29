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

func TestProject_RecalculateTasks(t *testing.T) {
	p := Project{
		FtsFiles: []FtsFileDetail{
			{Path: "f1.md", CompletedTasks: 2, TotalTasks: 5, Ignored: false},
			{Path: "f2.md", CompletedTasks: 3, TotalTasks: 4, Ignored: true},
			{Path: "f3.md", CompletedTasks: 1, TotalTasks: 1, Ignored: false},
		},
	}

	p.RecalculateTasks()

	if p.CompletedTasks != 3 {
		t.Errorf("CompletedTasks = %d, want 3", p.CompletedTasks)
	}
	if p.TotalTasks != 6 {
		t.Errorf("TotalTasks = %d, want 6", p.TotalTasks)
	}
}
