package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) paymentsHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		user := auth.User(r.Context())
		payments, err := s.findPayments(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Payments/Index", inertia.Props{
			"payments": payments,
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}
