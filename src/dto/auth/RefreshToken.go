package dto

type RefreshToken struct {
	Token *string `json:"token" validate:"required,min=2,max=100"`
}
