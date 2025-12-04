package workflow

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"
)

// Notifier is an optional hook for task lifecycle events.
type Notifier interface {
	NotifyTaskAssigned(ctx context.Context, task WorkflowTask) error
	NotifyTaskCompleted(ctx context.Context, task WorkflowTask) error
}

// Service manages workflow tasks and approvals.
// It keeps an in-memory store with concurrency safety and lifecycle hooks.
type Service struct {
	mu       sync.RWMutex
	tasks    map[string]WorkflowTask
	logger   *slog.Logger
	notifier Notifier
}

// NewService constructs a workflow service with optional notifier and logger.
func NewService(logger *slog.Logger, notifier Notifier) *Service {
	if logger == nil {
		logger = slog.Default().With("component", "workflow.Service")
	}
	return &Service{
		tasks:    make(map[string]WorkflowTask),
		logger:   logger,
		notifier: notifier,
	}
}

// CreateTask creates and persists a workflow task, notifying the assignee if configured.
func (s *Service) CreateTask(ctx context.Context, task WorkflowTask) (WorkflowTask, error) {
	if s == nil {
		return WorkflowTask{}, errors.New("workflow: service is nil")
	}
	if task.ID == "" {
		return WorkflowTask{}, errors.New("workflow: task ID is required")
	}
	if task.Status == "" {
		task.Status = "pending"
	}
	now := time.Now().UTC()
	if task.CreatedAt.IsZero() {
		task.CreatedAt = now
	}

	s.mu.Lock()
	s.tasks[task.ID] = task
	s.mu.Unlock()

	if s.notifier != nil {
		_ = s.notifier.NotifyTaskAssigned(ctx, task)
	}

	return task, nil
}

// Complete marks a task completed and records the completion time with optional notification.
func (s *Service) Complete(ctx context.Context, id string) error {
	if s == nil {
		return errors.New("workflow: service is nil")
	}

	s.mu.Lock()
	task, ok := s.tasks[id]
	if !ok {
		s.mu.Unlock()
		return errors.New("workflow: task not found")
	}

	now := time.Now().UTC()
	task.Status = "done"
	task.CompletedAt = &now
	s.tasks[id] = task
	s.mu.Unlock()

	if s.notifier != nil {
		_ = s.notifier.NotifyTaskCompleted(ctx, task)
	}
	return nil
}

// List returns a snapshot of all tasks.
func (s *Service) List(ctx context.Context) ([]WorkflowTask, error) {
	if s == nil {
		return nil, errors.New("workflow: service is nil")
	}
	_ = ctx

	s.mu.RLock()
	defer s.mu.RUnlock()

	out := make([]WorkflowTask, 0, len(s.tasks))
	for _, t := range s.tasks {
		out = append(out, t)
	}
	return out, nil
}
