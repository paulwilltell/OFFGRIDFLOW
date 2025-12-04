package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/example/offgridflow/internal/api/http/responders"
	"github.com/example/offgridflow/internal/workflow"
	"github.com/google/uuid"
)

// WorkflowHandlerConfig wires workflow service.
type WorkflowHandlerConfig struct {
	Service *workflow.Service
}

// NewWorkflowHandler returns a handler for CRUD on workflow tasks.
func NewWorkflowHandler(cfg WorkflowHandlerConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if cfg.Service == nil {
			responders.ServiceUnavailable(w, "workflow service not configured", 0)
			return
		}

		switch r.Method {
		case http.MethodGet:
			tasks, err := cfg.Service.List(r.Context())
			if err != nil {
				responders.InternalError(w, "failed to list tasks")
				return
			}
			responders.JSON(w, http.StatusOK, tasks)
		case http.MethodPost:
			var req createTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.Title) == "" {
				responders.BadRequest(w, "invalid_request", "title is required")
				return
			}
			task := workflow.WorkflowTask{
				Title:     strings.TrimSpace(req.Title),
				ID:        uuid.NewString(),
				Type:      req.Type,
				Status:    defaultStatus(req.Status),
				Assignee:  req.Assignee,
				Metadata:  req.Metadata,
				CreatedAt: time.Now().UTC(),
			}
			if req.DueAt != nil {
				task.DueAt = req.DueAt
			}

			created, err := cfg.Service.CreateTask(r.Context(), task)
			if err != nil {
				responders.InternalError(w, "failed to create task")
				return
			}
			responders.JSON(w, http.StatusCreated, created)
		case http.MethodPatch:
			var req updateTaskRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID == "" {
				responders.BadRequest(w, "invalid_request", "id is required")
				return
			}
			if req.Status != "" {
				if err := cfg.Service.Complete(r.Context(), req.ID); err != nil {
					responders.BadRequest(w, "update_failed", err.Error())
					return
				}
				responders.JSON(w, http.StatusOK, map[string]string{"status": "updated"})
				return
			}
			responders.BadRequest(w, "invalid_request", "status required")
		default:
			responders.MethodNotAllowed(w, http.MethodGet, http.MethodPost, http.MethodPatch)
		}
	}
}

type createTaskRequest struct {
	Title    string            `json:"title"`
	Type     string            `json:"type"`
	Status   string            `json:"status"`
	Assignee string            `json:"assignee"`
	DueAt    *time.Time        `json:"due_at"`
	Metadata map[string]string `json:"metadata"`
}

type updateTaskRequest struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

func defaultStatus(status string) string {
	s := strings.TrimSpace(strings.ToLower(status))
	if s == "" {
		return "pending"
	}
	return s
}
