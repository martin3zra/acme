package app

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/martin3zra/forge/auth"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/session"
	"github.com/martin3zra/forge/support"
)

const rememberCookieName = "remember"
const rememberCookieMaxAge = 30 * 24 * time.Hour

// rememberCookie builds the persistent-login cookie. Secure is on only when the
// app is served over https (so http://localhost still works in dev).
func (s *Server) rememberCookie(value string, maxAge time.Duration) *http.Cookie {
	return &http.Cookie{
		Name:     rememberCookieName,
		Value:    value,
		Path:     "/",
		Expires:  time.Now().Add(maxAge),
		MaxAge:   int(maxAge.Seconds()),
		HttpOnly: true,
		Secure:   strings.HasPrefix(s.config.host, "https"),
		SameSite: http.SameSiteLaxMode,
	}
}

// issueRememberToken mints a new token, stores its hash, and sets the cookie.
func (s *Server) issueRememberToken(ctx *routing.Context, userID int) {
	raw := generateRememberToken()
	if err := s.storeRememberToken(userID, hashRememberToken(raw)); err != nil {
		log.Printf("issueRememberToken: %v", err)
		return
	}
	http.SetCookie(ctx.Response, s.rememberCookie(raw, rememberCookieMaxAge))
}

// expireRememberCookie clears the cookie on the client.
func (s *Server) expireRememberCookie(ctx *routing.Context) {
	c := s.rememberCookie("", 0)
	c.MaxAge = -1
	c.Expires = time.Unix(0, 0)
	http.SetCookie(ctx.Response, c)
}

func (s *Server) login(ctx *routing.Context) {

	ctx.Render("Auth/Login", map[string]any{
		"translations": trans("auth"),
	})
}

func (s *Server) authHandler(ctx *routing.Context) {

	session := session.GetSession(ctx.Request)
	duration := 1 * time.Second
	startTime := time.Now()

	var form LoginFormRequest
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back(http.StatusSeeOther)
		return
	}

	auth := auth.NewAuth(ctx.Request.Context())
	user, err := auth.Authenticate(form.Email, form.Password)
	if err != nil {
		session.Errors("email", s.trans("global.credentialsDoesNotMatch"))
		ctx.Back(http.StatusSeeOther)
		return
	}

	attrs := s.sessionAttrsForUser(user)

	// Preventing Timing Attacks
	if time.Since(startTime) < duration {
		time.Sleep(duration - time.Since(startTime))
	}

	err = s.sessionManager.ReGenerate(ctx.Request, user, attrs)
	if err != nil {
		ctx.Error(err)
		return
	}

	if form.Remember {
		s.issueRememberToken(ctx, user.GetAuthIdentifier())
	}

	intended := session.Get("intended")
	if intended != nil {
		session.Delete("intended")
		ctx.Redirect(intended.(string), http.StatusSeeOther)
		return
	}
	ctx.Redirect("/home", http.StatusSeeOther)
}

// sessionAttrsForUser builds the session "attrs" (account + current company) for
// an authenticated identity and syncs the user's role. Shared by password login
// and the remember-me restore middleware.
func (s *Server) sessionAttrsForUser(user foundation.Authenticatable) map[string]any {
	userCtx := UserFromFoundationUser(user)
	attrs := map[string]any{"current_company": nil, "account": nil}
	account, err := userCtx.OwnedBy(s.db)
	if err == nil {
		attrs["account"] = map[string]any{
			"id":    account.ID,
			"uuid":  account.UUID,
			"owner": userCtx.Account(s.db) != nil,
		}
	} else {
		account = userCtx.Account(s.db)
		if account != nil {
			attrs["account"] = map[string]any{
				"id":    account.ID,
				"uuid":  account.UUID,
				"owner": userCtx.Account(s.db) != nil,
			}
		}
	}
	company, err := userCtx.currentCompany(s.db)
	if err == nil {
		user.SetRole(company.UserRole)
		attrs["current_company"] = company
	}
	return attrs
}

func (s *Server) logoutHandler(ctx *routing.Context) {
	// Drop the persistent-login token so the cookie can't restore the session.
	sess := session.GetSession(ctx.Request)
	if uid, ok := sess.Get("user_id").(float64); ok && uid > 0 {
		_ = s.clearRememberToken(int(uid))
	}
	s.expireRememberCookie(ctx)

	err := s.sessionManager.Invalidate(ctx.Request)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Redirect("/login", http.StatusSeeOther)
}
