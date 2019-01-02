package services

import (
	"cbs/models"
	"net/http"
)

// Auth represents the Authentication service layer
type Auth interface {
	Authenticate(email, password string) (*models.User, error)
	Login(user *models.User, w http.ResponseWriter) error
	Logout(w http.ResponseWriter) error
	User(req *http.Request) (*models.User, error)
}
