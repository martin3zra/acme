package inertia

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/session"
	"github.com/romsar/gonertia/v2"
)

type InertiaMiddleware struct {
	manager *session.SessionManager
}

func NewInertiaMiddleware(manager *session.SessionManager) InertiaMiddleware {
	return InertiaMiddleware{manager: manager}
}

func (a InertiaMiddleware) SharedProps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		session := r.Context().Value(session.SessionContextKey{}).(*session.Session)
		ctx := gonertia.SetProps(r.Context(), map[string]any{
			"auth": map[string]any{
				"user": session.Get("user"),
			},
			"csrf_token": session.Get("csrf_token"),
			"errors":     session.Get("errors"),
			"flash":      session.Get("flash"),
		})

		a.manager.AgeFlash(session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
