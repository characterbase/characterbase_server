package universes

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"net/http"

	"github.com/go-chi/chi"
)

// Router represents a router for the "universes" resource
type Router api.Router

// NewRouter creates a new router assigned to the "universes" resource
func NewRouter(server *api.Server) *Router {
	router := &Router{
		Mux:    chi.NewMux(),
		Server: server}
	router.Use(server.Middlewares.UserSession)
	router.Route("/{universeID}", func(sr chi.Router) {
		sr.Use(router.Server.Middlewares.Universe)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorMember)).Get(
			"/",
			api.Handler(router.GetUniverse).ServeHTTP,
		)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorMember)).Get(
			"/collaborators",
			api.Handler(router.GetCollaborators).ServeHTTP,
		)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorOwner)).Post(
			"/collaborators",
			api.Handler(router.AddCollaborator).ServeHTTP,
		)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorOwner)).Patch(
			"/collaborators",
			api.Handler(router.EditCollaborator).ServeHTTP,
		)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorOwner)).Delete(
			"/collaborators",
			api.Handler(router.RemoveCollaborator).ServeHTTP,
		)
		sr.With(server.Middlewares.Collaborator(models.CollaboratorOwner)).Patch(
			"/",
			api.Handler(router.EditUniverse).ServeHTTP,
		)
	})
	router.Post(
		"/",
		api.Handler(router.CreateUniverse).ServeHTTP,
	)
	return router
}

// CreateUniverse represents a route that creates a new universe
func (m *Router) CreateUniverse(w http.ResponseWriter, r *http.Request) error {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	var payload dtos.ReqCreateUniverse
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	universe := m.Services.Universe.New(payload)
	if err := m.Services.Universe.Create(universe, user); err != nil {
		return api.ErrInternal("Failed to create universe")
	}
	api.SendResponse(w, dtos.ResGetUniverse{Universe: universe}, http.StatusCreated)
	return nil
}

// GetUniverse represents a route that returns a universe based on its ID
func (m *Router) GetUniverse(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	api.SendResponse(w, dtos.ResGetUniverse{Universe: universe}, http.StatusOK)
	return nil
}

// EditUniverse represents a route that modifies a universe based on its ID
func (m *Router) EditUniverse(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	payload := dtos.ReqEditUniverse{Universe: universe}
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	if payload.ID != universe.ID {
		return api.ErrBadBody("ID cannot be changed")
	}
	if err := universe.Guide.SetFieldMeta(); err != nil {
		return api.ErrInternal("Failed to edit universe")
	}
	if err := m.Services.Universe.Update(payload.Universe, nil); err != nil {
		return api.ErrInternal("Failed to edit universe")
	}
	api.SendResponse(w, dtos.ResGetUniverse{Universe: payload.Universe}, http.StatusOK)
	return nil
}

// GetCollaborators represents a route that returns a list of collaborators associated with a universe
func (m *Router) GetCollaborators(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	collaborators, err := m.Services.Universe.FindCollaborators(universe)
	if err != nil {
		return api.ErrInternal("Failed to get collaborators")
	}
	api.SendResponse(w, dtos.ResGetCollaborators{Collaborators: collaborators}, http.StatusOK)
	return nil
}

// AddCollaborator represents a route that adds a collaborator to a universe
func (m *Router) AddCollaborator(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	var payload dtos.ReqAddCollaborator
	var user models.User
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	if payload.ID == "" && payload.Email == "" {
		return api.ErrBadBody("Either user ID or user email address must be provided")
	}
	if payload.ID != "" && payload.Email != "" {
		return api.ErrBadBody("Only user ID or user email address may be provided")
	}
	if payload.ID != "" {
		u, err := m.Services.User.FindByID(payload.ID)
		if err != nil {
			return err
		}
		user = *u
	}
	if payload.Email != "" {
		u, err := m.Services.User.FindByEmail(payload.Email)
		if err != nil {
			return err
		}
		user = *u
	}
	collaborator, err := m.Services.Universe.CreateCollaborator(universe, &user, payload.Role)
	if err != nil {
		return err
	}
	api.SendResponse(w, dtos.ResGetCollaborator{Collaborator: collaborator}, http.StatusOK)
	return nil
}

// EditCollaborator represents a route that modifies an existing collaborator
func (m *Router) EditCollaborator(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	var payload dtos.ReqEditCollaborator
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	collaborator, err := m.Services.Universe.FindCollaboratorByID(universe.ID, payload.ID)
	if err != nil {
		return err
	}
	if collaborator.Role == models.CollaboratorOwner {
		return api.ErrBadBody("Cannot edit owner")
	}
	collaborator.Role = payload.Role
	collaborator, err = m.Services.Universe.UpdateCollaborator(universe, collaborator)
	if err != nil {
		return err
	}
	user, err := m.Services.User.FindByID(payload.ID) // TODO: Combine this and the above statement
	if err != nil {
		return err
	}
	collaborator.User = user
	api.SendResponse(w, dtos.ResGetCollaborator{Collaborator: collaborator}, http.StatusOK)
	return nil
}

// RemoveCollaborator represents a route that removes a collaborator from an existing universe
func (m *Router) RemoveCollaborator(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	id := r.URL.Query().Get("id")
	collaborator, err := m.Services.Universe.FindCollaboratorByID(universe.ID, id)
	if err != nil {
		return err
	}
	if collaborator.Role == models.CollaboratorOwner {
		return api.ErrBadBody("Cannot remove owner")
	}
	if err := m.Services.Universe.RemoveCollaborator(universe, collaborator); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
	return nil
}
