package characters

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi"
)

// Router represents a router for the "characters" resource
type Router api.Router

// NewRouter creates a new router assigned to the "characters" resource
func NewRouter(server *api.Server) *Router {
	router := &Router{
		Mux:    chi.NewMux(),
		Server: server}
	router.Use(
		server.Middlewares.UserSession,
		server.Middlewares.Universe,
		server.Middlewares.Collaborator(models.CollaboratorMember),
	)
	router.Get("/", api.Handler(router.GetCharacters).ServeHTTP)
	router.Post("/", api.Handler(router.CreateCharacter).ServeHTTP)
	router.Route("/{characterID}", func(r chi.Router) {
		r.Use(server.Middlewares.Character)
		r.Get("/", api.Handler(router.GetCharacter).ServeHTTP)
		r.Patch("/", api.Handler(router.EditCharacter).ServeHTTP)
		r.Delete("/", api.Handler(router.DeleteCharacter).ServeHTTP)
	})
	return router
}

// CreateCharacter represents a route that creates a new character
func (m *Router) CreateCharacter(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	var payload dtos.ReqCreateCharacter
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	character := m.Services.Character.New(payload)
	if err := m.Services.Character.Validate(character, universe); err != nil {
		return err
	}
	saved, err := m.Services.Character.Create(universe, character, user)
	if err != nil {
		return err
	}
	api.SendResponse(w, dtos.ResGetCharacter{Character: saved}, http.StatusCreated)
	return nil
}

// GetCharacters represents a route that retrieves a collection of characters pertaining to a universe
func (m *Router) GetCharacters(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	collaborator, _ := r.Context().Value(api.CollaboratorContextKey).(*models.Collaborator)

	qpage := r.URL.Query().Get("p")
	page, err := strconv.Atoi(qpage)
	if err != nil {
		page = 0
	}

	ctx := dtos.CharacterQuery{Collaborator: collaborator, Page: page, Query: ""}
	characters, total, err := m.Services.Character.FindByUniverse(universe, ctx)
	if err != nil {
		return err
	}
	api.SendResponse(w, dtos.ResGetCharacters{Characters: characters, Page: page, Total: total}, http.StatusOK)
	return nil
}

// GetCharacter represents a route that retrieves a single character pertaining to a universe
func (m *Router) GetCharacter(w http.ResponseWriter, r *http.Request) error {
	collaborator, _ := r.Context().Value(api.CollaboratorContextKey).(*models.Collaborator)
	character, _ := r.Context().Value(api.CharacterContextKey).(*models.Character)
	if character.Meta.Hidden &&
		collaborator.Role == models.CollaboratorMember && collaborator.UserID != character.Owner.ID {
		return api.ErrBadAuth("You do not have permission to view this character")
	}
	api.SendResponse(w, dtos.ResGetCharacter{Character: character}, http.StatusOK)
	return nil
}

// EditCharacter represents a route that edits a character
func (m *Router) EditCharacter(w http.ResponseWriter, r *http.Request) error {
	collaborator, _ := r.Context().Value(api.CollaboratorContextKey).(*models.Collaborator)
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	merged, _ := r.Context().Value(api.CharacterContextKey).(*models.Character)
	merged.Fields = nil
	if collaborator.Role == models.CollaboratorMember && collaborator.UserID != merged.Owner.ID {
		return api.ErrBadAuth("You do not have permission to edit this character")
	}
	forbidden := struct {
		ID         string
		UniverseID string
		CreatedAt  time.Time
		UpdatedAt  time.Time
	}{
		ID:         merged.ID,
		UniverseID: merged.UniverseID,
		CreatedAt:  merged.CreatedAt,
		UpdatedAt:  merged.UpdatedAt,
	}
	if err := api.ReadAndValidateBody(r.Body, merged); err != nil {
		return err
	}
	if merged.ID != forbidden.ID || merged.UniverseID != forbidden.UniverseID {
		return api.ErrBadBody("ID cannot be changed")
	}
	if merged.CreatedAt != forbidden.CreatedAt || merged.UpdatedAt != forbidden.UpdatedAt {
		return api.ErrBadBody("Timestamps cannot be changed")
	}
	if err := m.Services.Character.Validate(merged, universe); err != nil {
		return err
	}
	updated, err := m.Services.Character.Update(merged)
	if err != nil {
		return err
	}
	api.SendResponse(w, dtos.ResGetCharacter{Character: updated}, http.StatusOK)
	return nil
}

// DeleteCharacter deletes a character
func (m *Router) DeleteCharacter(w http.ResponseWriter, r *http.Request) error {
	collaborator, _ := r.Context().Value(api.CollaboratorContextKey).(*models.Collaborator)
	character, _ := r.Context().Value(api.CharacterContextKey).(*models.Character)
	if collaborator.Role == models.CollaboratorMember && collaborator.UserID != character.Owner.ID {
		return api.ErrBadAuth("You do not have permission to delete this character")
	}
	if err := m.Services.Character.Delete(character); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
	return nil
}
