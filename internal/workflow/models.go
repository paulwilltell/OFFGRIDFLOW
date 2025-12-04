package workflow

import "time"

// WorkflowTask tracks approval workflow tasks.
type WorkflowTask struct {
	ID          string
	Title       string
	Type        string
	Status      string
	Assignee    string
	CreatedAt   time.Time
	CompletedAt *time.Time
	Metadata    map[string]string
	DueAt       *time.Time
}
