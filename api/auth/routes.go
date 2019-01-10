package auth

import (
	"cbs/api"
	"cbs/dtos"
	"cbs/models"
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
	router.With(server.Middlewares.UserSession).Get(
		"/me",
		api.Handler(router.Me).ServeHTTP,
	)
	router.With(server.Middlewares.UserSession).Get(
		"/me/collaborations",
		api.Handler(router.MyCollaborations).ServeHTTP,
	)
	router.With(server.Middlewares.UserSession).Get(
		"/logout",
		api.Handler(router.LogOut).ServeHTTP,
	)
	router.Post(
		"/login",
		api.Handler(router.LogIn).ServeHTTP,
	)
	return router
}

// LogIn represents a route that logs a request into a user session
func (m *Router) LogIn(w http.ResponseWriter, r *http.Request) error {
	var payload dtos.ReqLogIn
	if err := api.ReadAndValidateBody(r.Body, &payload); err != nil {
		return err
	}
	user, err := m.Services.Auth.Authenticate(payload.Email, payload.Password)
	if err != nil {
		return err
	}
	if err := m.Services.Auth.Login(user, w); err != nil {
		return err
	}
	api.SendResponse(w, &dtos.ResGetUser{User: user}, http.StatusOK)
	return nil
}

// Me represents a route that returns a user from the request's associated session
func (m *Router) Me(w http.ResponseWriter, r *http.Request) error {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	api.SendResponse(w, &dtos.ResGetUser{User: user}, http.StatusOK)
	return nil
}

// MyCollaborations represents a route that returns the collaborations the user is involved in
func (m *Router) MyCollaborations(w http.ResponseWriter, r *http.Request) error {
	user, _ := r.Context().Value(api.UserContextKey).(*models.User)
	universes, err := m.Services.Universe.FindFromUser(user)
	if err != nil {
		return api.ErrInternal("Could not retrieve universes")
	}
	api.SendResponse(w, &dtos.ResGetUniverses{References: universes}, http.StatusOK)
	return nil
}

// LogOut represents a route that logs a request out of a user session
func (m *Router) LogOut(w http.ResponseWriter, r *http.Request) error {
	m.Services.Auth.Logout(w)
	w.WriteHeader(http.StatusNoContent)
	w.Write([]byte(""))
	return nil
}
