package app

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func resolveTransactionKind(ctx *routing.Context) TransactionKind {
	kind := ctx.Request.Header.Get("X-Transaction-Kind")
	if kind != "" {
		return TransactionKind(kind)
	}

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
		company := CurrentCompany(ctx.Request.Context())
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:%s:%s", kind, uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
			invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), kind, company.ID, uuid)
			if err != nil {
				return nil, err
			}

			lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.CompanyID, invoice.ID)
			if err != nil {
				return nil, err
			}

			uri := fmt.Sprintf("%s/%ss/%s/print/%s", s.config.host, kind, uuid, foundation.NewHashable().Sha1(uuid))
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

func (s *Server) showInvoiceHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)
	uuid := ctx.Param("id")
	if uuid == "" {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"status": fmt.Sprintf("The %s given ID is not valid.", kind),
		})
		return
	}
	company := CurrentCompany(ctx.Request.Context())
	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:%s:%s", kind, uuid)
	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
		invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), kind, company.ID, uuid)
		if err != nil {
			return nil, err
		}

		lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.CompanyID, invoice.ID)
		if err != nil {
			return nil, err
		}

		return map[string]any{
			"header": invoice,
			"lines":  lines,
		}, nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"status": fmt.Sprintf("An error retrieving the %s.", kind),
			"data":   err,
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (s *Server) createInvoiceHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)

	if s.abortWhenPrerequisiteMissing(ctx, string(kind)) {
		return
	}

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

	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	props["tax_receipts"] = foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
		return map[string]any{
			"id":        receipt.ID,
			"name":      fmt.Sprintf("%s-%s", receipt.Serie, receipt.Name),
			"available": receipt.Current < receipt.SequenceEnd,
		}
	})
	props["showPaymentCTA"] = true

	ctx.Render("Invoices/Create", props)
}

func (s *Server) editInvoiceHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)
	company := CurrentCompany(ctx.Request.Context())
	invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), kind, company.ID, ctx.Param("id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.CompanyID, invoice.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	term := ctx.Query("search")
	props := map[string]any{
		"translations": trans(fmt.Sprintf("%ss", kind)),
		"kind":         kind,
		"invoice": map[string]any{
			"header": invoice,
			"lines":  lines,
		},
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

	taxReceipts, err := s.findTaxesReceipts(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	props["tax_receipts"] = foundation.MapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
		return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name)}
	})

	ctx.Render("Invoices/Edit", props)
}

func (s *Server) storeInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreInvoiceForm) {

		if form.Kind == TransactionKinds.Invoice {
			if form.Terms == "pia" && form.total != form.paymentTotalAmount() {
				ctx.BackWith("status", s.trans("@invoices.totalAmountDiffs"))
				return
			}
			if form.Terms != "pia" {
				customer, err := s.findCustomeByID(ctx.Request.Context(), form.CustomerID)
				if err != nil {
					ctx.BackWithError(err)
					return
				}

				if customer.CreditLimited && customer.AmountDue+form.total > customer.CreditLimit {
					ctx.BackWith("status", s.trans("@invoices.creditLimitExceeded"))
					return
				}
			}
		}

		id, err := s.storeInvoice(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating %s: %v", form.Kind, err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": fmt.Sprintf("@global.%s", form.Kind)}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": fmt.Sprintf("@global.%s", form.Kind)}))
		ctx.Flash("redirectTo", fmt.Sprintf("/%ss", form.Kind))

		preferences, err := s.findRedirectPreferences(ctx.Request.Context(), CurrentCompany(ctx.Request.Context()).UUID)
		if err != nil {
			log.Printf("Error fetching redirect preferences: %v", err)
			ctx.Back()
			return
		}

		switch form.Kind {
		case TransactionKinds.Invoice, TransactionKinds.Template:
			ctx.Flash("redirectTo", redirectAfterCreate("invoice", id, preferences.Redirect.Invoice))
		case TransactionKinds.Estimate:
			ctx.Flash("redirectTo", redirectAfterCreate(string(form.Kind), id, preferences.Redirect.Estimate))
		case TransactionKinds.Order:
			ctx.Flash("redirectTo", redirectAfterCreate(string(form.Kind), id, preferences.Redirect.Order))
		default:
		}

		ctx.Back()
	})
}

