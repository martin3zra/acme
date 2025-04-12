package app

import (
	"log"
	"net/http"
	"strconv"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) itemsHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := auth.User(r.Context())
		items, err := s.findItems(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Items/Index", inertia.Props{
			"items": items,
			"units": inertia.Defer(func() (any, error) {
				units, err := s.findUnits(*user.CurrentCompanyId)
				if err != nil {
					s.handleError(w, err)
					return nil, nil
				}
				return units, err
			}, "attributes"),
			"taxes": inertia.Defer(func() (any, error) {
				taxes, err := s.findTaxes(*user.CurrentCompanyId)
				if err != nil {
					s.handleError(w, err)
					return nil, err
				}
				return taxes, nil
			}, "attributes"),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) storeItemHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form StoreItemForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		user := auth.User(r.Context())
		err = s.storeItem(*user.CurrentCompanyId, form)
		if err != nil {
			log.Printf("Error creating item: %v", err)
			s.session.Errors("status", "Item wasn't created. Something went wrong.")
			i.Back(w, r)
			return
		}

		s.session.Flash("success", "Item was created successfully!")

		i.Back(w, r)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) updateItemHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			s.handleError(w, err)
			return
		}

		var form StoreItemForm
		err = support.ParseRequest(r, &form)
		if err != nil {
			s.handleError(w, err, func() {
				http.Redirect(w, r, "/items", http.StatusSeeOther)
			})
			return
		}

		user := auth.User(r.Context())
		err = s.updateItem(*user.CurrentCompanyId, id, form)
		if err != nil {
			s.session.Errors("status", "Item wasn't updated. Something went wrong.")
			i.Back(w, r)
			return

		}

		s.session.Flash("success", "Item was updated successfully!")

		http.Redirect(w, r, "/items", http.StatusSeeOther)
	}
	return http.HandlerFunc(fn)
}

func (s *Server) deleteItemHandler() http.Handler {
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
				http.Redirect(w, r, "/items", http.StatusSeeOther)
			})
			return
		}

		user := auth.User(r.Context())
		err = s.deleteItem(*user.CurrentCompanyId, id)
		if err != nil {
			s.session.Errors("current_password", "Item wasn't deleted. Something went wrong.")
			http.Redirect(w, r, "/items", http.StatusSeeOther)
			return
		}

		s.session.Flash("success", "Item was deleted successfully!")

		http.Redirect(w, r, "/items", http.StatusSeeOther)

	}
	return http.HandlerFunc(fn)
}

func (s *Server) changeStatusItemHandler(i *inertia.Inertia) http.Handler {
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
				http.Redirect(w, r, "/items", http.StatusSeeOther)
			})
			return
		}

		user := form.User()
		item, err := s.findItemByID(*user.CurrentCompanyId, id)
		if err != nil {
			s.session.Errors("status", "Item wasn't updated. Something went wrong.")
			i.Back(w, r)
			return

		}

		err = s.toggleItemStatus(*user.CurrentCompanyId, item)
		if err != nil {
			s.session.Errors("status", "Item wasn't updated. Something went wrong.")
			i.Back(w, r)
			return
		}

		s.session.Flash("success", "Item status was updated successfully!")

		http.Redirect(w, r, "/items", http.StatusSeeOther)
	}
	return http.HandlerFunc(fn)
}
