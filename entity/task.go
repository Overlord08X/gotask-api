package entity

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TaskStatus string

const (
	StatusTodo       TaskStatus = "todo"
	StatusInProgress TaskStatus = "in_progress"
	StatusDone       TaskStatus = "done"
)

type TaskPriority string

const (
	PriorityLow    TaskPriority = "low"
	PriorityMedium TaskPriority = "medium"
	PriorityHigh   TaskPriority = "high"
)

type Task struct {
	ID          primitive.ObjectID  `bson:"_id,omitempty" json:"id"`
	ProjectID   primitive.ObjectID  `bson:"project_id" json:"project_id"`
	Title       string              `bson:"title" json:"title"`
	Description string              `bson:"description" json:"description"`
	Status      TaskStatus          `bson:"status" json:"status"`
	Priority    TaskPriority        `bson:"priority" json:"priority"`
	AssignedTo  *primitive.ObjectID `bson:"assigned_to,omitempty" json:"assigned_to"`
	Deadline    *time.Time          `bson:"deadline,omitempty" json:"deadline"`
	CreatedBy   primitive.ObjectID  `bson:"created_by" json:"created_by"`
	CreatedAt   time.Time           `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time           `bson:"updated_at" json:"updated_at"`
}
