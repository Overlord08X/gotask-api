package dto

import (
	"time"
)

type CreateProjectRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AddMemberRequest struct {
	UserID string `json:"user_id" binding:"required"`
}

type ProjectResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProjectMemberResponse struct {
	ID        string        `json:"id"`
	ProjectID string        `json:"project_id"`
	UserID    string        `json:"user_id"`
	User      *UserResponse `json:"user,omitempty"`
	JoinedAt  time.Time     `json:"joined_at"`
}
