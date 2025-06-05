package app

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/romsar/gonertia/v2"
)

func (s *Server) BindMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.session = session.GetSession(r)

		ctx := context.WithValue(r.Context(), database.ConnectionKey{}, s.db)
		ctx = context.WithValue(ctx, ConfigKey{}, s.config)
		req := r.WithContext(ctx)

		next.ServeHTTP(w, req)
	})
}

func (s *Server) SharedProps(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		var ownedBy map[string]any
		var currentCompany *Company
		session := ctx.Request.Context().Value(session.SessionContextKey{}).(*session.Session)
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
			if ac, ok := attrsMap["account"]; ok {
				if oa, ok := ac.(map[string]any); ok {
					ownedBy = oa
				}
			}
		}

		ctxWithProps := gonertia.SetProps(ctx.Request.Context(), map[string]any{
			"auth": map[string]any{
				"user":    sessionUser,
				"company": currentCompany,
				"account": ownedBy,
			},
			"csrf_token":   session.Get("csrf_token"),
			"errors":       session.Get("errors"),
			"flash":        session.Get("flash"),
			"translations": loadTranslations("global"),
		})

		s.sessionManager.AgeFlash(session)

		next(ctx.WithContext(ctxWithProps))
	}
}

func Signed(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		config := ctx.Request.Context().Value(ConfigKey{}).(*Config)
		if routing.VerifyRequest(ctx.Request, string(config.secretKey)) {
			next(ctx)
			return
		}

		ctx.Error(foundation.Unauthorized{})
	}
}

func EnsurePasswordHasBeenChanged(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		var user any = auth.User(ctx.Request.Context())

		u, ok := user.(foundation.MustVerifyPassword)
		if ok && u.HasNotChangedPassword() {
			ctx.Redirect("/passwords/create", http.StatusFound)
			return
		}

		next(ctx)
	}
}

func Verified(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {

		user := auth.User(ctx.Request.Context())
		if user == nil || user.EmailVerifiedAt == nil {
			ctx.Redirect("/verify-email")
			return
		}

		next(ctx)
	}
}

func EnforceVerifiedUserAccess(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		db := ctx.Request.Context().Value(database.ConnectionKey{}).(*sql.DB)
		user := UserFromFoundationUser(auth.User(ctx.Request.Context()))

		if user.IsOrphan(db) {
			ctx.Redirect("/verify-account", http.StatusForbidden)
			return
		}

		if user.IsOwned(db) && !user.account.HasVerifiedAccount() {
			ctx.Redirect("/verify-account", http.StatusForbidden)
			return
		}

		next(ctx)
	}
}
