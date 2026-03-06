package main

import (
	"fmt"

	"acme/pkg/i18n"
	"acme/pkg/routing"
)

// stockLevelsHandler returns list of stock levels
func (s *Server) stockLevelsHandler() routing.HandlerFunc {
	return func(ctx *routing.Context) {
		filters := map[string]interface{}{}

		// Optional filter by warehouse
		if warehouseID := ctx.Query("warehouse_id"); warehouseID != "" {
			filters["warehouse_id"] = warehouseID
		}

		// Optional filter by variant
		if variantID := ctx.Query("variant_id"); variantID != "" {
			filters["variant_id"] = variantID
		}

		stocks, err := s.findStockLevels(ctx.Request.Context(), filters)
		if err != nil {
			ctx.Error(err)
			return
		}

		warehouses, err := s.findWarehouses(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Render("Stock/Index", map[string]any{
			"stocks":     stocks,
			"warehouses": warehouses,
		})
	}
}

// stockByItemHandler returns stock levels for a specific item
func (s *Server) stockByItemHandler() routing.HandlerFunc {
	return func(ctx *routing.Context) {
		itemID := ctx.Int("id")

		stocks, err := s.findStockLevels(ctx.Request.Context(), map[string]interface{}{
			"item_id": itemID,
		})
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Render("Stock/ByItem", map[string]any{
			"stocks":  stocks,
			"item_id": itemID,
		})
	}
}

// stockByWarehouseHandler returns stock levels for a specific warehouse
func (s *Server) stockByWarehouseHandler() routing.HandlerFunc {
	return func(ctx *routing.Context) {
		warehouseID := ctx.Int("id")

		warehouse, err := s.findWarehouseByID(ctx.Request.Context(), warehouseID)
		if err != nil {
			ctx.Error(err)
			return
		}

		stocks, err := s.getStockForWarehouse(ctx.Request.Context(), warehouseID)
		if err != nil {
			ctx.Error(err)
			return
		}

		ctx.Render("Stock/ByWarehouse", map[string]any{
			"stocks":    stocks,
			"warehouse": warehouse,
		})
	}
}

// adjustStockHandler handles stock quantity adjustments
func (s *Server) adjustStockHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StockAdjustmentForm) {
		err := s.adjustStock(ctx.Request.Context(), form.WarehouseID, form.VariantID, form.Quantity)
		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.stock"}))
		ctx.Redirect("/stock-levels")
	})
}

// initializeStockForVariantHandler initializes stock for a new variant across all warehouses
func (s *Server) initializeStockForVariantHandler() routing.HandlerFunc {
	return func(ctx *routing.Context) {
		variantID := ctx.Int("variant_id")

		warehouses, err := s.findWarehouses(ctx.Request.Context())
		if err != nil {
			ctx.Error(err)
			return
		}

		companyID := CurrentCompany(ctx).ID

		// Create stock records for all warehouses
		for _, warehouse := range warehouses {
			err := s.initializeStock(nil, companyID, warehouse.ID, variantID, 0)
			if err != nil {
				ctx.Error(err)
				return
			}
		}

		ctx.Flash("success", fmt.Sprintf("Stock initialized for variant across %d warehouses", len(warehouses)))
		ctx.Redirect(fmt.Sprintf("/items/%d/stock", variantID))
	}
}

// updateStockReorderLevelsHandler updates reorder settings for stock
func (s *Server) updateStockReorderLevelsHandler() routing.HandlerFunc {
	type updateStockForm struct {
		WarehouseID     int `json:"warehouse_id"`
		VariantID       int `json:"variant_id"`
		ReorderLevel    int `json:"reorder_level"`
		ReorderQuantity int `json:"reorder_quantity"`
	}

	return routing.WithRequest(func(ctx *routing.Context, form *updateStockForm) {
		_, err := s.db.ExecContext(
			ctx.Request.Context(),
			`UPDATE stock_levels 
			 SET reorder_level = $1, reorder_quantity = $2, updated_at = NOW()
			 WHERE company_id = $3 AND warehouse_id = $4 AND variant_id = $5`,
			form.ReorderLevel, form.ReorderQuantity,
			CurrentCompany(ctx.Request.Context()).ID, form.WarehouseID, form.VariantID,
		)

		if err != nil {
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", "Reorder levels updated")
		ctx.Redirect("/stock-levels")
	})
}
