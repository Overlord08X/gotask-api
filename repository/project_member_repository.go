package repository

import (
	"context"
	"time"

	"gotask/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ProjectMemberRepository interface {
	AddMember(ctx context.Context, member *entity.ProjectMember) error
	RemoveMember(ctx context.Context, projectID string, userID string) error
	FindMembersByProjectID(ctx context.Context, projectID string) ([]entity.ProjectMember, error)
	FindProjectsByUserID(ctx context.Context, userID string) ([]entity.ProjectMember, error)
	FindByProjectAndUser(ctx context.Context, projectID string, userID string) (*entity.ProjectMember, error)
	DeleteByProjectID(ctx context.Context, projectID string) error
}

type projectMemberRepository struct {
	col *mongo.Collection
}

func NewProjectMemberRepository(db *mongo.Database) ProjectMemberRepository {
	return &projectMemberRepository{col: db.Collection("project_members")}
}

func (r *projectMemberRepository) AddMember(ctx context.Context, member *entity.ProjectMember) error {
	member.ID = primitive.NewObjectID()
	member.JoinedAt = time.Now()
	_, err := r.col.InsertOne(ctx, member)
	return err
}

func (r *projectMemberRepository) RemoveMember(ctx context.Context, projectID string, userID string) error {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"project_id": pID, "user_id": uID})
	return err
}

func (r *projectMemberRepository) FindMembersByProjectID(ctx context.Context, projectID string) ([]entity.ProjectMember, error) {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"project_id": pID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []entity.ProjectMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *projectMemberRepository) FindProjectsByUserID(ctx context.Context, userID string) ([]entity.ProjectMember, error) {
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil
	}
	cursor, err := r.col.Find(ctx, bson.M{"user_id": uID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var members []entity.ProjectMember
	if err := cursor.All(ctx, &members); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *projectMemberRepository) FindByProjectAndUser(ctx context.Context, projectID string, userID string) (*entity.ProjectMember, error) {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return nil, nil
	}
	uID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		return nil, nil
	}
	var member entity.ProjectMember
	err = r.col.FindOne(ctx, bson.M{"project_id": pID, "user_id": uID}).Decode(&member)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &member, err
}

func (r *projectMemberRepository) DeleteByProjectID(ctx context.Context, projectID string) error {
	pID, err := primitive.ObjectIDFromHex(projectID)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteMany(ctx, bson.M{"project_id": pID})
	return err
}
