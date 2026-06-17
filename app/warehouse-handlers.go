package app

import (
	"log"

	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
)

func (s *Server) warehousesHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "inventory") {
		return
	}

	warehouses, err := s.findWarehouses(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"openState":    ctx.Query("mode") == "creating",
		"translations": trans("warehouses"),
		"warehouses":   warehouses,
	}

	ctx.Render("Inventories/Warehouses/Index", props)
}

func (s *Server) storeWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreWarehouseForm) {
		err := s.storeWarehouse(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating warehouse: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.warehouse"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/inventories/warehouses")
	})
}

func (s *Server) updateWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdateWarehouseForm) {
		err := s.updateWarehouse(ctx.Request.Context(), ctx.Int("id"), form)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/inventories/warehouses")
	})
}

func (s *Server) deleteWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		err := s.deleteWarehouse(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("current_password", s.trans("global.wasNotDeleted", i18n.Replacements{"subject": "@global.warehouse"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/inventories/warehouses")
	})
}

func (s *Server) changeStatusWarehouseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		w, err := s.findWarehouseByID(ctx.Request.Context(), ctx.Int("id"))
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
			return
		}

		err = s.toggleWarehouseStatus(ctx.Request.Context(), w)
		if err != nil {
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.warehouse"}))
		ctx.Redirect("/inventories/warehouses")
	})
}
