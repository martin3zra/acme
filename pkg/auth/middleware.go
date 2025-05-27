package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/martin3zra/acme/pkg/session"
)

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sess := session.GetSession(r)

		userId := sess.Get("user_id")
		if userId == nil || userId.(float64) == 0 {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		user := sess.Get("user")
		ctx := context.WithValue(r.Context(), ContextUserID{}, user)

		req := r.WithContext(ctx)

		attrs := sess.Get("attrs")
		if attrs == nil {
			next.ServeHTTP(w, req)
			return
		}

		attrsMap := attrs.(map[string]any)
		if cc, ok := attrsMap["current_company"]; ok {
			ccCtx := context.WithValue(ctx, ContextCompanyID{}, cc)
			req = r.WithContext(ccCtx)
		}

		next.ServeHTTP(w, req)
	})
}

func RedirectIfAuthenticated(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if strings.HasPrefix(r.RequestURI, "/verify-account") {
			next.ServeHTTP(w, r)
			return
		}
		sess := session.GetSession(r)

		userId := sess.Get("user_id")

		if userId != nil && userId.(float64) != 0 {
			http.Redirect(w, r, "/home", http.StatusSeeOther)
			return
		}

		next.ServeHTTP(w, r)
	})
}
