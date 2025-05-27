package app

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) onboardingHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		err := i.Render(w, r, "Onboarding/Index", inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("onboarding")),
			"status":       r.URL.Query().Get("status"),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}
