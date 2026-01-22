package app

import (
	"fmt"
	"log"
	"strings"

	"github.com/martin3zra/acme/pkg/cache"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func resolveTransactionKind(ctx *routing.Context) TransactionKind {
	path := ctx.Request.URL.Path
	switch {
	case strings.HasPrefix(path, "/estimates"):
		return TransactionKinds.Estimate
	case strings.HasPrefix(path, "/orders"):
		return TransactionKinds.Order
	default:
		return TransactionKinds.Invoice
	}
}

func (s *Server) invoicesHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)
	if s.abortWhenPrerequisiteMissing(ctx, string(kind)) {
		return
	}

	invoiceType := InvoiceType(ctx.Query("invoiceType"))
	if err := invoiceType.Validate(); err != nil {
		invoiceType = "all"
	}

	uuid := ctx.Query("id")
	invoices, err := s.findInvoices(ctx.Request.Context(), kind, invoiceType)
	if err != nil {
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"translations":             trans(fmt.Sprintf("%ss", kind)),
		"invoices":                 invoices,
		"currentInvoiceTypeFilter": invoiceType,
		"kind":                     kind,
	}

	if uuid != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:invoice:%s", uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
			invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), uuid)
			if err != nil {
				return nil, err
			}

			lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.ID)
			if err != nil {
				return nil, err
			}

			uri := fmt.Sprintf("%s/invoices/%s/print/%s", s.config.host, uuid, foundation.NewHashable().Sha1(uuid))
			pdfURL, err := routing.PermanentSignedURL(uri, map[string]string{}, string(s.config.secretKey))
			if err != nil {
				return nil, nil
			}

			return map[string]any{
				"header": invoice,
				"lines":  lines,
				"pdfURL": pdfURL,
			}, nil
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["invoice"] = data
		props["showInvoice"] = true
	}

	ctx.Render("Invoices/Index", props)
}

func (s *Server) createInvoiceHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)
	term := ctx.Query("search")

	var customer *customer
	var err error
	customerID := ctx.Query("customer_id")
	if strings.TrimSpace(customerID) != "" {
		customer, err = s.findCustomeByUUID(ctx.Request.Context(), customerID)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	props := map[string]any{
		"translations": trans(fmt.Sprintf("%ss", kind)),
		"kind":         kind,
		"customer":     customer,
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
	}

	if kind == TransactionKinds.Invoice {
		taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}
		props["tax_receipts"] = foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
			return map[string]any{
				"id":        receipt.ID,
				"name":      fmt.Sprintf("%s-%s", receipt.Type, receipt.Name),
				"available": receipt.Current < receipt.SequenceEnd,
			}
		})
		props["showPaymentCTA"] = true
	}
	ctx.Render("Invoices/Create", props)
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

		if form.Kind == TransactionKinds.Invoice {
			if form.Terms == "pia" && form.total != form.paymentTotalAmount() {
				ctx.BackWith("status", "Invoice total amount and the payment details are different.")
				return
			}

			customer, err := s.findCustomeByID(ctx.Request.Context(), form.CustomerID)
			if err != nil {
				ctx.BackWithError(err)
				return
			}

			if customer.CreditLimited && customer.AmountDue+form.total > customer.CreditLimit {
				ctx.BackWith("status", "Credit limit exceeded.")
				return
			}
		}

		err := s.storeInvoice(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating invoice: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.invoice"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.invoice"}))
		ctx.Flash("redirectTo", "/invoices")

		preferences, err := s.findRedirectPreferences(ctx.Request.Context(), CurrentCompany(ctx.Request.Context()).UUID)
		if err != nil {
			log.Printf("Error fetching redirect preferences: %v", err)
			ctx.Back()
			return
		}

		ctx.Flash("redirectTo", preferences.Redirect.Invoice)

		ctx.Back()
	})
}

