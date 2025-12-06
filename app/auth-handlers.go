package app

import (
	"net/http"
	"time"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/support"
)

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

	userCtx := UserFromFoundationUser(user.(*foundation.User))
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

	// Preventing Timing Attacks
	if time.Since(startTime) < duration {
		time.Sleep(duration - time.Since(startTime))
	}

	err = s.sessionManager.ReGenerate(ctx.Request, user, attrs)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Redirect("/home", http.StatusSeeOther)
}

func (s *Server) logoutHandler(ctx *routing.Context) {
	err := s.sessionManager.Invalidate(ctx.Request)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Redirect("/login", http.StatusSeeOther)
}
