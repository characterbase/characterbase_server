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
		sr.Use(router.UniverseCtx)
		sr.With(
			server.Middlewares.Collaborator(models.CollaboratorMember),
		).Get("/", router.GetUniverse)
		sr.With(
			server.Middlewares.Collaborator(models.CollaboratorOwner),
		).Patch("/", router.EditUniverse)
	})
	router.Post("/", router.CreateUniverse)
	return router
}

// CreateUniverse represents a route that creates a new universe
func (m *Router) CreateUniverse(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	var payload dtos.ReqCreateUniverse
	if err := api.MustReadAndValidateBody(w, r.Body, &payload); err != nil {
		return
	}
	universe := m.Services.Universe.New(payload)
	if err := m.Services.Universe.Save(universe, user); err != nil {
		api.SendError(w, api.ErrInternal("Failed to create universe"))
		return
	}
	api.SendResponse(w, dtos.ResGetUniverse{Universe: universe}, http.StatusCreated)
}

// GetUniverse represents a route that returns a universe based on its ID
func (m *Router) GetUniverse(w http.ResponseWriter, r *http.Request) {
	universe, _ := r.Context().Value(UniverseContextKey).(*models.Universe)
	api.SendResponse(w, dtos.ResGetUniverse{Universe: universe}, http.StatusOK)
}

// EditUniverse represents a route that modifies a universe based on its ID
func (m *Router) EditUniverse(w http.ResponseWriter, r *http.Request) {
	universe, _ := r.Context().Value(UniverseContextKey).(*models.Universe)
	payload := dtos.ReqEditUniverse{Universe: universe}
	if err := api.MustReadAndValidateBody(w, r.Body, &payload); err != nil {
		return
	}
	if err := m.Services.Universe.Save(payload.Universe, nil); err != nil {
		api.SendError(w, api.ErrInternal("Failed to edit universe"))
	}
	api.SendResponse(w, dtos.ResGetUniverse{Universe: payload.Universe}, http.StatusOK)
}