func (s *Server) updateInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateInvoiceForm) {

		if form.Kind == TransactionKinds.Invoice {
			if form.Terms == "pia" && form.total != form.paymentTotalAmount() {
				ctx.BackWith("status", s.trans("@invoices.totalAmountDiffs"))
				return
			}

			customer, err := s.findCustomeByID(ctx.Request.Context(), form.CustomerID)
			if err != nil {
				ctx.BackWithError(err)
				return
			}

			if customer.CreditLimited && customer.AmountDue+form.total > customer.CreditLimit {
				ctx.BackWith("status", s.trans("@invoices.creditLimitExceeded"))
				return
			}
		}

		uuid := ctx.Param("id")
		err := s.updateInvoice(ctx.Request.Context(), uuid, form)
		if err != nil {
			log.Printf("Error updating %s: %v", form.Kind, err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.invoice"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:%s:%s", form.Kind, uuid)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.invoice"}))

		ctx.Redirect("/invoices")
	})
}

func (s *Server) voidInvoiceHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		kind := resolveTransactionKind(ctx)
		uuid := ctx.Param("id")

		err := s.voidInvoice(ctx.Request.Context(), kind, uuid)
		if err != nil {
			log.Printf("Error voiding %s: %v", kind, err)
			ctx.BackWith("status", s.trans("global.wasNotVoided", i18n.Replacements{"subject": "@global." + string(kind)}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:%s:%s", kind, uuid)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasVoided", i18n.Replacements{"subject": "@global." + string(kind)}))

		ctx.Redirect("/invoices")
	})
}

func (s *Server) markInvoiceAsRecurrentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreRecurrenceForm) {
		company := CurrentCompany(ctx.Request.Context())
		invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), TransactionKinds.Invoice, company.ID, ctx.Param("id"))
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": err.Error(),
			})
			return
		}

		lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.CompanyID, invoice.ID)
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": err.Error(),
			})
			return
		}

		invoiceForm := mapInvoiceToStoreForm(invoice, lines)
		invoiceForm.Recurrence = form.AsRecurrence()

		_, err = s.storeInvoice(ctx.Request.Context(), invoiceForm)
		if err != nil {
			log.Printf("Error creating recurring invoice: %v", err)
			ctx.JSON(http.StatusInternalServerError, map[string]any{
				"status": s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.template"}),
			})
			return
		}
		ctx.JSON(http.StatusOK, map[string]any{
			"message": s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.template"}),
		})

	})
}

func (s *Server) printInvoiceHandler(ctx *routing.Context) {
	kind := resolveTransactionKind(ctx)

	uuid := ctx.Param("id")
	hash := ctx.Param("hash")
	if !foundation.NewHashable().Sha1Equals(uuid, hash) {
		ctx.BackWith("status", s.trans("global.mismatch", i18n.Replacements{"subject": "HASH"}))
		return
	}

	type invoiceData struct {
		Header *invoice `json:"header"`
		Lines  []*line  `json:"lines"`
	}
	company := CurrentCompany(ctx.Request.Context())
	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:%s:%s", kind, uuid)
	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (invoiceData, error) {
		invoice, err := s.findInvoicesByUUID(ctx.Request.Context(), kind, company.ID, uuid)
		if err != nil {
			return invoiceData{}, nil
		}

		lines, err := s.findInvoiceLines(ctx.Request.Context(), invoice.CompanyID, invoice.ID)
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
		return
	}

	invoice.Header(company)
	invoice.Lines()
	invoice.Footer(s.config.appName)

	ctx.Response.Header().Set("Content-Type", "application/pdf")
	ctx.Response.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=%s.pdf", data.Header.Number))

	if err := invoice.Output(ctx.Response); err != nil {
		ctx.Error(err)
	}
}

func decodeLogo() ([]byte, error) {
	cleaned := strings.TrimSpace(string(logoBase64))

	if cleaned == "" {
		return nil, errors.New("logo base64 is empty")
	}

	data, err := base64.StdEncoding.DecodeString(cleaned)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo: %w", err)
	}

	return data, nil
}
