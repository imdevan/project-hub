package domain

// TaskItem represents a single parsed task or subtask.
type TaskItem struct {
	Line      int
	Text      string
	Completed bool
}

// Feature represents a collection of tasks.
type Feature struct {
	Name  string
	Tasks []TaskItem
}

type TodoComment struct {
	File string
	Line int
	Text string
}

type FtsFileDetail struct {
	Path           string
	CompletedTasks int
	TotalTasks     int
	Features       []Feature
	Ignored        bool `json:"ignored,omitempty"`
}

// Project represents a scanned local development project.
type Project struct {
	Name           string
	Path           string
	CompletedTasks int
	TotalTasks     int
	TodoCount      int
	Scanned        bool
	Todos          []TodoComment
	FtsFiles       []FtsFileDetail
	Pinned         bool
	PinOrder       int
	TmuxAlias      string `json:"tmux_alias,omitempty"`
}

// RecalculateTasks updates CompletedTasks and TotalTasks by summing non-ignored FtsFiles.
func (p *Project) RecalculateTasks() {
	completed := 0
	total := 0
	for _, fts := range p.FtsFiles {
		if !fts.Ignored {
			completed += fts.CompletedTasks
			total += fts.TotalTasks
		}
	}
	p.CompletedTasks = completed
	p.TotalTasks = total
}

// TmuxWindowName returns TmuxAlias if set, otherwise Name.
func (p Project) TmuxWindowName() string {
	if p.TmuxAlias != "" {
		return p.TmuxAlias
	}
	return p.Name
}

// Progress returns the completion percentage of tasks, or 0 if no tasks exist.
func (p Project) Progress() int {
	if p.TotalTasks == 0 {
		return 0
	}
	return (p.CompletedTasks * 100) / p.TotalTasks
}
