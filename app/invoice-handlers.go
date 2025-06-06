package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/support"
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
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("invoices")),
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
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("invoices")),
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
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("invoices")),
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

func (s *Server) storeInvoiceHandler(ctx *routing.Context) {
	var form StoreInvoiceForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	if form.Terms == 1 && form.total != form.paymentTotalAmount() {
		s.session.Errors("status", "Invoice total amount and the payment details are different.")
		ctx.Back()
		return
	}

	err = s.storeInvoice(ctx.Request.Context(), form)
	if err != nil {
		log.Printf("Error creating invoice: %v", err)
		s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.invoice"}))
		ctx.Back()
		return
	}
	s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.invoice"}))

	ctx.Redirect("/invoices")
}

func (s *Server) updateInvoiceHandler(ctx *routing.Context) {
	var form UpdateInvoiceForm
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	if form.Terms == 1 && form.total != form.paymentTotalAmount() {
		s.session.Errors("status", "Invoice total amount and the payment details are different.")
		ctx.Back()
		return
	}

	err = s.updateInvoice(ctx.Request.Context(), ctx.Param("id"), form)
	if err != nil {
		log.Printf("Error updating invoice: %v", err)
		s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.invoice"}))
		ctx.Back()
		return
	}

	s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.invoice"}))

	ctx.Redirect("/invoices")
}

func (s *Server) voidInvoiceHandler(ctx *routing.Context) {
	var form ConfirmsPasswords
	err := support.ParseRequest(ctx.Request, &form)
	if err != nil {
		ctx.Back()
		return
	}

	err = s.voidInvoice(ctx.Request.Context(), ctx.Param("uuid"))
	if err != nil {
		log.Printf("Error voiding invoice: %v", err)
		s.session.Errors("status", s.trans("global.wasNotVoided", i18n.Replacements{"subject": "@global.invoice"}))
		ctx.Back()
		return
	}
	s.session.Flash("success", s.trans("global.wasVoided", i18n.Replacements{"subject": "@global.invoice"}))

	ctx.Redirect("/invoices")
}