func (s *Server) updateInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateInvoiceForm) {

		if form.Terms == "pia" && form.total != form.paymentTotalAmount() {
			ctx.BackWith("status", "Invoice total amount and the payment details are different.")
			return
		}

		customer, err := s.findCustomeByID(ctx.Request.Context(), form.CustomerID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		if customer.CreditLimited && customer.AmountDue+form.total > customer.CreditLimit {
			ctx.BackWith("status", "Credit limit exceeded.")
			return
		}

		err = s.updateInvoice(ctx.Request.Context(), ctx.Param("id"), form)
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

func (s *Server) printInvoiceHandler(ctx *routing.Context) {

	uuid := ctx.Param("id")
	hash := ctx.Param("hash")
	if !foundation.NewHashable().Sha1Equals(uuid, hash) {
		ctx.BackWith("status", "The hash does not match")
		return
	}

	type invoiceData struct {
		Header *invoice `json:"header"`
		Lines  []*line  `json:"lines"`
	}
	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:invoice:%s", uuid)
	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (invoiceData, error) {

		invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), uuid)
		if err != nil {
			return invoiceData{}, nil
		}

		lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.ID)
		if err != nil {
			return invoiceData{}, nil
		}
		return invoiceData{
			Header: invoice,
			Lines:  lines,
		}, nil
	})

	invoice, err := NewInvoicePDF(s.translator, data.Header, data.Lines)
	if err != nil {
		ctx.Error(err)
	}

	invoice.Header(ctx.Request.Context())
	invoice.Lines()
	invoice.Footer(s.config.appName)

	ctx.Response.Header().Set("Content-Type", "application/pdf")
	ctx.Response.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", data.Header.Number))

	if err := invoice.Output(ctx.Response); err != nil {
		ctx.Error(err)
	}
}

