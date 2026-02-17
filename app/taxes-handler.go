package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) storeTaxes() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTaxForm) {

		if err := s.storeTax(ctx.Request.Context(), form); err != nil {
			log.Printf("Error creating tax: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.tax"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.tax"}))
		ctx.Back()
	})
}
func (s *Server) updateTaxes() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTaxForm) {

		uuid := ctx.Param("id")
		if err := s.updateTax(ctx.Request.Context(), uuid, form); err != nil {
			log.Printf("Error updating tax: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.tax"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.tax"}))
		ctx.Back()
	})
}
