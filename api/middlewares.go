package api

import (
	"cbs/models"
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

// MwUserSession generates a middleware closure that fetches a User from the request session
func MwUserSession(services *Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, err := services.Auth.User(r)
			if err != nil {
				SendError(w, ErrBadAuth(""))
				return
			}
			// Checking for OK in satisfied routes will be unnecessary because our
			// middleware will catch a session error before the route is reached
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MwCollaborator generates a middleware closure designed to sit on top of a MwUserSession
// middleware which fetches the appropriate collaborator respective to the User and Universe
func MwCollaborator(services *Services) func(models.CollaboratorRole) func(http.Handler) http.Handler {
	return func(role models.CollaboratorRole) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				unid := chi.URLParam(r, "universeID")
				user, _ := r.Context().Value(UserContextKey).(*models.User)
				collaborator, err := services.Universe.FindCollaboratorFromUser(unid, user)
				if err != nil {
					SendError(w, ErrBadAuth("You are not a collaborator in this universe"))
					return
				}
				switch role {
				case models.CollaboratorAdmin:
					if collaborator.Role == models.CollaboratorMember {
						SendError(w, ErrBadAuth("Only admins can access this resource"))
					}
				case models.CollaboratorOwner:
					if collaborator.Role == models.CollaboratorAdmin || collaborator.Role == models.CollaboratorMember {
						SendError(w, ErrBadAuth("Only the owner can access this resource"))
					}
				}
				ctx := context.WithValue(r.Context(), CollaboratorContextKey, collaborator)
				next.ServeHTTP(w, r.WithContext(ctx))
			})
		}
	}

}
