package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) paymentsHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "payment") {
		return
	}

	payments, err := s.findPayments(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"translations": trans("payments"),
		"payments":     payments,
	}

	uuid := ctx.Query("id")
	if uuid != "" {
		payment, err := s.findPaymentByUUID(ctx.Request.Context(), uuid)
		if err != nil {
			ctx.Error(err)
			return
		}

		lines, err := s.findPaymentLines(ctx.Request.Context(), payment.ID)
		if err != nil {
			ctx.Error(err)
			return
		}
		props["payment"] = map[string]any{
			"header": payment,
			"lines":  lines,
		}
		props["showPayment"] = true
	}

	ctx.Render("Payments/Index", props)
}

func (s *Server) createPaymentHandler(ctx *routing.Context) {

	// what we do when the Payment id is given? only return that Payments?
	paymentUuid := ctx.Query("payment_id")
	term := ctx.Query("search")
	customerUuid := ctx.Query("customer_id")
	invoiceUuid := ctx.Query("invoice_id")

	props := map[string]any{
		"translations": trans("payments"),
		"customers": inertia.Optional(func() (any, error) {
			customers, err := s.findCustomersBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}

			return customers, err
		}),
		"receivables": inertia.Optional(func() (any, error) {
			receivables, err := s.findCustomeReceivables(ctx.Request.Context(), customerUuid)
			if err != nil {
				return nil, err
			}

			return receivables, err
		}),
	}

	if paymentUuid != "" {
		props["payment_uuid"] = paymentUuid
	}

	if customerUuid != "" {
		customer, err := s.findCustomeByUUID(ctx.Request.Context(), customerUuid)
		if err != nil {
			ctx.Error(err)
			return
		}
		receivables, err := s.findCustomeReceivables(ctx.Request.Context(), customerUuid)
		if err != nil {
			ctx.Error(err)
			return
		}

		props["customer"] = customer
		props["receivables"] = receivables
		props["invoice_uuid"] = invoiceUuid
		props["forceInitial"] = true
	}
	ctx.Render("Payments/Create", props)
}

func (s *Server) storePaymentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StorePaymentForm) {

		err := s.storePayment(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error recording payment: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.payment"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.payment"}))

		ctx.Redirect("/payments")
	})
}

func (s *Server) voidPaymentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.voidPayment(ctx.Request.Context(), ctx.Param("id"))
		if err != nil {
			log.Printf("Error voiding payment: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotVoided", i18n.Replacements{"subject": "@global.payment"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasVoided", i18n.Replacements{"subject": "@global.payment"}))

		ctx.Redirect("/payments")
	})
}

func (s *Server) editPaymentHandler(ctx *routing.Context) {

	payment, err := s.findPaymentByUUID(ctx.Request.Context(), ctx.Param("id"))
	if err != nil {
		ctx.Error(err)
		return
	}

	lines, err := s.findPaymentLines(ctx.Request.Context(), payment.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Payments/Edit", map[string]any{
		"translations": trans("payments"),
		"payment": map[string]any{
			"header": payment,
			"lines":  lines,
		},
		"showPayment": true,
	})
}

func (s *Server) updatePaymentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdatePaymentForm) {

		err := s.updatePayment(ctx.Request.Context(), ctx.Param("id"), form)
		if err != nil {
			log.Printf("Error recording payment: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.payment"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.payment"}))

		ctx.Redirect("/payments")
	})
}
