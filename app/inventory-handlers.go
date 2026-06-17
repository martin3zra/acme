package app

import (
	"log"
	"net/http"

	"github.com/martin3zra/forge/routing"
)

func (s *Server) stocksHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	stocks, err := s.findStocks(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("stocksHandler: %v", err)
		ctx.Error(err)
		return
	}
	if stocks == nil {
		stocks = []*stockBalance{}
	}
	ctx.Render("Inventories/Stocks/Index", map[string]any{
		"stocks": stocks,
	})
}

func (s *Server) transfersHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	movements, err := s.findMovements(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("transfersHandler: %v", err)
		ctx.Error(err)
		return
	}
	if movements == nil {
		movements = []*inventoryMovementRow{}
	}
	ctx.Render("Inventories/Transfers/Index", map[string]any{
		"movements": movements,
	})
}

func (s *Server) adjustmentsHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	adjustments, err := s.findAdjustments(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("adjustmentsHandler: %v", err)
		ctx.Error(err)
		return
	}
	if adjustments == nil {
		adjustments = []*adjustmentRow{}
	}

	variants, err := s.findTrackableVariants(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("adjustmentsHandler: variants: %v", err)
		ctx.Error(err)
		return
	}

	warehouses, err := s.findWarehouses(ctx.Request.Context())
	if err != nil {
		log.Printf("adjustmentsHandler: warehouses: %v", err)
		ctx.Error(err)
		return
	}

	ctx.Render("Inventories/Adjustments/Index", map[string]any{
		"adjustments": adjustments,
		"variants":    variants,
		"warehouses":  warehouses,
	})
}

func (s *Server) storeAdjustmentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAdjustmentForm) {
		if err := s.storeAdjustment(ctx.Request.Context(), form); err != nil {
			log.Printf("storeAdjustmentHandler: %v", err)
			ctx.BackWithError(err)
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", nil))
		ctx.JSON(http.StatusCreated, map[string]any{"message": "Adjustment recorded"})
	})
}
