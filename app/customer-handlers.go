package app

import (
	"fmt"

	"github.com/martin3zra/acme/pkg/cache"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/validator/locale"
)

func (s *Server) customersHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "customer") {
		return
	}

	customerType := CustomerType(ctx.Query("customerType"))
	if err := customerType.Validate(); err != nil {
		customerType = "all"
	}

	uuid := ctx.Query("id")
	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	customers, err := s.findCustomers(ctx.Request.Context(), customerType)
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"openState":                 ctx.Query("mode") == "creating",
		"translations":              trans("customers"),
		"customers":                 customers,
		"currentCustomerTypeFilter": customerType,
		"tax_receipts": foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
			return map[string]any{
				"id":        receipt.ID,
				"name":      fmt.Sprintf("%s-%s", receipt.Type, receipt.Name),
				"available": receipt.Current < receipt.SequenceEnd,
			}
		}),
	}
	if uuid != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:customer:%s", uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (*customer, error) {
			return s.findCustomeByUUID(ctx.Request.Context(), uuid)
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["customer"] = data
	}

	ctx.Render("Customers/Index", props)
}

func (s *Server) storeCustomerHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreCustomerForm) {

		if form.OpenBalance > 0 && form.OpenBalanceAsOf.IsZero() {
			ctx.BackWith("open_balance_as_of", fmt.Sprintf(locale.SpanishMessages()["required"].(string), "open_balance_as_of"))
			return
		}

		err := s.storeCustomer(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.customer"}))

		ctx.Redirect("/customers")
	})
}

func (s *Server) updateCustomerHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateCustomerForm) {

		err := s.updateCustomer(ctx.Request.Context(), ctx.Int("id"), form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

		ctx.Redirect("/customers")
	})
}

func (s *Server) deleteCustomerHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.deleteCustomer(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.customer"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.customer"}))

		ctx.Redirect("/customers")
	})
}

func (s *Server) changeStatusCustomerHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		customer, err := s.findCustomeByID(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			return
		}

		err = s.toggleCustomerStatus(ctx.Request.Context(), customer)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.customer"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.customer"}))

		ctx.Redirect("/customers")
	})
}
