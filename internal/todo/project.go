package todo

// TaskItem represents a single parsed task or subtask.
type TaskItem struct {
	Text      string
	Completed bool
}

// Feature represents a collection of tasks.
type Feature struct {
	Name  string
	Tasks []TaskItem
}

// Project represents a scanned local development project.
type Project struct {
	Name           string
	Path           string
	CompletedTasks int
	TotalTasks     int
	TodoCount      int
	Scanned        bool
}

// Progress returns the completion percentage of tasks, or 0 if no tasks exist.
func (p Project) Progress() int {
	if p.TotalTasks == 0 {
		return 0
	}
	return (p.CompletedTasks * 100) / p.TotalTasks
}
