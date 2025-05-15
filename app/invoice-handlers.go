package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/support"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) invoicesHandler(i *inertia.Inertia) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {

		uuid := r.URL.Query().Get("id")
		user := auth.User(r.Context())
		invoices, err := s.findInvoices(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		props := inertia.Props{
			"translations": mergeTranslations(r.Context(), loadTranslations("invoices")),
			"invoices":     invoices,
		}

		if ensureUUIDIsValid(uuid) {
			invoice, err := s.findInvoicesByUUID(*user.CurrentCompanyId, uuid)
			if err != nil {
				s.handleError(w, err)
				return
			}

			lines, err := s.findInvoiceLines(*user.CurrentCompanyId, invoice.ID)
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
		user := auth.User(r.Context())
		taxReceipts, err := s.findTaxesReceipts(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Invoices/Create", inertia.Props{
			"tax_receipts": mapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
				return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name), "available": receipt.Current < receipt.SequenceEnd}
			}),
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(*user.CurrentCompanyId, term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"item": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByReference(*user.CurrentCompanyId, term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
			"items": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByCriteria(*user.CurrentCompanyId, term)
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
		user := auth.User(r.Context())

		invoice, err := s.findInvoicesByUUID(*user.CurrentCompanyId, invoiceUUID)
		if err != nil {
			s.handleError(w, err)
			return
		}

		lines, err := s.findInvoiceLines(*user.CurrentCompanyId, invoice.ID)
		if err != nil {
			s.handleError(w, err)
		}

		taxReceipts, err := s.findTaxesReceipts(*user.CurrentCompanyId)
		if err != nil {
			s.handleError(w, err)
			return
		}

		err = i.Render(w, r, "Invoices/Edit", inertia.Props{
			"invoice": map[string]any{
				"header": invoice,
				"lines":  lines,
			},
			"tax_receipts": mapSlice(taxReceipts, func(receipt *taxReceipt) map[string]any {
				return map[string]any{"id": receipt.ID, "name": fmt.Sprintf("%s-%s", receipt.Type, receipt.Name)}
			}),
			"customers": inertia.Optional(func() (any, error) {
				customers, err := s.findCustomersBySearchCriteria(*user.CurrentCompanyId, term)
				if err != nil {
					return nil, err
				}

				return customers, err
			}),
			"item": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByReference(*user.CurrentCompanyId, term)
				if err != nil {
					return nil, err
				}

				return item, err
			}),
			"items": inertia.Optional(func() (any, error) {
				item, err := s.findItemsByCriteria(*user.CurrentCompanyId, term)
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

		user := auth.User(r.Context())
		err = s.storeInvoice(*user.CurrentCompanyId, form)
		if err != nil {
			log.Printf("Error creating invoice: %v", err)
			s.session.Errors("status", "Invoice wasn't created. Something went wrong.")
			i.Back(w, r)
			return
		}
		s.session.Flash("success", "Invoice was created successfully!")

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

		user := auth.User(r.Context())
		err = s.updateInvoice(*user.CurrentCompanyId, uuid, form)
		if err != nil {
			log.Printf("Error updating invoice: %v", err)
			s.session.Errors("status", "Invoice wasn't updated. Something went wrong.")
			back()
			return
		}
		s.session.Flash("success", "Invoice was created successfully!")

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

		user := auth.User(r.Context())
		err = s.voidInvoice(*user.CurrentCompanyId, uuid)
		if err != nil {
			log.Printf("Error voiding invoice: %v", err)
			s.session.Errors("status", "Invoice wasn't voided. Something went wrong.")
			back()
			return
		}
		s.session.Flash("success", "Invoice was voided successfully!")

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
