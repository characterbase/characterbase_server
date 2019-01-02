package universes

import (
	"cbs/api"
	"context"
	"net/http"

	"github.com/go-chi/chi"
)

// ContextKey represents a context key
type ContextKey int

const (
	// UniverseContextKey represents a context key for accessing the current universe from the request context
	UniverseContextKey ContextKey = iota
)

// UniverseCtx generates a middleware closure that stores a Universe in the request context
func (m *Router) UniverseCtx(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "universeID")
		universe, err := m.Services.Universe.FindByID(id)
		if err != nil {
			api.SendError(w, api.ErrNotFound("Universe not found"))
		}
		ctx := context.WithValue(r.Context(), UniverseContextKey, universe)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
