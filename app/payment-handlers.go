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
		uuid := r.URL.Query().Get("id")
		user := auth.User(r.Context())
		payments, err := s.findPayments(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}
		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("payments")),
			"payments":     payments,
		}

		if ensureUUIDIsValid(uuid) {
			payment, err := s.findPaymentByUUID(*user.CurrentCompanyId, uuid)
			if err != nil {
				s.handleError(w, err)
				return
			}

			lines, err := s.findPaymentLines(*user.CurrentCompanyId, payment.ID)
			if err != nil {
				s.handleError(w, err)
				return
			}
			props["payment"] = map[string]any{
				"header": payment,
				"lines":  lines,
			}
			props["showPayment"] = true
		}

		err = i.Render(w, r, "Payments/Index", props)
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) createPaymentHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// what we do when the Payment id is given? only return that Payments?
		PaymentUuid := r.URL.Query().Get("Payment_id")
		term := r.URL.Query().Get("search")
		customerUuid := r.URL.Query().Get("customer_id")
		user := auth.User(r.Context())

		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("payments")),
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

		if ensureUUIDIsValid(PaymentUuid) {
			props["Payment_uuid"] = PaymentUuid
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

func (s *Server) voidPaymentHandler(i *inertia.Inertia) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("id")
		back := func() {
			i.Back(w, r, http.StatusSeeOther)
		}
		var form ConfirmsPasswords
		err := support.ParseRequest(r, &form)
		if err != nil {
			back()
			return
		}

		user := auth.User(r.Context())
		err = s.voidPayment(*user.CurrentCompanyId, uuid)
		if err != nil {
			log.Printf("Error voiding payment: %v", err)
			s.session.Errors("status", "Payment wasn't voided. Something went wrong.")
			back()
			return
		}
		s.session.Flash("success", "Payment was voided successfully!")

		i.Redirect(w, r, "/payments", http.StatusSeeOther)
	}

	return http.HandlerFunc(fn)
}

func (s *Server) editPaymentHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("id")
		back := func() {
			i.Back(w, r, http.StatusSeeOther)
		}
		user := auth.User(r.Context())

		if !ensureUUIDIsValid(uuid) {
			back()
			return
		}

		payment, err := s.findPaymentByUUID(*user.CurrentCompanyId, uuid)
		if err != nil {
			s.handleError(w, err)
			return
		}

		lines, err := s.findPaymentLines(*user.CurrentCompanyId, payment.ID)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Payments/Edit", inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("payments")),
			"payment": map[string]any{
				"header": payment,
				"lines":  lines,
			},
			"showPayment": true,
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) updatePaymentHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("id")
		back := func() {
			i.Back(w, r, http.StatusSeeOther)
		}
		var form UpdatePaymentForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			back()
			return
		}

		user := auth.User(r.Context())
		err = s.updatePayment(*user.CurrentCompanyId, uuid, form)
		if err != nil {
			log.Printf("Error recording payment: %v", err)
			s.session.Errors("status", "Payment wasn't recorded. Something went wrong.")
			back()
			return
		}

		s.session.Flash("success", "Payment was recorded successfully!")

		i.Redirect(w, r, "/payments", http.StatusSeeOther)
	}

	return http.HandlerFunc(fn)
}
