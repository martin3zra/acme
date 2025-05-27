package app

import (
	"net/http"
	"strconv"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) customersHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.URL.Query().Get("id")

		customers, err := s.findCustomers(r.Context())
		if err != nil {
			s.handleError(w, err)
			return
		}
		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("customers")),
			"customers":    customers,
		}
		if ensureUUIDIsValid(uuid) {
			customer, err := s.findCustomeByUUID(r.Context(), uuid)
			if err != nil {
				s.handleError(w, err)
				return
			}
			props["customer"] = customer
		}
		err = i.Render(w, r, "Customers/Index", props)
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) storeCustomerHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		var form StoreCustomerForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		err = s.storeCustomer(r.Context(), form)
		if err != nil {
			s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.customer"}))
			i.Back(w, r)
			return
		}

		s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.customer"}))

		i.Back(w, r)
	}

	return http.HandlerFunc(fn)
}

func (s *Server) updateCustomerHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.handleError(w, err)
			return
		}

		var form UpdateCustomerForm
		err = support.ParseRequest(r, &form)
		if err != nil {
			s.handleError(w, err, func() {
				http.Redirect(w, r, "/customers", http.StatusSeeOther)
			})
			return
		}

		err = s.updateCustomer(r.Context(), id, form)
		if err != nil {
			s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			i.Back(w, r)
			return

		}

		s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) deleteCustomerHandler() http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.handleError(w, err)
			return
		}

		var form ConfirmsPasswords
		err = support.ParseRequest(r, &form)
		if err != nil {
			http.Redirect(w, r, "/customers", http.StatusSeeOther)
			return
		}

		err = s.deleteCustomer(r.Context(), id)
		if err != nil {
			s.session.Errors("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.customer"}))
			http.Redirect(w, r, "/customers", http.StatusSeeOther)
			return
		}

		s.session.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.customer"}))

		http.Redirect(w, r, "/customers", http.StatusSeeOther)

	}
	return http.HandlerFunc(fn)
}

func (s *Server) changeStatusCustomerHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.handleError(w, err)
			return
		}

		var form ConfirmsPasswords
		err = support.ParseRequest(r, &form)
		if err != nil {
			s.handleError(w, err, func() {
				http.Redirect(w, r, "/customers", http.StatusSeeOther)
			})
			return
		}

		customer, err := s.findCustomeByID(r.Context(), id)
		if err != nil {
			s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			i.Back(w, r)
			return
		}

		err = s.toggleCustomerStatus(r.Context(), customer)
		if err != nil {
			s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			i.Back(w, r)
			return
		}

		s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
	return http.HandlerFunc(fn)
}
