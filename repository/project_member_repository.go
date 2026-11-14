package repository

import (
	"context"
	"gotask/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectMemberRepository interface {
	Add(member *entity.ProjectMember) error
	Remove(projectID, userID primitive.ObjectID) error
	GetByProjectID(projectID primitive.ObjectID) ([]entity.ProjectMember, error)
	IsMember(projectID, userID primitive.ObjectID) (bool, error)
}

type projectMemberRepository struct {
	col *mongo.Collection
}

func NewProjectMemberRepository(db *mongo.Database) ProjectMemberRepository {
	return &projectMemberRepository{col: db.Collection("project_members")}
}

func (r *projectMemberRepository) Add(member *entity.ProjectMember) error {
	member.ID = primitive.NewObjectID()
	member.JoinedAt = time.Now()
	_, err := r.col.InsertOne(context.Background(), member)
	return err
}

func (r *projectMemberRepository) Remove(projectID, userID primitive.ObjectID) error {
	_, err := r.col.DeleteOne(context.Background(), bson.M{
		"project_id": projectID,
		"user_id":    userID,
	})
	return err
}

func (r *projectMemberRepository) GetByProjectID(projectID primitive.ObjectID) ([]entity.ProjectMember, error) {
	cursor, err := r.col.Find(context.Background(), bson.M{"project_id": projectID})
	if err != nil {
		return nil, err
	}
	var members []entity.ProjectMember
	if err := cursor.All(context.Background(), &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *projectMemberRepository) IsMember(projectID, userID primitive.ObjectID) (bool, error) {
	count, err := r.col.CountDocuments(context.Background(), bson.M{
		"project_id": projectID,
		"user_id":    userID,
	})
	return count > 0, err
}
