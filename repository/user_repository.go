package repository

import (
	"context"
	"gotask/entity"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepository interface {
	Create(user *entity.User) error
	GetByID(id primitive.ObjectID) (*entity.User, error)
	GetByEmail(email string) (*entity.User, error)
	GetAll() ([]entity.User, error)
	Update(user *entity.User) error
	Delete(id primitive.ObjectID) error
}

type userRepository struct {
	col *mongo.Collection
}

func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{col: db.Collection("users")}
}

func (r *userRepository) Create(user *entity.User) error {
	user.ID = primitive.NewObjectID()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	_, err := r.col.InsertOne(context.Background(), user)
	return err
}

func (r *userRepository) GetByID(id primitive.ObjectID) (*entity.User, error) {
	var user entity.User
	err := r.col.FindOne(context.Background(), bson.M{"_id": id}).Decode(&user)
	return &user, err
}

func (r *userRepository) GetByEmail(email string) (*entity.User, error) {
	var user entity.User
	err := r.col.FindOne(context.Background(), bson.M{"email": email}).Decode(&user)
	return &user, err
}

func (r *userRepository) GetAll() ([]entity.User, error) {
	cursor, err := r.col.Find(context.Background(), bson.M{})
	if err != nil {
		return nil, err
	}
	var users []entity.User
	if err := cursor.All(context.Background(), &users); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(user *entity.User) error {
	user.UpdatedAt = time.Now()
	_, err := r.col.ReplaceOne(context.Background(), bson.M{"_id": user.ID}, user)
	return err
}

func (r *userRepository) Delete(id primitive.ObjectID) error {
	_, err := r.col.DeleteOne(context.Background(), bson.M{"_id": id})
	return err
}
