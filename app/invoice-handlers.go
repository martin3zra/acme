package app

import (
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) invoicesHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := auth.User(r.Context())
		invoices, err := s.findInvoices(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Invoices/Index", inertia.Props{
			"invoices": invoices,
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

func (s *Server) storeInvoiceHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form StoreInvoiceForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		if form.Terms == 1 && form.total != form.paymentTotalAmount() {
			s.session.Errors("status", "Invoice total amount and the payment details are different.")
			i.Back(w, r)
			return
		}

		user := auth.User(r.Context())
		err = s.storeInvoice(*user.CurrentCompanyId, form)
		if err != nil {
			log.Printf("Error creating invoice: %v", err)
			s.session.Errors("status", "Invoice wasn't created. Something went wrong.")
			i.Back(w, r)
			return
		}
		s.session.Flash("success", "Invoice was created successfully!")

		i.Back(w, r)
	}

	return http.HandlerFunc(fn)
}
