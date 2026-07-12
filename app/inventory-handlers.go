package app

import (
	"context"
	"errors"
	"log"

	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
	inertia "github.com/romsar/gonertia/v2"
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
	transfers, err := s.findTransfers(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("transfersHandler: %v", err)
		ctx.Error(err)
		return
	}
	if transfers == nil {
		transfers = []*transferRow{}
	}

	ctx.Render("Inventories/Transfers/Index", map[string]any{
		"translations": trans("global", "movements", "transfers"),
		"transfers":    transfers,
	})
}

// createTransferHandler renders the transfer document form. Warehouses are loaded
// up front; products are searched on demand (partial reloads) like purchases.
func (s *Server) createTransferHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	term := ctx.Query("search")

	warehouses, err := s.findWarehouses(ctx.Request.Context())
	if err != nil {
		log.Printf("createTransferHandler: warehouses: %v", err)
		ctx.Error(err)
		return
	}

	ctx.Render("Inventories/Transfers/Create", map[string]any{
		"translations": trans("global", "movements", "transfers"),
		"warehouses":   warehouses,
		"item": inertia.Optional(func() (any, error) {
			return s.findTransferItem(ctx.Request.Context(), companyID, term)
		}),
		"items": inertia.Optional(func() (any, error) {
			return s.findTransferItems(ctx.Request.Context(), companyID, term)
		}),
	})
}

// showTransferHandler renders the transfer detail (header + lines + actions).
func (s *Server) showTransferHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	uuid := ctx.Param("id")

	header, err := s.findTransferByUUID(ctx.Request.Context(), companyID, uuid)
	if err != nil {
		if errors.Is(err, ErrTransferNotFound) {
			ctx.Redirect("/inventories/transfers")
			return
		}
		log.Printf("showTransferHandler: %v", err)
		ctx.Error(err)
		return
	}

	lines, err := s.findTransferLines(ctx.Request.Context(), companyID, header.ID)
	if err != nil {
		log.Printf("showTransferHandler: lines: %v", err)
		ctx.Error(err)
		return
	}
	if lines == nil {
		lines = []*transferLineRow{}
	}

	ctx.Render("Inventories/Transfers/Show", map[string]any{
		"translations": trans("global", "movements", "transfers"),
		"transfer":     header,
		"lines":        lines,
	})
}

// movementsHandler renders the full inventory movement ledger (every stock in/out
// across sales, purchases, adjustments and transfers).
func (s *Server) movementsHandler(ctx *routing.Context) {
	companyID := CurrentCompany(ctx.Request.Context()).ID
	movements, err := s.findMovements(ctx.Request.Context(), companyID)
	if err != nil {
		log.Printf("movementsHandler: %v", err)
		ctx.Error(err)
		return
	}
	if movements == nil {
		movements = []*inventoryMovementRow{}
	}

	ctx.Render("Inventories/Movements/Index", map[string]any{
		"translations": trans("global", "movements"),
		"movements":    movements,
	})
}

// transferTransitionHandler returns a handler that drives a single transfer
// status transition (dispatch / receive / cancel) by uuid.
func (s *Server) transferTransitionHandler(action func(context.Context, string) error) routing.HandlerFunc {
	return func(ctx *routing.Context) {
		uuid := ctx.Param("id")
		if uuid == "" {
			ctx.BackWith("status", s.trans("movements.errors.invalidTransition", nil))
			return
		}

		if err := action(ctx.Request.Context(), uuid); err != nil {
			switch {
			case errors.Is(err, ErrTransferNotFound), errors.Is(err, ErrInvalidTransition):
				ctx.Flash("error", s.trans("movements.errors.invalidTransition", nil))
				ctx.Back()
			case errors.Is(err, ErrInsufficientStock):
				ctx.Flash("error", s.trans("movements.errors.insufficientStock", nil))
				ctx.Back()
			default:
				log.Printf("transferTransitionHandler: %v", err)
				ctx.BackWithError(err)
			}
			return
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.transfer"}))
		ctx.Redirect("/inventories/transfers")
	}
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

	defaults, err := s.findCompanyDefaults(ctx.Request.Context())
	if err != nil {
		log.Printf("adjustmentsHandler: defaults: %v", err)
		ctx.Error(err)
		return
	}

	ctx.Render("Inventories/Adjustments/Index", map[string]any{
		"adjustments":        adjustments,
		"variants":           variants,
		"warehouses":         warehouses,
		"defaultWarehouseId": defaults.WarehouseID,
	})
}

func (s *Server) storeAdjustmentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreAdjustmentForm) {
		if err := s.storeAdjustment(ctx.Request.Context(), form); err != nil {
			log.Printf("storeAdjustmentHandler: %v", err)
			ctx.BackWithError(err)
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.adjustment"}))
		ctx.Redirect("/inventories/adjustments")
	})
}

func (s *Server) storeTransferHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreTransferForm) {
		if err := s.storeTransfer(ctx.Request.Context(), form); err != nil {
			// Surface known business errors on the relevant form field.
			switch {
			case errors.Is(err, ErrSameWarehouse):
				ctx.Errors("to_warehouse_id", s.trans("movements.errors.sameWarehouse", nil))
				ctx.Back()
			case errors.Is(err, ErrNoTransferLines):
				ctx.Errors("lines", s.trans("transfers.errors.noLines", nil))
				ctx.Back()
			case errors.Is(err, ErrInsufficientStock):
				ctx.Errors("qty", s.trans("movements.errors.insufficientStock", nil))
				ctx.Back()
			default:
				log.Printf("storeTransferHandler: %v", err)
				ctx.BackWithError(err)
			}
			return
		}
		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.transfer"}))
		ctx.Redirect("/inventories/transfers")
	})
}
