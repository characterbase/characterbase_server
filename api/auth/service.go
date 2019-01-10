package auth

import (
	"cbs/api"
	"cbs/models"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/segmentio/ksuid"
	"golang.org/x/crypto/bcrypt"
)

// Service represents a service implementation for the "auth" resource
type Service api.Service

func newAPIKey() string {
	return ksuid.New().String()
}

// Authenticate returns a User if the passed credentials are valid
func (s *Service) Authenticate(email, password string) (*models.User, error) {
	var user models.User
	if err := s.Providers.DB.Get(&user, "SELECT password_hash FROM users WHERE email = $1", email); err != nil {
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, api.ErrBadAuth("")
	}
	user = models.User{}
	if err := s.Providers.DB.Get(
		&user,
		"SELECT id, display_name, email FROM users WHERE email = $1",
		email,
	); err != nil {
		return nil, err
	}
	return &user, nil
}

// Login creates a new session between the request and the user
func (s *Service) Login(user *models.User, w http.ResponseWriter) error {
	sesskey := newAPIKey()
	serialized, err := json.Marshal(user)
	if err != nil {
		return err
	}
	maxage, err := time.ParseDuration(s.Config.MaxSessionAge)
	if err != nil {
		return err
	}
	s.Providers.Redis.Set(fmt.Sprintf("session:%v", sesskey), serialized, maxage)
	http.SetCookie(w, &http.Cookie{
		Name:   "user_session",
		Value:  sesskey,
		MaxAge: int(maxage.Seconds()),
	})
	return nil
}

// Logout destroys all sessions belonging to the request
func (s *Service) Logout(w http.ResponseWriter) error {
	http.SetCookie(w, &http.Cookie{
		Name:    "user_session",
		MaxAge:  -1,
		Expires: time.Now().Add(-100 * time.Hour), // Negative expire to support old browsers (e.g. IE)
	})
	return nil
}

// User returns the User associated with the request's session
func (s *Service) User(req *http.Request) (*models.User, error) {
	var user models.User
	sesskey, err := req.Cookie("user_session")
	if err != nil {
		return nil, err
	}
	serialized, err := s.Providers.Redis.Get(fmt.Sprintf("session:%v", sesskey.Value)).Result()
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal([]byte(serialized), &user); err != nil {
		return nil, err
	}
	return &user, nil
}
