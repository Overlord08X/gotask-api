package repository

import (
	"context"
	"time"

	"gotask/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskRepository interface {
	Create(ctx context.Context, task *entity.Task) error
	FindByID(ctx context.Context, id string) (*entity.Task, error)
	FindByProjectID(ctx context.Context, projectID string) ([]entity.Task, error)
	FindByAssignedTo(ctx context.Context, userID string) ([]entity.Task, error)
	Update(ctx context.Context, id string, task *entity.Task) error
	UpdateStatus(ctx context.Context, id string, status entity.TaskStatus) error
	Delete(ctx context.Context, id string) error
	DeleteByProjectID(ctx context.Context, projectID string) error
	Count(ctx context.Context) (int64, error)
	CountByStatus(ctx context.Context, status entity.TaskStatus) (int64, error)
}

type taskRepository struct {
	col *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) TaskRepository {
	return &taskRepository{col: db.Collection("tasks")}
}

func (r *taskRepository) Create(ctx context.Context, task *entity.Task) error {
	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, task)
	return err
}

func (r *taskRepository) FindByID(ctx context.Context, id string) (*entity.Task, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	var task entity.Task
	err = r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&task)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &task, err
}

func (r *taskRepository) FindByProjectID(ctx context.Context, projectID string) ([]entity.Task, error) {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"project_id": pID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []entity.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) FindByAssignedTo(ctx context.Context, userID string) ([]entity.Task, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"assigned_to": uID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var tasks []entity.Task
	if err := cursor.All(ctx, &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(ctx context.Context, id string, task *entity.Task) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	task.ID = oid
	task.UpdatedAt = time.Now()
	_, err = r.col.ReplaceOne(ctx, bson.M{"_id": oid}, task)
	return err
}

func (r *taskRepository) UpdateStatus(ctx context.Context, id string, status entity.TaskStatus) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{
		"$set": bson.M{
			"status":     status,
			"updated_at": time.Now(),
		},
	})
	return err
}

func (r *taskRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *taskRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteMany(ctx, bson.M{"project_id": pID})
	return err
}

func (r *taskRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}

func (r *taskRepository) CountByStatus(ctx context.Context, status entity.TaskStatus) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{"status": status})
}
