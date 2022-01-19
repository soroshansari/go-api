package db

import (
	dto "GoApp/dto/auth"
	"GoApp/providers"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Email          *string            `bson:"email,omitempty"`
	Password       *string            `bson:"password,omitempty"`
	Firstname      *string            `bson:"firstname,omitempty"`
	Lastname       *string            `bson:"lastname,omitempty"`
	ActivationCode string             `bson:"actovationCode,omitempty"`
	Activated      bool               `bson:"activated,omitempty"`
	Profile        string             `bson:"profile,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt      time.Time          `bson:"updatedAt,omitempty"`
}

type UserService interface {
	CreateUser(user dto.RegisterCredentials) (*User, error)
	FindUser(email string) (*User, error)
	FindById(id string) (*User, error)
	UserExists(email string) (bool, error)
	ActivateUser(email, code, password string) (*User, error)
	UpdateActivationCode(email string) (*User, error)
	UpdatePassword(id primitive.ObjectID, password string) error
	UpdateProfile(id primitive.ObjectID, profile string) error
	UpdateDetail(userId, firstname, lastname string) error
}
type userService struct {
	collection *mongo.Collection
}

func NewUserService(client *mongo.Client, configs *providers.Config) UserService {
	return &userService{
		collection: OpenCollection(client, "user", configs.DatabaseName),
	}
}

func (service *userService) CreateUser(dto dto.RegisterCredentials) (*User, error) {
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
	user := User{
		ID:             ID,
		Email:          dto.Email,
		Password:       &password,
		Firstname:      dto.Firstname,
		Lastname:       dto.Lastname,
		ActivationCode: uuid.NewString(),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	_, err = service.collection.InsertOne(ctx, user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (service *userService) FindById(id string) (*User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user User

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

func (service *userService) FindUser(email string) (*User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user User
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

func (service *userService) ActivateUser(email, code, password string) (*User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var user User
	filter := bson.M{"email": email, "actovationCode": code}
	updateSetter := bson.M{"activated": true}
	if password != "" {
		passwordArr, err := bcrypt.GenerateFromPassword([]byte(password), 10)
		if err != nil {
			return nil, err
		}
		updateSetter["password"] = string(passwordArr[:])
	}
	update := bson.M{"$set": updateSetter}
	err := service.collection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (service *userService) UpdatePassword(id primitive.ObjectID, password string) error {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	passwordArr, err := bcrypt.GenerateFromPassword([]byte(password), 10)
	if err != nil {
		return err
	}
	update := bson.M{"$set": bson.M{"password": string(passwordArr[:])}}
	res, err := service.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return errors.New("user not found")
	}
	return nil
}

func (service *userService) UpdateProfile(id primitive.ObjectID, profile string) error {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"_id": id}

	update := bson.M{"$set": bson.M{"profile": profile}}
	res, err := service.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return errors.New("user not found")
	}
	return nil
}

func (service *userService) UpdateDetail(userId, firstname, lastname string) error {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	objectId, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return fmt.Errorf("invalid id")
	}
	filter := bson.M{"_id": objectId}

	update := bson.M{"$set": bson.M{
		"firstname": firstname,
		"lastname":  lastname,
	}}
	res, err := service.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}
	if res.MatchedCount != 1 {
		return errors.New("user not found")
	}
	return nil
}

func (service *userService) UpdateActivationCode(email string) (*User, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	actovationCode := uuid.NewString()

	var user User
	filter := bson.M{"email": email}
	update := bson.M{"$set": bson.M{"actovationCode": actovationCode}}
	err := service.collection.FindOneAndUpdate(ctx, filter, update).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	user.ActivationCode = actovationCode
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
