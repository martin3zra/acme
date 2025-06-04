package app

import (
	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/support"
)

func (s *Server) createPasswordHandler(ctx *routing.Context) {

	session := session.GetSession(ctx.Request)
	user := auth.User(ctx.Request.Context())
	var form CreatePasswordForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	if err = user.MarkPasswordAsChanged(s.db, form.Password); err != nil {
		session.Errors("password", err.Error())
		ctx.Back()
		return
	}

	// TODO Is ther user isOrphan or have any company then redirect to the homepage,
	// TODO : otherwise if owner redirect to the onboarding

	// redirect to home || onboarding
	ctx.Redirect("/onboarding")
}
