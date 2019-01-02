package users

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
)

// Service represents a service implementation for the "users" resource
type Service api.Service

// New creates a new User
func (s *Service) New(data dtos.ReqCreateUser) *models.User {
	user := &models.User{
		DisplayName: data.DisplayName,
		Email:       data.Email}
	user.SetPassword(data.Password)
	return user
}

// Find returns all Users
func (s *Service) Find() (*[]models.User, error) {
	var users []models.User
	if err := s.Providers.DB.Find(&users).Error; err != nil {
		return nil, err
	}
	return &users, nil
}

// FindByID returns a User by their ID
func (s *Service) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := s.Providers.DB.Where("id = ?", id).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail returns a User by their email address
func (s *Service) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.Providers.DB.First(&user).Where("name = ?", email).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// Save saves a user to the database
func (s *Service) Save(user *models.User) error {
	if s.Providers.DB.NewRecord(&user) {
		if err := s.Providers.DB.Create(&user).Error; err != nil {
			return err
		}
	} else {
		if err := s.Providers.DB.Save(&user).Error; err != nil {
			return err
		}
	}
	return nil
}
