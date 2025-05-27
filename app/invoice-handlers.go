package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) invoicesHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		uuid := r.URL.Query().Get("id")
		invoices, err := s.findInvoices(r.Context())
		if err != nil {
			s.handleError(w, err)
			return
		}

		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("invoices")),
			"invoices":     invoices,
		}

		if ensureUUIDIsValid(uuid) {
			invoice, err := s.findInvoicesByUUID(r.Context(), uuid)
			if err != nil {
				s.handleError(w, err)
				return
			}

			lines, err := s.findInvoiceLines(r.Context(), invoice.ID)
			if err != nil {
				s.handleError(w, err)
				return
			}

			props["invoice"] = map[string]any{
				"header": invoice,
				"lines":  lines,
			}
			props["showInvoice"] = true
		}

		err = i.Render(w, r, "Invoices/Index", props)
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) createInvoiceHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		term := r.URL.Query().Get("search")
		taxReceipts, err := s.findTaxesReceipts(r.Context())
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Invoices/Create", inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("invoices")),
			"tax_receipts": mapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
				return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name), "available": receipt.Current < receipt.SequenceEnd}
			}),
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"item": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByReference(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
			"items": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByCriteria(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) editInvoiceHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		invoiceUUID := r.PathValue("id")
		term := r.URL.Query().Get("search")

		invoice, err := s.findInvoicesByUUID(r.Context(), invoiceUUID)
		if err != nil {
			s.handleError(w, err)
			return
		}

		lines, err := s.findInvoiceLines(r.Context(), invoice.ID)
		if err != nil {
			s.handleError(w, err)
		}

		taxReceipts, err := s.findTaxesReceipts(r.Context())
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Invoices/Edit", inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("invoices")),
			"invoice": map[string]any{
				"header": invoice,
				"lines":  lines,
			},
			"tax_receipts": mapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
				return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name)}
			}),
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"item": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByReference(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
			"items": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByCriteria(r.Context(), term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
		})
		if err != nil {
			s.handleError(w, err)
			return
		}
	}

	return http.HandlerFunc(fn)
}

func (s *Server) storeInvoiceHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		var form StoreInvoiceForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			i.Back(w, r)
			return
		}

		if form.Terms == 1 && form.total != form.paymentTotalAmount() {
			s.session.Errors("status", "Invoice total amount and the payment details are different.")
			i.Back(w, r)
			return
		}

		err = s.storeInvoice(r.Context(), form)
		if err != nil {
			log.Printf("Error creating invoice: %v", err)
			s.session.Errors("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.invoice"}))
			i.Back(w, r)
			return
		}
		s.session.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.invoice"}))

		i.Back(w, r)
	}

	return http.HandlerFunc(fn)
}

func (s *Server) updateInvoiceHandler(i *inertia.Inertia) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("id")
		back := func() {
			i.Back(w, r, http.StatusSeeOther)
		}
		var form UpdateInvoiceForm
		err := support.ParseRequest(r, &form)
		if err != nil {
			back()
			return
		}

		if form.Terms == 1 && form.total != form.paymentTotalAmount() {
			s.session.Errors("status", "Invoice total amount and the payment details are different.")
			back()
			return
		}

		err = s.updateInvoice(r.Context(), uuid, form)
		if err != nil {
			log.Printf("Error updating invoice: %v", err)
			s.session.Errors("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.invoice"}))
			back()
			return
		}
		s.session.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.invoice"}))

		i.Redirect(w, r, "/invoices", http.StatusSeeOther)
	}

	return http.HandlerFunc(fn)
}

func (s *Server) voidInvoiceHandler(i *inertia.Inertia) http.Handler {

	fn := func(w http.ResponseWriter, r *http.Request) {
		uuid := r.PathValue("id")
		back := func() {
			i.Back(w, r, http.StatusSeeOther)
		}
		var form ConfirmsPasswords
		err := support.ParseRequest(r, &form)
		if err != nil {
			back()
			return
		}

		err = s.voidInvoice(r.Context(), uuid)
		if err != nil {
			log.Printf("Error voiding invoice: %v", err)
			s.session.Errors("status", s.trans("global.wasNotVoided", i18n.Replacements{"subject": "@global.invoice"}))
			back()
			return
		}
		s.session.Flash("success", s.trans("global.wasVoided", i18n.Replacements{"subject": "@global.invoice"}))

		i.Redirect(w, r, "/invoices", http.StatusSeeOther)
	}

	return http.HandlerFunc(fn)
}

func mapSlice[T, U any](s []T, f func(T) U) []U {
	result := make([]U, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}
