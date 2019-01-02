package users

import (
	"cbs/api"
	"cbs/dtos"
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
	router.Post("/", router.CreateUser)
	router.Get("/{userID}", router.GetUser)
	return router
}

// CreateUser represents a route that creates a new user
func (m *Router) CreateUser(w http.ResponseWriter, r *http.Request) {
	var payload dtos.ReqCreateUser
	if err := api.MustReadAndValidateBody(w, r.Body, &payload); err != nil {
		return
	}
	user := m.Services.User.New(payload)
	if err := m.Services.User.Save(user); err != nil {
		api.SendError(w, api.ErrInternal("Failed to register user"))
		return
	}
	api.SendResponse(w, dtos.ResGetUser{User: user}, http.StatusCreated)
}

// GetUser represents a route that returns a user based on their ID
func (m *Router) GetUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "userID")
	user, err := m.Services.User.FindByID(id)
	if err != nil {
		api.SendError(w, api.NewError(api.ErrCodeNotFound, "User not found", http.StatusNotFound))
		return
	}
	api.SendResponse(w, dtos.ResGetUser{User: user}, http.StatusOK)
}
