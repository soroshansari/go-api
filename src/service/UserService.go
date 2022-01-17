package service

import (
	"GoApp/src/database"
	dto "GoApp/src/dto/auth"
	"GoApp/src/provider"
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(user dto.RegisterCredentials) (*database.User, error)
	FindUser(email string) (*database.User, error)
	FindById(id string) (*database.User, error)
	UserExists(email string) (bool, error)
}
type userService struct {
	collection *mongo.Collection
}

func StaticUserService(client *mongo.Client, configs *provider.Configs) UserService {
	return &userService{
		collection: database.OpenCollection(client, "user", configs.DatabaseName),
	}
}
func (service *userService) CreateUser(dto dto.RegisterCredentials) (*database.User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	ID := primitive.NewObjectID()
	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	passwordArr, err := bcrypt.GenerateFromPassword([]byte(*dto.Password), 10)
	if err != nil {
		return nil, err
	}
	password := string(passwordArr[:])
	user := database.User{
		ID:        ID,
		Email:     dto.Email,
		Password:  &password,
		FirstName: dto.FirstName,
		LastName:  dto.LastName,
		CreatedAt: now,
		UpdatedAt: now,
		User_id:   ID.Hex(),
	}

	_, insertErr := service.collection.InsertOne(ctx, user)
	if insertErr != nil {
		return nil, insertErr
	}
	return &user, nil
}

func (service *userService) FindById(id string) (*database.User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user database.User

	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid id")
	}
	filter := bson.M{"_id": objectId}
	err = service.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (service *userService) FindUser(email string) (*database.User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user database.User
	filter := bson.M{"email": email}
	err := service.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (service *userService) UserExists(email string) (bool, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"email": email}
	count, err := service.collection.CountDocuments(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}
