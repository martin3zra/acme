package app

import (
	"log"

	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
)

func (s *Server) storeExpenseCategoryHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreExpenseCategoryForm) {

		if err := s.storeExpenseCategory(ctx.Request.Context(), form); err != nil {
			log.Printf("Error creating expense category: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.category"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.category"}))
		ctx.Back()
	})
}

func (s *Server) updateExpenseCategoryHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreExpenseCategoryForm) {

		uuid := ctx.Param("id")
		if err := s.updateExpenseCategory(ctx.Request.Context(), uuid, form); err != nil {
			log.Printf("Error updating expense category: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.category"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.category"}))
		ctx.Back()
	})
}
