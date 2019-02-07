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
		return Handler(func(w http.ResponseWriter, r *http.Request) error {
			user, err := services.Auth.User(r)
			if err != nil {
				return ErrBadAuth("")
			}
			// Checking for OK in satisfied routes will be unnecessary because our
			// middleware will catch a session error before the route is reached
			ctx := context.WithValue(r.Context(), UserContextKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	}
}

// MwCollaborator generates a middleware closure designed to sit on top of a MwUserSession
// middleware which fetches the appropriate collaborator respective to the User and Universe
func MwCollaborator(services *Services) func(models.CollaboratorRole) func(http.Handler) http.Handler {
	return func(role models.CollaboratorRole) func(http.Handler) http.Handler {
		return func(next http.Handler) http.Handler {
			return Handler(func(w http.ResponseWriter, r *http.Request) error {
				unid := chi.URLParam(r, "universeID")
				user, _ := r.Context().Value(UserContextKey).(*models.User)
				collaborator, err := services.Universe.FindCollaboratorByID(unid, user.ID)
				if err != nil {
					return ErrBadAuth("You are not a collaborator in this universe")
				}
				switch role {
				case models.CollaboratorAdmin:
					if collaborator.Role == models.CollaboratorMember {
						return ErrBadAuth("Only admins can access this resource")
					}
				case models.CollaboratorOwner:
					if collaborator.Role == models.CollaboratorAdmin || collaborator.Role == models.CollaboratorMember {
						return ErrBadAuth("Only the owner can access this resource")
					}
				}
				ctx := context.WithValue(r.Context(), CollaboratorContextKey, collaborator)
				next.ServeHTTP(w, r.WithContext(ctx))
				return nil
			})
		}
	}

}

// MwUniverse generates a middleware closure that stores a Universe in the request context
func MwUniverse(services *Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id := chi.URLParam(r, "universeID")
			universe, err := services.Universe.FindByID(id)
			if err != nil {
				SendError(w, ErrNotFound("Universe not found"))
				return
			}
			ctx := context.WithValue(r.Context(), UniverseContextKey, universe)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// MwCharacter generates a middleware closure that stores a Character in the request context
func MwCharacter(services *Services) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return Handler(func(w http.ResponseWriter, r *http.Request) error {
			collaborator, _ := r.Context().Value(CollaboratorContextKey).(*models.Collaborator)
			id := chi.URLParam(r, "characterID")
			character, err := services.Character.FindByID(id)
			if err != nil {
				return err
			}
			if collaborator.Role == models.CollaboratorMember && character.OwnerID != collaborator.UserID {
				character.HideHiddenFields()
			}
			ctx := context.WithValue(r.Context(), CharacterContextKey, character)
			next.ServeHTTP(w, r.WithContext(ctx))
			return nil
		})
	}
}
