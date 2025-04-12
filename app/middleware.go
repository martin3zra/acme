package app

import (
	"context"
	"net/http"

	"github.com/justinas/alice"
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/inertia"
	"github.com/martin3zra/acme/pkg/session"
)

func (s *Server) registerGuestMiddlewares() alice.Chain {
	am := inertia.NewInertiaMiddleware(s.sessionManager)
	return alice.New(
		am.SharedProps,
		auth.RedirectIfAuthenticated,
	)
}

func (s *Server) registerAuthMiddlewares() alice.Chain {
	am := inertia.NewInertiaMiddleware(s.sessionManager)
	return alice.New(
		am.SharedProps,
		auth.Middleware,
	)
}

func (s *Server) BindMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.session = session.GetSession(r)

		ctx := context.WithValue(r.Context(), database.ConnectionKey{}, s.db)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}
