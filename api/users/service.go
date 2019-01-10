package users

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"

	"github.com/segmentio/ksuid"
)

// Service represents a service implementation for the "users" resource
type Service api.Service

// New creates a new User
func (s *Service) New(data dtos.ReqCreateUser) *models.User {
	user := &models.User{
		ID:          ksuid.New().String(),
		DisplayName: data.DisplayName,
		Email:       data.Email}
	user.SetPassword(data.Password)
	return user
}

// Find returns all Users
func (s *Service) Find() (*[]models.User, error) {
	var users []models.User
	if err := s.Providers.DB.Select(&users, "SELECT id, email, display_name FROM users"); err != nil {
		return nil, err
	}
	return &users, nil
}

// FindByID returns a User by their ID
func (s *Service) FindByID(id string) (*models.User, error) {
	var user models.User
	if err := s.Providers.DB.Get(&user, "SELECT id, email, display_name FROM users WHERE id = $1", id); err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail returns a User by their email address
func (s *Service) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := s.Providers.DB.Get(
		&user,
		"SELECT id, email, display_name FROM users WHERE email = $1",
		email,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

// Create inserts a user into the database
func (s *Service) Create(user *models.User) error {
	rows, err := s.Providers.DB.NamedQuery(
		`INSERT INTO users (id, display_name, email, password_hash) VALUES (:id,:display_name,:email,:password_hash)
		RETURNING id, display_name, email, password_hash`,
		user,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return err
		}
	}
	return nil
}

// Update updates an existing user in the database
func (s *Service) Update(user *models.User) error {
	rows, err := s.Providers.DB.NamedQuery(
		`UPDATE users SET id = :id, display_name = :display_name, email = :email, password_hash = :password_hash)
		RETURNING id, display_name, email, password_hash`,
		user,
	)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		if err := rows.StructScan(&user); err != nil {
			return err
		}
	}
	return nil
}
