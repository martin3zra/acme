package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) expensesHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "expense") {
		return
	}

	expenses, err := s.findExpenses(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	categories, err := s.findExpensesCategories(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"openState":    ctx.Query("mode") == "creating",
		"translations": trans("expenses"),
		"expenses":     expenses,
		"categories":   categories,
	}

	// uuid := ctx.Query("id")
	// if uuid != "" {
	// 	c := cache.NewPgCache(s.db)
	// 	key := fmt.Sprintf("preview:expense:%s", uuid)
	// 	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
	// 		payment, err := s.findPaymentByUUID(ctx.Request.Context(), uuid)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		lines, err := s.findPaymentLines(ctx.Request.Context(), payment.ID)
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		uri := fmt.Sprintf("%s/expenses/%s/print/%s", s.config.host, uuid, foundation.NewHashable().Sha1(uuid))
	// 		pdfURL, err := routing.PermanentSignedURL(uri, map[string]string{}, string(s.config.secretKey))
	// 		if err != nil {
	// 			return nil, err
	// 		}

	// 		return map[string]any{
	// 			"header": payment,
	// 			"lines":  lines,
	// 			"pdfURL": pdfURL,
	// 		}, nil
	// 	})
	// 	if err != nil {
	// 		ctx.Error(err)
	// 		return
	// 	}
	// 	props["expense"] = data
	// 	props["showExpense"] = true
	// }

	ctx.Render("Expenses/Index", props)
}

func (s *Server) storeExpenseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreExpenseForm) {

		err := s.storeExpense(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error recording expense: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.expense"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.expense"}))

		ctx.Redirect("/expenses")
	})
}
