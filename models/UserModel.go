package models

import (
	"GoApp/db"
	"GoApp/providers"
	"fmt"
	"time"
)

type User struct {
	Id          string    `json:"id"`
	Email       *string   `json:"email"`
	DisplayName string    `json:"displayName"`
	Firstname   *string   `json:"firstname"`
	Lastname    *string   `json:"lastname"`
	Profile     string    `json:"profile"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

func GetUser(user *db.User, config *providers.Config) *User {
	_user := User{
		Id:          user.ID.Hex(),
		Email:       user.Email,
		DisplayName: fmt.Sprintf("%s %s", *user.Firstname, *user.Lastname),
		Firstname:   user.Firstname,
		Lastname:    user.Lastname,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
	if user.Profile != "" {
		_user.Profile = config.Domain + "/public/profile/" + user.Profile
	}
	return &_user
}
