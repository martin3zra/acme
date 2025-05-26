package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) createPasswordHandler(i *inertia.Inertia) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {

		session := session.GetSession(r)
		user := auth.User(r.Context())
		var form CreatePasswordForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r, http.StatusSeeOther)
			return
		}

		if err = user.MarkPasswordAsChanged(s.db, form.Password); err != nil {
			session.Errors("password", err.Error())
			i.Back(w, r, http.StatusSeeOther)
			return
		}

		// redirect to home || onboarding
		i.Redirect(w, r, "/onboarding")
	}

	return http.HandlerFunc(fn)
}
