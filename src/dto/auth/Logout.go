package dto

type Logout struct {
	Token *string `json:"token" validate:"required,min=2,max=100"`
}
