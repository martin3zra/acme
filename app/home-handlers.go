package app

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) homeHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		err := i.Render(w, r, "Home/Index", inertia.Props{})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}
