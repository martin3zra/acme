package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/forge/routing"
)

func (s *Server) storeUnitHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreUnitForm) {

		if err := s.storeUnit(ctx.Request.Context(), form); err != nil {
			log.Printf("Error creating unit: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.unit"}))
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.unit"}))
		ctx.Back()
	})
}

func (s *Server) updateUnitHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreUnitForm) {

		id := ctx.Int("id")
		if err := s.updateUnit(ctx.Request.Context(), id, form); err != nil {
			log.Printf("Error updating unit: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.unit"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.unit"}))
		ctx.Back()
	})
}
