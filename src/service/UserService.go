package service

import (
	"GoApp/src/database"
	dto "GoApp/src/dto/auth"
	"GoApp/src/provider"
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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
	ActivateUser(email string, code string) (*database.User, error)
	CreateRefreshToken(userId primitive.ObjectID) (string, error)
	FindUserIdbyRefreshToken(tokenId string) (primitive.ObjectID, error)
	RemoveRefreshToken(tokenId string) error
}
type userService struct {
	userCollection         *mongo.Collection
	refreshTokenCollection *mongo.Collection
}

func StaticUserService(client *mongo.Client, configs *provider.Configs) UserService {
	return &userService{
		userCollection:         database.OpenCollection(client, "user", configs.DatabaseName),
		refreshTokenCollection: database.OpenCollection(client, "refreshToken", configs.DatabaseName),
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
		ID:             ID,
		Email:          dto.Email,
		Password:       &password,
		FirstName:      dto.FirstName,
		LastName:       dto.LastName,
		ActivationCode: uuid.NewString(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = service.userCollection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
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
	err = service.userCollection.FindOne(ctx, filter).Decode(&user)
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
	err := service.userCollection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (service *userService) ActivateUser(email string, code string) (*database.User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user database.User
	filter := bson.M{"email": email, "actovationCode": code}
	update := bson.M{"$set": bson.M{"activated": true}}
	err := service.userCollection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
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
	count, err := service.userCollection.CountDocuments(ctx, filter)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return false, nil
		}
		return false, err
	}
	return count > 0, nil
}

func (service *userService) CreateRefreshToken(userId primitive.ObjectID) (string, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	ID := primitive.NewObjectID()
	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	refreshToken := database.RefreshToken{
		ID:        ID,
		UserId:    userId,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := service.refreshTokenCollection.InsertOne(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return ID.Hex(), nil
}

func (service *userService) FindUserIdbyRefreshToken(tokenId string) (primitive.ObjectID, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(tokenId)
	if err != nil {
		return primitive.NilObjectID, fmt.Errorf("invalid refreshToken")
	}

	var refreshToken database.RefreshToken
	filter := bson.M{"_id": objectId}
	err = service.refreshTokenCollection.FindOne(ctx, filter).Decode(&refreshToken)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return refreshToken.UserId, nil
}

func (service *userService) RemoveRefreshToken(tokenId string) error {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(tokenId)
	if err != nil {
		return fmt.Errorf("invalid refreshToken")
	}

	filter := bson.M{"_id": objectId}
	_, err = service.refreshTokenCollection.DeleteOne(ctx, filter)
	return err
}
