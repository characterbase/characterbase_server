package services

import (
	"cbs/dtos"
	"cbs/models"
)

// User represents the User service layer
type User interface {
	New(data dtos.ReqCreateUser) *models.User
	Find() (*[]models.User, error)
	FindByID(id string) (*models.User, error)
	FindByEmail(email string) (*models.User, error)
	Create(user *models.User) error
	Update(user *models.User) error
}
