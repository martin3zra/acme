package app

import (
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/support"
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

func (s *Server) createPaymentHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// what we do when the invoice id is given? only return that invoices?
		invoiceUuid := r.URL.Query().Get("invoice_id")
		term := r.URL.Query().Get("search")
		customerUuid := r.URL.Query().Get("customer_id")
		user := auth.User(r.Context())

		props := inertia.Props{
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(*user.CurrentCompanyId, term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"receivables": inertia.Optional(func() (any, error) {
				receivables, err := s.findCustomeReceivables(*user.CurrentCompanyId, customerUuid)
				if err != nil {
					return nil, err
				}

				return receivables, err
			}),
		}

		if ensureUUIDIsValid(invoiceUuid) {
			props["invoice_uuid"] = invoiceUuid
		}

		if ensureUUIDIsValid(customerUuid) {
			customer, err := s.findCustomeByUUID(*user.CurrentCompanyId, customerUuid)
			if err != nil {
				s.handleError(w, err)
				return
			}
			receivables, err := s.findCustomeReceivables(*user.CurrentCompanyId, customerUuid)
			if err != nil {
				s.handleError(w, err)
				return
			}

			props["customer"] = customer
			props["receivables"] = receivables
			props["forceInitial"] = true
		}
		err := i.Render(w, r, "Payments/Create", props)
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) storePaymentHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form StorePaymentForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		user := auth.User(r.Context())
		err = s.storePayment(*user.CurrentCompanyId, form)
		if err != nil {
			log.Printf("Error recording payment: %v", err)
			s.session.Errors("status", "Payment wasn't recorded. Something went wrong.")
			i.Back(w, r)
			return
		}

		s.session.Flash("success", "Payment was recorded successfully!")

		i.Back(w, r)
	}

	return http.HandlerFunc(fn)
}
