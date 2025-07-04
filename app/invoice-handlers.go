package app

import (
	"fmt"
	"log"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
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

	if uuid != "" {
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

		uri := fmt.Sprintf("%s/invoices/%s/print/%s", s.config.host, uuid, foundation.NewHashable().Sha1(uuid))
		pdfURL, err := routing.TemporarySignedURL(uri, map[string]string{}, string(s.config.secretKey), 5*time.Minute)
		if err != nil {
			ctx.Error(err)
			return
		}

		props["invoice"] = map[string]any{
			"header": invoice,
			"lines":  lines,
			"pdfURL": pdfURL,
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

	var customer *customer
	customerID := ctx.Query("customer_id")
	if strings.TrimSpace(customerID) != "" {
		customer, err = s.findCustomeByUUID(ctx.Request.Context(), customerID)
		if err != nil {
			ctx.Error(err)
			return
		}
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
		"customer": customer,
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

		err = s.storeInvoice(ctx.Request.Context(), form)
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

	// uuid := ctx.Param("id")
	// hash := ctx.Param("hash")
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(15, 20, 15)
	pdf.AddPage()

	// Company Info
	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(0, 10, "INVOICE")
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(100, 6, "From:")
	pdf.Cell(0, 6, "To:")
	pdf.Ln(6)
	pdf.Cell(100, 6, "Acme Corp")
	pdf.Cell(0, 6, "John Doe")
	pdf.Ln(6)
	pdf.Cell(100, 6, "123 Business Rd.")
	pdf.Cell(0, 6, "456 Client St.")
	pdf.Ln(6)
	pdf.Cell(100, 6, "New York, NY")
	pdf.Cell(0, 6, "Los Angeles, CA")
	pdf.Ln(12)

	// Invoice Info
	pdf.Cell(40, 6, "Invoice #: 001")
	pdf.Cell(60, 6, "")
	pdf.Cell(40, 6, "Date:")
	pdf.Cell(0, 6, time.Now().Format("2006-01-02"))
	pdf.Ln(10)

	// Table Header
	pdf.SetFillColor(240, 240, 240)
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(80, 8, "Description", "1", 0, "", true, 0, "")
	pdf.CellFormat(30, 8, "Qty", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Unit Price", "1", 0, "R", true, 0, "")
	pdf.CellFormat(40, 8, "Total", "1", 1, "R", true, 0, "")

	// Dummy Data
	items := []struct {
		Description string
		Quantity    int
		UnitPrice   float64
	}{
		{"Website Design", 1, 1200.00},
		{"Hosting (12 months)", 1, 240.00},
		{"Domain (1 year)", 1, 15.00},
	}

	pdf.SetFont("Arial", "", 12)
	var subtotal float64
	for _, item := range items {
		lineTotal := float64(item.Quantity) * item.UnitPrice
		subtotal += lineTotal

		pdf.CellFormat(80, 8, item.Description, "1", 0, "", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", lineTotal), "1", 1, "R", false, 0, "")
	}

	// Totals
	taxRate := 0.18
	tax := subtotal * taxRate
	total := subtotal + tax

	pdf.Ln(4)
	pdf.Cell(150, 8, "Subtotal:")
	pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", subtotal), "", 1, "R", false, 0, "")
	pdf.Cell(150, 8, "Tax (18%):")
	pdf.CellFormat(40, 8, fmt.Sprintf("$%.2f", tax), "", 1, "R", false, 0, "")
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(150, 10, "Total:")
	pdf.CellFormat(40, 10, fmt.Sprintf("$%.2f", total), "", 1, "R", false, 0, "")

	// Output
	ctx.Response.Header().Set("Content-Type", "application/pdf")
	ctx.Response.Header().Set("Content-Disposition", "inline; filename=invoice.pdf")

	if err := pdf.Output(ctx.Response); err != nil {
		ctx.Error(err)
	}
}
