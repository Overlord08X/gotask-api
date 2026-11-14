package repository

import (
	"context"
	"gotask/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectRepository interface {
	Create(project *entity.Project) error
	GetByID(id primitive.ObjectID) (*entity.Project, error)
	GetAll() ([]entity.Project, error)
	GetByUserID(userID primitive.ObjectID) ([]entity.Project, error)
	Update(project *entity.Project) error
	Delete(id primitive.ObjectID) error
}

type projectRepository struct {
	col       *mongo.Collection
	memberCol *mongo.Collection
}

func NewProjectRepository(db *mongo.Database) ProjectRepository {
	return &projectRepository{
		col:       db.Collection("projects"),
		memberCol: db.Collection("project_members"),
	}
}

func (r *projectRepository) Create(project *entity.Project) error {
	project.ID = primitive.NewObjectID()
	project.CreatedAt = time.Now()
	project.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(context.Background(), project)
	return err
}

func (r *projectRepository) GetByID(id primitive.ObjectID) (*entity.Project, error) {
	var project entity.Project
	err := r.col.FindOne(context.Background(), bson.M{"_id": id}).Decode(&project)
	return &project, err
}

func (r *projectRepository) GetAll() ([]entity.Project, error) {
	cursor, err := r.col.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	var projects []entity.Project
	if err := cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) GetByUserID(userID primitive.ObjectID) ([]entity.Project, error) {
	// Cari project yang dibuat user atau user jadi member
	memberCursor, err := r.memberCol.Find(context.Background(), bson.M{"user_id": userID})
	if err != nil {
		return nil, err
	}
	var members []entity.ProjectMember
	if err := memberCursor.All(context.Background(), &members); err != nil {
		return nil, err
	}

	projectIDs := []primitive.ObjectID{}
	for _, m := range members {
		projectIDs = append(projectIDs, m.ProjectID)
	}

	filter := bson.M{
		"$or": []bson.M{
			{"created_by": userID},
			{"_id": bson.M{"$in": projectIDs}},
		},
	}

	cursor, err := r.col.Find(context.Background(), filter)
	if err != nil {
		return nil, err
	}
	var projects []entity.Project
	if err := cursor.All(context.Background(), &projects); err != nil {
		return nil, err
	}
	return projects, nil
}

func (r *projectRepository) Update(project *entity.Project) error {
	project.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(context.Background(), bson.M{"_id": project.ID}, project)
	return err
}

func (r *projectRepository) Delete(id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
