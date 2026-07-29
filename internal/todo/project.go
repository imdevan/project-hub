package todo

import "github.com/project-hub/internal/domain"

// TaskItem represents a single parsed task or subtask.
type TaskItem = domain.TaskItem

// Feature represents a collection of tasks.
type Feature = domain.Feature

type TodoComment = domain.TodoComment

type FtsFileDetail = domain.FtsFileDetail

// Project represents a scanned local development project.
type Project = domain.Project
