package model

import (
	"RentBuddi/src/database"
	"fmt"
	"time"
)

type User struct {
	User_id     string    `json:"id"`
	Email       *string   `json:"email"`
	DisplayName string    `json:"displayName"`
	FirstName   *string   `json:"firstName"`
	LastName    *string   `json:"lastName"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func GetUser(user *database.User) *User {
	return &User{
		User_id:     user.User_id,
		Email:       user.Email,
		DisplayName: fmt.Sprintf("%s %s", *user.FirstName, *user.LastName),
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
