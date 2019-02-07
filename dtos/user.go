package dtos

import (
	"cbs/models"
)

// ReqCreateUser represents a request DTO for creating a new user
type ReqCreateUser struct {
	DisplayName string `json:"displayName" validate:"required,min=3,max=16"`
	Email       string `json:"email" validate:"required,email"`
	Password    string `json:"password" validate:"required"`
}

// ResGetUser represents a response DTO containing public-facing user data
type ResGetUser struct {
	*models.User
}
