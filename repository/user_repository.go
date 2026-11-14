package repository

import (
	"context"
	"time"

	"gotask/entity"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id string) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindAll(ctx context.Context) ([]entity.User, error)
	Update(ctx context.Context, id string, user *entity.User) error
	Delete(ctx context.Context, id string) error
}

type userRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{col: db.Collection("users")}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, user)
	return err
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, nil
	}
	var user entity.User
	err = r.col.FindOne(ctx, bson.M{"_id": oid}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	err := r.col.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == mongo.ErrNoDocuments {
		return nil, nil
	}
	return &user, err
}

func (r *userRepository) FindAll(ctx context.Context) ([]entity.User, error) {
	cursor, err := r.col.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var users []entity.User
	if err := cursor.All(ctx, &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, id string, user *entity.User) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	user.ID = oid
	user.UpdatedAt = time.Now()
	_, err = r.col.ReplaceOne(ctx, bson.M{"_id": oid}, user)
	return err
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return err
	}
	_, err = r.col.DeleteOne(ctx, bson.M{"_id": oid})
	return err
}
