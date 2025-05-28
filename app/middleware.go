package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/justinas/alice"
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/romsar/gonertia/v2"
)

func (s *Server) registerGuestMiddlewares() alice.Chain {
	return alice.New(
		s.SharedProps,
		auth.RedirectIfAuthenticated,
	)
}

func (s *Server) registerAuthMiddlewares() alice.Chain {
	return alice.New(
		s.SharedProps,
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

func (s *Server) SharedProps(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var currentCompany *Company
		session := r.Context().Value(session.SessionContextKey{}).(*session.Session)
		sessionUser := session.Get("user")
		attrs := session.Get("attrs")
		if attrs != nil && len(attrs.(map[string]any)) > 0 {
			attrsMap := attrs.(map[string]any)
			if cc, ok := attrsMap["current_company"]; ok {
				if company, ok := cc.(map[string]any); ok {
					cct, err := mapTo[Company](company)
					if err != nil {
						fmt.Printf("Error converting current company: %v\n", err)
					} else {
						currentCompany = &cct
					}
				}
			}
		}

		ctx := gonertia.SetProps(r.Context(), map[string]any{
			"auth": map[string]any{
				"user":    sessionUser,
				"company": currentCompany,
			},
			"csrf_token":   session.Get("csrf_token"),
			"errors":       session.Get("errors"),
			"flash":        session.Get("flash"),
			"translations": loadTranslations("global"),
		})

		s.sessionManager.AgeFlash(session)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (s *Server) Signed(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if routing.VerifyRequest(r, string(s.config.secretKey)) {
			next.ServeHTTP(w, r)
			return
		}

		s.handleError(w, foundation.Unauthorized{})
	})
}

func EnsurePasswordHasBeenChanged(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var user any = auth.User(r.Context())

		u, ok := user.(foundation.MustVerifyPassword)
		if ok && u.HasNotChangedPassword() {
			http.Redirect(w, r, "/passwords/create", http.StatusFound)
			return
		}

		next.ServeHTTP(w, r)
	})
}
