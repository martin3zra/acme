package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/database"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/session"
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
				log.Printf("Error converting current company: %v", err)
			}
			ownedBy = getAccount(attrs)
		}

		// The session company predates the variants flag (stored in
		// companies_settings), so refresh it here to keep auth.company.handles_variants
		// correct on every page — the sidebar gates the Attributes nav on it.
		if currentCompany != nil {
			enabled, err := s.handlesVariantsByCompanyID(ctx.Request.Context(), currentCompany.ID)
			if err != nil {
				log.Printf("Error fetching handles_variants for shared props: %v", err)
			} else {
				currentCompany.HandlesVariants = enabled
			}
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
			"sse_url":      s.config.sse.url,
		}

		s.sessionManager.AgeFlash(session)

		ctxWithProps := gonertia.SetProps(ctx.Request.Context(), props)
		next(ctx.WithContext(ctxWithProps))
	}
}

// RequiresVariants gates a route on the company's variants feature flag. When
// off the route responds 404 so the attribute endpoints are fully hidden,
// matching the sidebar/editor gating.
func (s *Server) RequiresVariants(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		enabled, err := s.companyHandlesVariants(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		if !enabled {
			ctx.Error(foundation.HTTPError{StatusCode: http.StatusNotFound, Message: "Not Found"})
			return
		}
		next(ctx)
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

		user := AuthUserFromContext(ctx.Request.Context())
		if !user.IsEmpty() && user.EmailVerifiedAt == nil {
			ctx.Redirect("/verify-email")
			return
		}

		next(ctx)
	}
}

func EnforceVerifiedUserAccess(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		db := ctx.Request.Context().Value(database.ConnectionKey{}).(*sql.DB)
		loggedUser := AuthUserFromContext(ctx.Request.Context())
		if loggedUser.IsEmpty() {
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

// RememberMe restores a session from the persistent-login cookie before the
// auth check runs. On a valid cookie it rebuilds the session (so the downstream
// AuthenticatedMiddleware sees user_id) and rotates the token. No-ops when the
// user already has a session or the cookie is missing/invalid.
func (s *Server) RememberMe(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		sess := session.GetSession(ctx.Request)
		if uid, ok := sess.Get("user_id").(float64); ok && uid > 0 {
			next(ctx)
			return
		}

		cookie, err := ctx.Request.Cookie(rememberCookieName)
		if err != nil || cookie.Value == "" {
			next(ctx)
			return
		}

		userID, err := s.findUserIDByRememberToken(hashRememberToken(cookie.Value))
		if err != nil {
			log.Printf("RememberMe: lookup: %v", err)
			next(ctx)
			return
		}
		if userID == 0 {
			s.expireRememberCookie(ctx) // stale/invalid cookie
			next(ctx)
			return
		}

		user, err := auth.NewAuth(ctx.Request.Context()).LoginUsingId(userID)
		if err != nil {
			log.Printf("RememberMe: load user: %v", err)
			next(ctx)
			return
		}

		attrs := s.sessionAttrsForUser(user)
		if err := s.sessionManager.ReGenerate(ctx.Request, user, attrs); err != nil {
			log.Printf("RememberMe: regenerate: %v", err)
			next(ctx)
			return
		}

		// ReGenerate stores native Go types (user_id as int, user as a struct).
		// On every other request these come back from the session store already
		// JSON-decoded (float64, map[string]any) and the rest of the app asserts
		// those forms. Normalize here so the SAME request doesn't panic.
		sess.Put("user_id", float64(userID))
		if userMap := toJSONMap(user); userMap != nil {
			sess.Put("user", userMap)
		}

		// Token is intentionally NOT rotated here: concurrent requests restoring
		// at once would otherwise invalidate each other's cookie. It stays valid
		// (hashed at rest, HttpOnly) until logout or expiry.
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
			sess.Put("intended", ctx.Request.RequestURI)
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

		acCtx := context.WithValue(userCtx, AccountKey{}, ac)
		if cc == nil {
			next(ctx.WithContext(acCtx))
			return
		}
		ccCtx := context.WithValue(acCtx, CompanyKey{}, cc)

		ctxWithProps := context.WithValue(ccCtx, routing.PermissionKey{}, permissions(cc.UserRole))

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

func AutoResourcePrerequisiteMiddleware(next routing.HandlerFunc) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		resource, ok := resourceFromPath(ctx.Request.URL.Path, true)
		if !ok {
			next(ctx)
			return
		}

		company := CurrentCompany(ctx.Request.Context())
		if company == nil {
			next(ctx)
			return
		}

		rCtx, err := CheckResourcePrerequisites(ctx.Request.Context(), resource, company.ID)
		if err != nil && !errors.Is(err, ErrPrerequisitesMissing) {
			ctx.Error(err, http.StatusPreconditionFailed)
			return
		}

		next(ctx.WithContext(rCtx))
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

	// After a session round-trip the value is a map[string]any; on the same
	// request that restored the session (remember-me) it is still the original
	// *Company. Accept both.
	switch v := raw.(type) {
	case *Company:
		return v, nil
	case Company:
		return &v, nil
	case map[string]any:
		companyStruct, err := mapTo[Company](v)
		if err != nil {
			return nil, err
		}
		return &companyStruct, nil
	default:
		return nil, fmt.Errorf("current_company is not a valid map")
	}
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

func resourceFromPath(path string, base bool) (string, bool) {
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) < 1 {
		return "", false
	}

	var resource string
	if base {
		resource = parts[0] // first child
	} else {
		resource = parts[len(parts)-1] // last child
	}

	// Basic pluralization handling for route resources.
	// Examples: inventories -> inventory, companies -> company, taxes -> tax
	if strings.HasSuffix(resource, "ies") {
		resource = strings.TrimSuffix(resource, "ies") + "y"
	} else if resource == "taxes" {
		resource = "tax"
	} else {
		resource = strings.TrimSuffix(resource, "s")
	}
	return resource, true
}
