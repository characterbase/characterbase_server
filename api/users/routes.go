package users

import (
	"cbs/api"
	"cbs/dtos"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// Router represents a router for the "users" resource
type Router api.Router

// NewRouter creates a new router assigned to the "users" resource
func NewRouter(server *api.Server) *Router {
	router := &Router{
		Mux:    chi.NewMux(),
		Server: server}
	router.Post("/", api.Handler(router.CreateUser).ServeHTTP)
	router.Get("/{userID}", api.Handler(router.GetUser).ServeHTTP)
	return router
}

// CreateUser represents a route that creates a new user
func (m *Router) CreateUser(w http.ResponseWriter, r *http.Request) error {
	var payload dtos.ReqCreateUser
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		fmt.Println(err)
		return err
	}
	user := m.Services.User.New(payload)
	if err := m.Services.User.Create(user); err != nil {
		fmt.Println(err)
		return api.ErrInternal("Failed to register user")
	}
	api.SendResponse(w, dtos.ResGetUser{User: user}, http.StatusCreated)
	return nil
}

// GetUser represents a route that returns a user based on their ID
func (m *Router) GetUser(w http.ResponseWriter, r *http.Request) error {
	id := chi.URLParam(r, "userID")
	user, err := m.Services.User.FindByID(id)
	if err != nil {
		return api.ErrNotFound("User not found")
	}
	api.SendResponse(w, dtos.ResGetUser{User: user}, http.StatusOK)
	return nil
}
