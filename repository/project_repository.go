package repository

import (
	"context"
	"time"

	"gotask/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	FindByID(ctx context.Context, id string) (*entity.Project, error)
	FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]entity.Project, error)
	FindAll(ctx context.Context) ([]entity.Project, error)
	FindByUser(ctx context.Context, userID string) ([]entity.Project, error)
	Update(ctx context.Context, id string, project *entity.Project) error
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type projectRepository struct {
	col *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) ProjectRepository {
	return &projectRepository{col: db.Collection("projects")}
}

func (r *projectRepository) Create(ctx context.Context, project *entity.Project) error {
	project.ID = primitive.NewObjectID()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, project)
	return err
}

func (r *projectRepository) FindByID(ctx context.Context, id string) (*entity.Project, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	var project entity.Project
	err = r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&project)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &project, err
}

func (r *projectRepository) FindByIDs(ctx context.Context, ids []primitive.ObjectID) ([]entity.Project, error) {
	cursor, err := r.col.Find(ctx, bson.M{"_id": bson.M{"$in": ids}})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []entity.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) FindAll(ctx context.Context) ([]entity.Project, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []entity.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) FindByUser(ctx context.Context, userID string) ([]entity.Project, error) {
	oid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"created_by": oid})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var projects []entity.Project
	if err := cursor.All(ctx, &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(ctx context.Context, id string, project *entity.Project) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	project.ID = oid
	project.UpdatedAt = time.Now()
	_, err = r.col.ReplaceOne(ctx, bson.M{"_id": oid}, project)
	return err
}

func (r *projectRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}

func (r *projectRepository) Count(ctx context.Context) (int64, error) {
	return r.col.CountDocuments(ctx, bson.M{})
}
