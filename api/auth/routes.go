package auth

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
)

// Router represents a router for the "auth" resource
type Router api.Router

// NewRouter creates a new router assigned to the "auth" resource
func NewRouter(server *api.Server) *Router {
	router := &Router{
		Mux:    chi.NewMux(),
		Server: server,
	}
	router.With(server.Middlewares.UserSession).Get("/me", router.Me)
	router.With(server.Middlewares.UserSession).Get("/me/collaborations", router.MyCollaborations)
	router.With(server.Middlewares.UserSession).Get("/logout", router.LogOut)
	router.Post("/login", router.LogIn)
	return router
}

// LogIn represents a route that logs a request into a user session
func (m *Router) LogIn(w http.ResponseWriter, r *http.Request) {
	var payload dtos.ReqLogIn
	if err := api.MustReadAndValidateBody(w, r.Body, &payload); err != nil {
		return
	}
	user, err := m.Services.Auth.Authenticate(payload.Email, payload.Password)
	if err != nil {
		fmt.Println(err)
		api.SendError(w, api.ErrBadAuth(""))
		return
	}
	if err := m.Services.Auth.Login(user, w); err != nil {
		api.SendError(w, api.NewError(api.ErrCodeInternal, "Could not create session", http.StatusInternalServerError))
		return
	}
	api.SendResponse(w, &dtos.ResGetUser{User: user}, http.StatusOK)
}

// Me represents a route that returns a user from the request's associated session
func (m *Router) Me(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	api.SendResponse(w, &dtos.ResGetUser{User: user}, http.StatusOK)
}

// MyCollaborations represents a route that returns the collaborations the user is involved in
func (m *Router) MyCollaborations(w http.ResponseWriter, r *http.Request) {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	universes, err := m.Services.Universe.FindFromUser(user)
	if err != nil {
		api.SendError(w, api.ErrInternal("Could not retrieve universes"))
		return
	}
	api.SendResponse(w, &dtos.ResGetUniverses{References: universes}, http.StatusOK)
}

// LogOut represents a route that logs a request out of a user session
func (m *Router) LogOut(w http.ResponseWriter, r *http.Request) {
	m.Services.Auth.Logout(w)
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
}