const logoBase64 = "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ECBA4ERA0Dzh8IAAAQIECBAgQKCWgKBZS9J+CBAgQIAAAQIETgQEzRMODwgQIECAAAECBGoJCJq1JO2HAAECBAgQIEDgREDQPOHwgAABAgQIECBAoJaAoFlL0n4IECBAgAABAgROBATNEw4PCBAgQIAAAQIEagkImrUk7YcAAQIECBAgQOBEQNA84fCAAAECBAgQIECgloCgWUvSfggQIECAAAECBE4EBM0TDg8IECBAgAABAgRqCQiatSTthwABAgQIECBA4ERA0Dzh8IAAAQIECBAgQKCWgKBZS9J+CBAgQIAAAQIETgQEzRMODwgQIECAAAECBGoJCJq1JO2HAAECBAgQIEDgREDQPOHwgAABAgQIECBAoJaAoFlL0n4IECBAgAABAgROBATNEw4PCBAgQIAAAQIEagkImrUk7YcAAQIECBAgQOBEQNA84fCAAAECBAgQIECgloCgWUvSfggQIECAAAECBE4EBM0TDg8IECBAgAABAgRqCQiatSTthwABAgQIECBA4ERA0Dzh8IAAAQIECBAgQKCWgKBZS9J+CBAgQIAAAQIETgQEzRMODwgQIECAAAECBGoJCJq1JO2HAAECBAgQIEDgREDQPOHwgAABAgQIECBAoJaAoFlL0n4IECBAgAABAgROBATNEw4PCBAgQIAAAQIEagkImrUk7YcAAQIECBAgQOBEQNA84fCAAAECBAgQIECgRODFixePqgmaj0g8QYAAAQIECBAgMFTg7u7uURVB8xGJJwgQIECAAAECBGoICJo1FO2DAAECBAgQIEDgkYCg+YjEEwQIECBAgAABAkMFXKM5VMz2BAgQIECAAAECNwm4RvMmJhsRIECAAAECBAgMFbCiOVTM9gQIECBAgAABAjcJWNG8iclGBAgQIECAAAECQwWsaA4Vsz0BAgQIECBAgMBNAlY0b2KyEQECBAgQIECAwFABK5pDxWxPgAABAgQIECBQLOB7NIvpVCRAgAABAgQIEMgJCJo5HWUECBAgQIAAAQLFAoJmMZ2KBAgQIECAAAECOQFBM6ejjAABAgQIECBAoFhA0CymU5EAAQIECBAgQCAnIGjmdJQRIECAAAECBAgUCwiaxXQqEiBAgAABAgQI5AQEzZyOMgIECBAgQIAAgWIBQbOYTkUCBAgQIECAAIGcgKCZ01FGgAABAgQIECBQLCBoFtOpSIAAAQIECBAgkBMQNHM6yggQIECAAAECBIoFBM1iOhUJECBAgAABAgRyAoJmTkcZAQIECBAgQIBAsYCgWUynIgECBAgQIECAQE5A0MzpKCNAgAABAgQIECgWEDSL6VQkQIAAAQIECBDICQiaOR1lBAgQIECAAAECxQKCZjGdigQIECBAgAABAjkBQTOno4wAAQIECBAgQKBYQNAsplORAAECBAgQIEAgJyBo5nSUESBAgAABAgQIFAsImsV0KhIgQIAAAQIECOQEBM2cjjICBAgQIECAAIFiAUGzmE5FAgQIECBAgACBnICgmdNRRoAAAQIECBAgUCwgaBbTqUiAAAECBAgQIJATEDRzOsoIECBAgAABAgSKBQTNYjoVCRAgQIAAAQIEcgKCZk5HGQECBAgQIECAQLGAoFlMpyIBAgQIECBAgEBOQNDM6SgjQIAAAQIECBAoFhA0i+lUJECAAAECBAgQyAkImjkdZQQIECBAgAABAsUCgmYxnYoECBAgQIAAAQI5AUEzp6OMAAECBAgQIECgWEDQLKZTkQABAgQIECBAICcgaOZ0lBEgQIAAAQIECBQLCJrFdCoSIECAAAECBAjkBATNnI4yAgQIECBAgACBYgFBs5hORQIECBAgQIAAgZyAoJnTUUaAAAECBAgQIFAsIGgW06lIgAABAgQIECCQExA0czrKCBAgQIAAAQIEigUEzWI6FQkQIECAAAECBHICgmZORxkBAgQIECBAgECxgKBZTKciAQIECBAgQIBATkDQzOkoI0CAAAECBAgQKBYQNIvpVCRAgAABAgQIEMgJCJo5HWUECBAgQIAAAQLFAoJmMZ2KBAgQIECAAAECOQFBM6ejjAABAgQIECBAoFhA0CymU5EAAQIECBAgQCAnIGjmdJQRIECAAAECBAgUCwiaxXQqEiBAgAABAgQI5AQEzZyOMgIECBAgQIAAgWIBQbOYTkUCBAgQIECAAIGcgKCZ01FGgAABAgQIECBQLCBoFtOpSIAAAQIECBAgkBMQNHM6yggQIECAAAECBIoFBM1iOhUJECBAgAABAgRyAoJmTkcZAQIECBAgQIBAsYCgWUynIgECBAgQIECAQE5A0MzpKCNAgAABAgQIECgWEDSL6VQkQIAAAQIECBDICQiaOR1lBAgQIECAAAECxQKCZjGdigQIECBAgAABAjkBQTOno4wAAQIECBAgQKBYQNAsplORAAECBAgQIEAgJyBo5nSUESBAgAABAgQIFAsImsV0KhIgQIAAAQIECOQEBM2cjjICBAgQIECAAIFiAUGzmE5FAgQIECBAgACBnICgmdNRRoAAAQIECBAgUCwgaBbTqUiAAAECBAgQIJATEDRzOsoIECBAgAABAgSKBQTNYjoVCRAgQIAAAQIEcgKCZk5HGQECBAgQIECAQLGAoFlMpyIBAgQIECBAgEBOQNDM6SgjQIAAAQIECBAoFhA0i+lUJECAAAECBAgQyAkImjkdZQQIECBAgAABAsUCgmYxnYoECBAgQIAAAQI5AUEzp6OMAAECBAgQIECgWEDQLKZTkQABAgQIECBAICcgaOZ0lBEgQIAAAQIECBQLCJrFdCoSIECAAAECBAjkBATNnI4yAgQIECBAgACBYgFBs5hORQIECBAgQIAAgZyAoJnTUUaAAAECBAgQIFAsIGgW06lIgAABAgQIECCQExA0czrKCBAgQIAAAQIEigUEzWI6FQkQIECAAAECBHICgmZORxkBAgQIECBAgECxgKBZTKciAQIECBAgQIBATkDQzOkoI0CAAAECBAgQKBYQNIvpVCRAgAABAgQIEMgJCJo5HWUECBAgQIAAAQLFAoJmMZ2KBAgQIECAAAECOYEuaH788ce57ZQRIECAAAECBAgQ6BX41Ke6WNlt0z3zmc98pnvSDwQIECBAgAABAgSGCPz6179+tHkXNF9//fXDr371q0cbeIIAAQIECBAgQIBATuCjjz46fOELX3i0SRc0v/KVrxyeP3/+aANPECBAgAABAgQIEMgJxCWYX/7ylx9t0gXNKPniF79oVfMRkScIECBAgAABAgT6BN5///3DpeszY/uToBmrmh9++KGw2SfpeQIECBAgQIAAgU4g3jL/5JNPLq5mxkZ3L+5v3db//8NPf/rTw2uvvXZ4+vTpeZHHBAgQIECAAAECBA4RMmOB8mtf+1qvxsWgGVt/8MEHh9/97neHN9544/D5z3++dwcKCBAgQIAAAQIE9iMQAfPnP//5Qz68dF3msURv0EwbReCMtBphM4KnW/sCd3d37TdSCwncC5irpgEBAucCXhfORdp5HB/4iW8p+vSnP937Vvl5a68GzfMKHhMgQIAAAQIECBC4ReDkw0C3VLANAQIECBAgQIAAgVsEBM1blGxDgAABAgQIECAwWEDQHEymAgECBAgQIECAwC0CguYtSrYhQIAAAQIECBAYLPB/cxzkjJqmG3oAAAAASUVORK5CYII="
