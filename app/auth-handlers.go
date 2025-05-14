package app

import (
	"net/http"
	"time"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/support"

	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) loginHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		err := i.Render(w, r, "Auth/Login", inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("auth")),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) authHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		session := session.GetSession(r)
		duration := 1 * time.Second
		startTime := time.Now()

		var form LoginFormRequest
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r, http.StatusSeeOther)
			return
		}

		auth := auth.NewAuth(r.Context())
		user, err := auth.Authenticate(form.Email, form.Password)
		if err != nil {
			session.Errors("email", "These credentials do not match our records.")
			i.Back(w, r, http.StatusSeeOther)
			return
		}

		// Preventing Timing Attacks
		if time.Since(startTime) < duration {
			time.Sleep(duration - time.Since(startTime))
		}

		err = s.sessionManager.ReGenerate(r, user)
		if err != nil {
			s.handleError(w, err)
			return
		}

		session.Flash("success", "Congrats!!!")

		i.Redirect(w, r, "/home")
	}

	return http.HandlerFunc(fn)
}

func (s *Server) logoutHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		err := s.sessionManager.Invalidate(r)
		if err != nil {
			s.handleError(w, err)
			return
		}

		i.Back(w, r)
	}

	return http.HandlerFunc(fn)
}
