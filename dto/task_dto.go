package dto

import (
	"time"
)

type CreateTaskRequest struct {
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
	Deadline    *time.Time `json:"deadline"`
}

type UpdateTaskRequest struct {
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
	Deadline    *time.Time `json:"deadline"`
	Status      string     `json:"status"`
}

type UpdateTaskStatusRequest struct {
	Status string `json:"status" binding:"required"`
}

type TaskResponse struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"project_id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	Priority    string     `json:"priority"`
	AssignedTo  *string    `json:"assigned_to"`
	Deadline    *time.Time `json:"deadline"`
	CreatedBy   string     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type DashboardResponse struct {
	TotalProjects   int64 `json:"total_projects"`
	TotalTasks      int64 `json:"total_tasks"`
	TasksTodo       int64 `json:"tasks_todo"`
	TasksInProgress int64 `json:"tasks_in_progress"`
	TasksDone       int64 `json:"tasks_done"`
}

type WebResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}
