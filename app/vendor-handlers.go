package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
	"github.com/martin3zra/forge/validator/locale"
)

func (s *Server) vendorsHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "vendor") {
		return
	}

	vendorType := VendorType(ctx.Query("vendorType"))
	if err := vendorType.Validate(); err != nil {
		vendorType = "all"
	}

	uuid := ctx.Query("id")
	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	vendors, err := s.findVendors(ctx.Request.Context(), vendorType)
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"openState":               ctx.Query("mode") == "creating",
		"translations":            trans("vendors"),
		"vendors":                 vendors,
		"currentVendorTypeFilter": vendorType,
		"tax_receipts": foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
			return map[string]any{
				"id":        receipt.ID,
				"name":      fmt.Sprintf("%s-%s", receipt.Serie, receipt.Name),
				"available": receipt.Current < receipt.SequenceEnd,
			}
		}),
	}

	if uuid != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:vendor:%s", uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (*vendor, error) {
			return s.findVendorByUUID(ctx.Request.Context(), uuid)
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["vendor"] = data
		props["openState"] = true
	}

	ctx.Render("Vendors/Index", props)
}

func (s *Server) storeVendorHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreVendorForm) {

		if form.OpenBalance > 0 && form.OpenBalanceAsOf.IsZero() {
			ctx.BackWith("open_balance_as_of", fmt.Sprintf(locale.SpanishMessages()["required"].(string), "open_balance_as_of"))
			return
		}

		err := s.storeVendor(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.vendor"}))

		ctx.Redirect("/vendors")
	})
}

func (s *Server) updateVendorHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateVendorForm) {

		err := s.updateVendor(ctx.Request.Context(), ctx.Int("id"), form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.vendor"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:vendor:%s", ctx.Param("id"))
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.vendor"}))

		ctx.Redirect("/vendors")
	})
}

func (s *Server) deleteVendorHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.deleteVendor(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.vendor"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.vendor"}))

		ctx.Redirect("/vendors")
	})
}

func (s *Server) changeStatusVendorHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		vendor, err := s.findVendorByID(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.vendor"}))
			return
		}

		err = s.toggleVendorStatus(ctx.Request.Context(), vendor)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.vendor"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.vendor"}))

		ctx.Redirect("/vendors")
	})
}
