package app

import (
	"fmt"
	"log"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/forge/routing"
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

	uuid := ctx.Query("id")
	if uuid != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:expense:%s", uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (*expense, error) {
			return s.findExpenseByUUID(ctx.Request.Context(), uuid)
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["expense"] = data
		props["openState"] = true
	}

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

func (s *Server) deleteExpenseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {

		err := s.deleteExpense(ctx.Request.Context(), ctx.Param("id"))
		if err != nil {
			log.Printf("Error deleting expense: %v", err)
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.expense"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.expense"}))

		ctx.Redirect("/expenses")
	})
}

func (s *Server) updateExpenseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreExpenseForm) {

		err := s.updateExpense(ctx.Request.Context(), ctx.Param("id"), form)
		if err != nil {
			log.Printf("Error updating expense: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.expense"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:expense:%s", ctx.Param("id"))
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.expense"}))

		ctx.Redirect("/expenses")
	})
}
