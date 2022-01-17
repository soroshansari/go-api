package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"`
	Email          *string            `bson:"email,omitempty"`
	Password       *string            `bson:"password,omitempty"`
	FirstName      *string            `bson:"firstName,omitempty"`
	LastName       *string            `bson:"lastName,omitempty"`
	ActivationCode string             `bson:"actovationCode,omitempty"`
	Activated      bool               `bson:"activated,omitempty"`
	CreatedAt      time.Time          `bson:"createdAt,omitempty"`
	UpdatedAt      time.Time          `bson:"updatedAt,omitempty"`
}
