package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) invoicesHandler(ctx *routing.Context) {
	uuid := ctx.Query("id")
	invoices, err := s.findInvoices(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"translations": trans("invoices"),
		"invoices":     invoices,
	}

	if ensureUUIDIsValid(uuid) {
		invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), uuid)
		if err != nil {
			ctx.Error(err)
			return
		}

		lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.ID)
		if err != nil {
			ctx.Error(err)
			return
		}

		props["invoice"] = map[string]any{
			"header": invoice,
			"lines":  lines,
		}
		props["showInvoice"] = true
	}

	ctx.Render("Invoices/Index", props)
}

func (s *Server) createInvoiceHandler(ctx *routing.Context) {
	term := ctx.Query("search")
	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Invoices/Create", map[string]any{
		"translations": trans("invoices"),
		"tax_receipts": foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
			return map[string]any{
				"id":        receipt.ID,
				"name":      fmt.Sprintf("%s-%s", receipt.Type, receipt.Name),
				"available": receipt.Current < receipt.SequenceEnd,
			}
		}),
		"customers": inertia.Optional(func() (any, error) {
			customers, err := s.findCustomersBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return customers, err
		}),
		"item": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByReference(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return item, err
		}),
		"items": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return item, err
		}),
	})
}

func (s *Server) editInvoiceHandler(ctx *routing.Context) {
	invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.ID)
	if err != nil {
		ctx.Error(err)
	}

	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	term := ctx.Query("search")
	ctx.Render("Invoices/Edit", map[string]any{
		"translations": trans("invoices"),
		"invoice": map[string]any{
			"header": invoice,
			"lines":  lines,
		},
		"tax_receipts": foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
			return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name)}
		}),
		"customers": inertia.Optional(func() (any, error) {
			customers, err := s.findCustomersBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return customers, err
		}),
		"item": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByReference(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return item, err
		}),
		"items": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return item, err
		}),
	})
}

func (s *Server) storeInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreInvoiceForm) {

		if form.Terms == 1 && form.total != form.paymentTotalAmount() {
			ctx.BackWith("status", "Invoice total amount and the payment details are different.")
			return
		}

		err := s.storeInvoice(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating invoice: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.invoice"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.invoice"}))

		ctx.Redirect("/invoices")
	})
}

func (s *Server) updateInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateInvoiceForm) {

		if form.Terms == 1 && form.total != form.paymentTotalAmount() {
			ctx.BackWith("status", "Invoice total amount and the payment details are different.")
			return
		}

		err := s.updateInvoice(ctx.Request.Context(), ctx.Param("id"), form)
		if err != nil {
			log.Printf("Error updating invoice: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.invoice"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.invoice"}))

		ctx.Redirect("/invoices")
	})
}

func (s *Server) voidInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.voidInvoice(ctx.Request.Context(), ctx.Param("uuid"))
		if err != nil {
			log.Printf("Error voiding invoice: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotVoided", i18n.Replacements{"subject": "@global.invoice"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasVoided", i18n.Replacements{"subject": "@global.invoice"}))

		ctx.Redirect("/invoices")
	})
}
