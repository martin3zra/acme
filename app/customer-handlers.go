package app

import (
	"net/http"
	"strconv"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) customersHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.URL.Query().Get("id")
		user := auth.User(r.Context())
		customers, err := s.findCustomers(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}
		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("customers")),
			"customers":    customers,
		}
		if ensureUUIDIsValid(uuid) {
			customer, err := s.findCustomeByUUID(*user.CurrentCompanyId, uuid)
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

		user := auth.User(r.Context())
		err = s.storeCustomer(*user.CurrentCompanyId, form)
		if err != nil {
			s.session.Errors("status", "Customer wasn't created. Something went wrong.")
			i.Back(w, r)
			return
		}

		s.session.Flash("success", "Customer was created successfully!")

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

		err = s.updateCustomer(*form.User().CurrentCompanyId, id, form)
		if err != nil {
			s.session.Errors("status", "Customer wasn't updated. Something went wrong.")
			i.Back(w, r)
			return

		}

		s.session.Flash("success", "Customer was updated successfully!")

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

		user := form.User()
		err = s.deleteCustomer(*user.CurrentCompanyId, id)
		if err != nil {
			s.session.Errors("current_password", "Customer wasn't deleted. Something went wrong.")
			http.Redirect(w, r, "/customers", http.StatusSeeOther)
			return
		}

		s.session.Flash("success", "Customer was deleted successfully!")

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

		user := form.User()
		customer, err := s.findCustomeByID(*user.CurrentCompanyId, id)
		if err != nil {
			s.session.Errors("status", "Customer wasn't updated. Something went wrong.")
			i.Back(w, r)
			return
		}

		err = s.toggleCustomerStatus(*user.CurrentCompanyId, customer)
		if err != nil {
			s.session.Errors("status", "Customer wasn't updated. Something went wrong.")
			i.Back(w, r)
			return
		}

		s.session.Flash("success", "Customer status was updated successfully!")

		http.Redirect(w, r, "/customers", http.StatusSeeOther)
	}
	return http.HandlerFunc(fn)
}
