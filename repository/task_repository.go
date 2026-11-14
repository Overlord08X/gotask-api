package repository

import (
	"context"
	"gotask/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TaskRepository interface {
	Create(task *entity.Task) error
	GetByID(id primitive.ObjectID) (*entity.Task, error)
	GetByProjectID(projectID primitive.ObjectID) ([]entity.Task, error)
	GetByAssignedTo(userID primitive.ObjectID) ([]entity.Task, error)
	Update(task *entity.Task) error
	Delete(id primitive.ObjectID) error
	CountByStatus(userID primitive.ObjectID) (total, todo, inProgress, done int64, err error)
	CountProjectsByUser(userID primitive.ObjectID) (int64, error)
}

type taskRepository struct {
	col        *mongo.Collection
	projectCol *mongo.Collection
	memberCol  *mongo.Collection
}

func NewTaskRepository(db *mongo.Database) TaskRepository {
	return &taskRepository{
		col:        db.Collection("tasks"),
		projectCol: db.Collection("projects"),
		memberCol:  db.Collection("project_members"),
	}
}

func (r *taskRepository) Create(task *entity.Task) error {
	task.ID = primitive.NewObjectID()
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(context.Background(), task)
	return err
}

func (r *taskRepository) GetByID(id primitive.ObjectID) (*entity.Task, error) {
	var task entity.Task
	err := r.col.FindOne(context.Background(), bson.M{"_id": id}).Decode(&task)
	return &task, err
}

func (r *taskRepository) GetByProjectID(projectID primitive.ObjectID) ([]entity.Task, error) {
	cursor, err := r.col.Find(context.Background(), bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	var tasks []entity.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) GetByAssignedTo(userID primitive.ObjectID) ([]entity.Task, error) {
	cursor, err := r.col.Find(context.Background(), bson.M{"assigned_to": userID})
	if err != nil {
		return nil, err
	}
	var tasks []entity.Task
	if err := cursor.All(context.Background(), &tasks); err != nil {
		return nil, err
	}
	return tasks, nil
}

func (r *taskRepository) Update(task *entity.Task) error {
	task.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(context.Background(), bson.M{"_id": task.ID}, task)
	return err
}

func (r *taskRepository) Delete(id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}

func (r *taskRepository) CountByStatus(userID primitive.ObjectID) (total, todo, inProgress, done int64, err error) {
	ctx := context.Background()
	filter := bson.M{"assigned_to": userID}

	total, err = r.col.CountDocuments(ctx, filter)
	if err != nil {
		return
	}
	todo, err = r.col.CountDocuments(ctx, bson.M{"assigned_to": userID, "status": entity.StatusTodo})
	if err != nil {
		return
	}
	inProgress, err = r.col.CountDocuments(ctx, bson.M{"assigned_to": userID, "status": entity.StatusInProgress})
	if err != nil {
		return
	}
	done, err = r.col.CountDocuments(ctx, bson.M{"assigned_to": userID, "status": entity.StatusDone})
	return
}

func (r *taskRepository) CountProjectsByUser(userID primitive.ObjectID) (int64, error) {
	ctx := context.Background()

	// Hitung project yang dibuat user
	createdCount, err := r.projectCol.CountDocuments(ctx, bson.M{"created_by": userID})
	if err != nil {
		return 0, err
	}

	// Hitung project yang user jadi member
	memberCount, err := r.memberCol.CountDocuments(ctx, bson.M{"user_id": userID})
	if err != nil {
		return 0, err
	}

	// ponytail: bisa overlap jika creator juga member, upgrade ke aggregation $lookup + $group jika perlu exact count
	return createdCount + memberCount, nil
}
