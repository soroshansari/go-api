package service

import (
	"GoApp/src/database"
	"GoApp/src/provider"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type RefreshTokenService interface {
	CreateRefreshToken(userId primitive.ObjectID) (string, error)
	FindUserIdbyRefreshToken(tokenId string) (primitive.ObjectID, error)
	RemoveRefreshToken(tokenId string) error
}
type refreshTokenService struct {
	collection *mongo.Collection
}

func StaticRefreshTokenService(client *mongo.Client, configs *provider.Configs) RefreshTokenService {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	collection := database.OpenCollection(client, "refreshToken", configs.DatabaseName)
	mod := mongo.IndexModel{
		Keys: bson.M{
			"tokenId": 1, // index in ascending order
		},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, mod)

	// Check if the CreateOne() method returned any errors
	if err != nil {
		fmt.Println("RefreshToken Indexes().CreateOne() ERROR:", err)
		os.Exit(1) // exit in case of error
	}
	return &refreshTokenService{
		collection: collection,
	}
}

func (service *refreshTokenService) CreateRefreshToken(userId primitive.ObjectID) (string, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	ID := primitive.NewObjectID()
	now, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	refreshToken := database.RefreshToken{
		ID:        ID,
		UserId:    userId,
		TokenId:   uuid.NewString(),
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err := service.collection.InsertOne(ctx, refreshToken)
	if err != nil {
		return "", err
	}

	return refreshToken.TokenId, nil
}

func (service *refreshTokenService) FindUserIdbyRefreshToken(tokenId string) (primitive.ObjectID, error) {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	var refreshToken database.RefreshToken
	filter := bson.M{"tokenId": tokenId}
	err := service.collection.FindOne(ctx, filter).Decode(&refreshToken)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return refreshToken.UserId, nil
}

func (service *refreshTokenService) RemoveRefreshToken(tokenId string) error {
	//this is used to determine how long the API call should last
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	filter := bson.M{"tokenId": tokenId}
	_, err := service.collection.DeleteOne(ctx, filter)
	return err
}
