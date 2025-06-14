package app

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/database"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/support"
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
		session := ctx.Request.Context().Value(session.SessionContextKey{}).(*session.Session)

		sessionUser := session.Get("user")
		csrfToken := session.Get("csrf_token")
		errors := session.Get("errors")
		flash := session.Get("flash")
		translations := trans("global")

		var (
			currentCompany *Company
			ownedBy        map[string]any
		)

		if attrs, ok := session.Get("attrs").(map[string]any); ok && len(attrs) > 0 {
			var err error
			currentCompany, err = getCurrentCompany(attrs)
			if err != nil {
				fmt.Printf("Error converting current company: %v\n", err)
			}
			ownedBy = getAccount(attrs)
		}

		props := map[string]any{
			"auth": map[string]any{
				"user":    sessionUser,
				"company": currentCompany,
				"account": ownedBy,
			},
			"csrf_token":   csrfToken,
			"errors":       errors,
			"flash":        flash,
			"translations": translations,
		}

		s.sessionManager.AgeFlash(session)

		ctxWithProps := gonertia.SetProps(ctx.Request.Context(), props)
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
		if user != nil && !user.IsEmpty() && user.EmailVerifiedAt == nil {
			ctx.Redirect("/verify-email")
			return
		}

		next(ctx)
	}
}

func EnforceVerifiedUserAccess(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		db := ctx.Request.Context().Value(database.ConnectionKey{}).(*sql.DB)
		loggedUser := auth.User(ctx.Request.Context())
		if loggedUser == nil || loggedUser.IsEmpty() {
			next(ctx)
			return
		}

		user := UserFromFoundationUser(loggedUser)

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

func AuthenticatedMiddleware(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		sess := session.GetSession(ctx.Request)

		userId := sess.Get("user_id")
		if userId == nil || userId.(float64) == 0 {
			if ctx.Request.RequestURI == "/" {
				next(ctx)
				return
			}
			ctx.Redirect("/login", http.StatusSeeOther)
			return
		}

		user := sess.Get("user")
		userCtx := context.WithValue(ctx.Request.Context(), auth.ContextUserID{}, user)

		attrs := sess.Get("attrs")
		if attrs == nil {
			next(ctx.WithContext(userCtx))
			return
		}

		attrsMap := attrs.(map[string]any)
		ac := getAccount(attrsMap)
		cc, _ := getCurrentCompany(attrsMap)

		acCtx := context.WithValue(userCtx, support.AccountKey{}, ac)
		ccCtx := context.WithValue(acCtx, support.CompanyKey{}, cc)

		rolePermissions := permissions(cc.UserRole)
		abilities, _ := json.Marshal(rolePermissions)

		// Fetch all permission for the user role or any specific user permission
		ctxWithProps := gonertia.SetTemplateData(ccCtx, gonertia.TemplateData{
			"abilities": template.JS(abilities),
		})

		ctxWithProps = context.WithValue(ctxWithProps, routing.PermissionKey{}, rolePermissions)

		next(ctx.WithContext(ctxWithProps))
	}
}

func RedirectIfAuthenticated(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {

		if strings.HasPrefix(ctx.Request.RequestURI, "/verify-account") {
			next(ctx)
			return
		}
		sess := session.GetSession(ctx.Request)

		userId := sess.Get("user_id")

		if userId != nil && userId.(float64) != 0 {
			ctx.Redirect("/home", http.StatusSeeOther)
			return
		}

		next(ctx)
	}
}

func RestrictedAccess(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		if CurrentCompany(ctx.Request.Context()) == nil {
			ctx.Redirect("/awaiting-association")
			return
		}

		next(ctx)
	}
}

func getCurrentCompany(attrs map[string]any) (*Company, error) {
	raw, ok := attrs["current_company"]
	if !ok {
		return nil, nil
	}

	if raw == nil {
		return nil, nil
	}

	companyMap, ok := raw.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("current_company is not a valid map")
	}

	companyStruct, err := mapTo[Company](companyMap)
	if err != nil {
		return nil, err
	}

	return &companyStruct, nil
}

func getAccount(attrs map[string]any) map[string]any {
	raw, ok := attrs["account"]
	if !ok {
		return nil
	}

	accountMap, ok := raw.(map[string]any)
	if !ok {
		return nil
	}

	return accountMap
}
