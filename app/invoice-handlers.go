package app

import (
	"net/http"

	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) invoicesHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		err := i.Render(w, r, "Invoices/Index", inertia.Props{
			"invoices": map[string]any{},
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) createInvoiceHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		term := r.URL.Query().Get("search")

		err := i.Render(w, r, "Invoices/Create", inertia.Props{
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(1, term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"item": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByReference(1, term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}
