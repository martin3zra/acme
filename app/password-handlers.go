package app

import (
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) createPasswordHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *CreatePasswordForm) {

		user := auth.User(ctx.Request.Context())
		if err := user.MarkPasswordAsChanged(s.db, form.Password); err != nil {
			ctx.BackWith("password", err.Error())
			return
		}

		u := UserFromFoundationUser(user)
		if u.IsOwner(s.db) {
			ctx.Redirect("/onboarding")
			return
		}

		ctx.Redirect("/home")
	})
}
