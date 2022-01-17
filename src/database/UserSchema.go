package database

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {
	ID        primitive.ObjectID `bson:"_id"`
	Email     *string            `json:"email" validate:"required,min=2,max=100"`
	Password  *string            `json:"password" validate:"required,min=2"`
	FirstName *string            `json:"firstName" validate:"required,min=2,max=100"`
	LastName  *string            `json:"lastName" validate:"required,min=2,max=100"`
	CreatedAt time.Time          `json:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt"`
}
