package characters

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi"
)

// MaxRequestSize represents the maximum allowed size for multipart-form requests
const MaxRequestSize = 5 * 1024 * 1024

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
	router.With(server.Middlewares.Collaborator(models.CollaboratorOwner)).Delete(
		"/",
		api.Handler(router.DeleteCharacters).ServeHTTP,
	)
	router.Route("/{characterID}", func(r chi.Router) {
		r.Use(server.Middlewares.Character)
		r.Get("/", api.Handler(router.GetCharacter).ServeHTTP)
		r.Patch("/", api.Handler(router.EditCharacter).ServeHTTP)
		r.Delete("/", api.Handler(router.DeleteCharacter).ServeHTTP)
		r.Delete("/avatar", api.Handler(router.DeleteAvatar).ServeHTTP)
	})
	return router
}

// CreateCharacter represents a route that creates a new character
func (m *Router) CreateCharacter(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	var payload dtos.ReqCreateCharacter

	// Limits the request size to MaxRequestSize
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)

	if err := api.ReadAndValidateBody(strings.NewReader(r.FormValue("data")), &payload); err != nil {
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

	avatar, _, err := r.FormFile("avatar")
	if err == nil {
		if err := m.Services.Character.SetImage(saved, "avatar", avatar); err != nil {
			return err
		}
	}

	images, err := m.Services.Character.FindCharacterImages(saved.ID)
	if err != nil {
		saved.Images = make(models.CharacterImages)
	}
	saved.Images = images

	api.SendResponse(w, dtos.ResGetCharacter{Character: saved}, http.StatusCreated)
	return nil
}

// GetCharacters represents a route that retrieves a collection of characters pertaining to a universe
func (m *Router) GetCharacters(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	collaborator, _ := r.Context().Value(api.CollaboratorContextKey).(*models.Collaborator)

	// Extract the page from the URL parameters
	upage := r.URL.Query().Get("p")
	page, err := strconv.Atoi(upage)
	if err != nil {
		page = 0
	}

	// Extract the search query from the URL parameters
	query := r.URL.Query().Get("q")

	// Extract the sorting order from the URL parameters
	usort := r.URL.Query().Get("s")
	sort := dtos.CharacterQuerySortNominal
	if usort == string(dtos.CharacterQuerySortLexicographical) {
		sort = dtos.CharacterQuerySortLexicographical
	}

	// Extract whether hidden characters should be included from the URL parameters
	uAllowHidden := r.URL.Query().Get("hidden")
	fmt.Println(uAllowHidden)
	allowHidden, err := strconv.ParseBool(uAllowHidden)
	if err != nil {
		allowHidden = true
	}

	// Create the database query context
	ctx := dtos.CharacterQuery{Collaborator: collaborator, Page: page, Query: query, Sort: sort,
		IncludeHidden: allowHidden}

	characters, total, err := m.Services.Character.FindByUniverse(universe, ctx)
	if err != nil {
		return err
	}
	api.SendResponse(w, dtos.ResGetCharacters{Characters: characters, Page: page, Total: total}, http.StatusOK)
	return nil
}

// DeleteCharacters represents a route that deletes all characters from a universe
func (m *Router) DeleteCharacters(w http.ResponseWriter, r *http.Request) error {
	universe, _ := r.Context().Value(api.UniverseContextKey).(*models.Universe)
	if err := m.Services.Character.DeleteAll(universe); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
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

	// Limits the request size to MaxRequestSize
	r.Body = http.MaxBytesReader(w, r.Body, MaxRequestSize)

	if err := api.ReadAndValidateBody(strings.NewReader(r.PostFormValue("data")), merged); err != nil {
		return err
	}

	avatar, _, err := r.FormFile("avatar")
	if err == nil {
		if err := m.Services.Character.SetImage(merged, "avatar", avatar); err != nil {
			return err
		}
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

	updated.Owner = merged.Owner
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
	m.Services.Character.DeleteImage(character, "avatar")
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
	return nil
}

// DeleteAvatar deletes the avatar assigned to a character
func (m *Router) DeleteAvatar(w http.ResponseWriter, r *http.Request) error {
	character, _ := r.Context().Value(api.CharacterContextKey).(*models.Character)
	if err := m.Services.Character.DeleteImage(character, "avatar"); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
	return nil
}
