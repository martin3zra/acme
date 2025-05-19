package app

import (
	"context"
	"log"
	"net/http"

	"github.com/justinas/alice"
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
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
		user := session.Get("user")
		if user != nil {
			if user, ok := user.(map[string]any); ok {
				currentCompanyId := int(user["current_company_id"].(float64))
				company, err := s.findCompanyById(currentCompanyId)
				if err != nil {
					log.Printf("error: %v", err)
					http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
					return
				}

				currentCompany = company
			}
		}

		ctx := gonertia.SetProps(r.Context(), map[string]any{
			"auth": map[string]any{
				"user":    user,
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
