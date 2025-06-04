package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
)

func (s *Server) customersHandler(ctx *routing.Context) {

	uuid := ctx.Query("id")
	customers, err := s.findCustomers(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("customers")),
		"customers":    customers,
	}
	if ensureUUIDIsValid(uuid) {
		customer, err := s.findCustomeByUUID(ctx.Request.Context(), uuid)
		if err != nil {
			ctx.Error(err)
			return
		}
		props["customer"] = customer
	}

	ctx.Render("Customers/Index", props)
}

func (s *Server) storeCustomerHandler(ctx *routing.Context) {

	var form StoreCustomerForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back(http.StatusBadRequest)
		return
	}

	err = s.storeCustomer(ctx.Request.Context(), form)
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.customer"}))
		ctx.Back(http.StatusBadRequest)
		return
	}

	s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.customer"}))

	ctx.Redirect("/customers")
}

func (s *Server) updateCustomerHandler(ctx *routing.Context) {
	var form UpdateCustomerForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.updateCustomer(ctx.Request.Context(), ctx.Int("id"), form)
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

	ctx.Redirect("/customers")
}

func (s *Server) deleteCustomerHandler(ctx *routing.Context) {
	var form ConfirmsPasswords
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.deleteCustomer(ctx.Request.Context(), ctx.Int("id"))
	if err != nil {
		s.session.Errors("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.customer"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.customer"}))

	ctx.Redirect("/customers")
}

func (s *Server) changeStatusCustomerHandler(ctx *routing.Context) {

	var form ConfirmsPasswords
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	customer, err := s.findCustomeByID(ctx.Request.Context(), ctx.Int("id"))
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
		ctx.Back()
		return
	}

	err = s.toggleCustomerStatus(ctx.Request.Context(), customer)
	if err != nil {
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

	ctx.Redirect("/customers")
}
