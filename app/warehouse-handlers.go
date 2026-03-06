package app

import (
	"fmt"

	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/validator/locale"
)

// warehousesHandler returns the list of warehouses
func (s *Server) warehousesHandler(ctx *routing.Context) {
	warehouses, err := s.findWarehouses(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	ctx.Render("Warehouses/Index", map[string]any{
		"warehouses": warehouses,
	})
}

// storeWarehouseHandler creates a new warehouse
func (s *Server) storeWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreWarehouseForm) {
		err := s.storeWarehouse(ctx.Request.Context(), form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/warehouses")
	})
}

// updateWarehouseHandler updates an existing warehouse
func (s *Server) updateWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateWarehouseForm) {
		form.ID = ctx.Int("id")

		err := s.updateWarehouse(ctx.Request.Context(), form.ID, form)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/warehouses")
	})
}

// deleteWarehouseHandler soft-deletes a warehouse
func (s *Server) deleteWarehouseHandler() routing.HandlerFunc {
	type form struct {
		ConfirmsPasswords
	}

	return routing.WithRequest(func(ctx *routing.Context, frm *form) {
		warehouseID := ctx.Int("id")

		err := s.deleteWarehouse(ctx.Request.Context(), warehouseID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", fmt.Sprintf(locale.SpanishMessages()["messages.deleted"].(string), locale.SpanishMessages()["global.warehouse"]))
		ctx.Redirect("/warehouses")
	})
}

// changeStatusWarehouseHandler toggles warehouse status
func (s *Server) changeStatusWarehouseHandler() routing.HandlerFunc {
	return func(ctx *routing.Context) {
		warehouseID := ctx.Int("id")

		err := s.toggleWarehouseStatus(ctx.Request.Context(), warehouseID)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/warehouses")
	}
}
