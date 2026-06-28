package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func resolvePurchaseTransactionKind(ctx *routing.Context) PurchaseTransactionKind {
	kind := ctx.Query("kind")
	if kind != "" {
		return PurchaseTransactionKind(kind)
	}

	path := ctx.Request.URL.Path
	switch {
	case strings.HasPrefix(path, "/purchases/orders"):
		return PurchaseTransactionKinds.PurchaseOrder
	case strings.HasPrefix(path, "/purchases/receipts"):
		return PurchaseTransactionKinds.PurchaseReceipt
	case strings.HasPrefix(path, "/purchases/vendor-bills"):
		return PurchaseTransactionKinds.VendorBill
	default:
		return PurchaseTransactionKinds.VendorBill
	}
}

func purchasesPathForKind(kind PurchaseTransactionKind) string {
	switch kind {
	case PurchaseTransactionKinds.PurchaseOrder:
		return "/purchases/orders"
	case PurchaseTransactionKinds.PurchaseReceipt:
		return "/purchases/receipts"
	case PurchaseTransactionKinds.VendorBill:
		return "/purchases/vendor-bills"
	default:
		return "/purchases"
	}
}

func (s *Server) purchasesHandler(ctx *routing.Context) {
	kind := resolvePurchaseTransactionKind(ctx)
	if s.abortWhenPrerequisiteMissing(ctx, "purchase") {
		return
	}

	purchaseIDStr := ctx.Query("id")
	purchases, err := s.findPurchases(ctx.Request.Context(), kind)
	if err != nil {
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"translations": trans("purchases"),
		"purchases":    purchases,
		"kind":         kind,
	}

	if purchaseIDStr != "" {
		company := CurrentCompany(ctx.Request.Context())
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:purchase:%s", purchaseIDStr)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
			purchase, err := s.findPurchaseByUUID(ctx.Request.Context(), company.ID, purchaseIDStr)
			if err != nil {
				return nil, err
			}

			lines, err := s.findPurchaseLines(ctx.Request.Context(), purchase.CompanyID, purchase.ID)
			if err != nil {
				return nil, err
			}

			if purchase.Kind == PurchaseTransactionKinds.PurchaseOrder {
				purchase.LinkedReceipts, err = s.findLinkedReceiptsForOrder(ctx.Request.Context(), purchase.CompanyID, purchase.UUID)
				if err != nil {
					return nil, err
				}
				if err = s.enrichLinesWithRemainingQty(ctx.Request.Context(), purchase.CompanyID, purchase.UUID, lines); err != nil {
					return nil, err
				}
			}

			return map[string]any{
				"header": purchase,
				"lines":  lines,
			}, nil
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["purchase"] = data
		props["showPurchase"] = true
	}

	ctx.Render("Purchases/Index", props)
}

func (s *Server) showPurchaseHandler(ctx *routing.Context) {
	uuid := ctx.Param("id")
	if uuid == "" {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"status": "The given purchase ID is not valid.",
		})
		return
	}
	company := CurrentCompany(ctx.Request.Context())
	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:purchase:%s", uuid)
	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
		purchase, err := s.findPurchaseByUUID(ctx.Request.Context(), company.ID, uuid)
		if err != nil {
			return nil, err
		}

		lines, err := s.findPurchaseLines(ctx.Request.Context(), purchase.CompanyID, purchase.ID)
		if err != nil {
			return nil, err
		}

		if purchase.Kind == PurchaseTransactionKinds.PurchaseOrder {
			purchase.LinkedReceipts, err = s.findLinkedReceiptsForOrder(ctx.Request.Context(), purchase.CompanyID, purchase.UUID)
			if err != nil {
				return nil, err
			}
			if err = s.enrichLinesWithRemainingQty(ctx.Request.Context(), purchase.CompanyID, purchase.UUID, lines); err != nil {
				return nil, err
			}
		}

		return map[string]any{
			"header": purchase,
			"lines":  lines,
		}, nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"status": "An error retrieving the purchase.",
			"data":   err,
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}

func (s *Server) createPurchaseHandler(ctx *routing.Context) {
	if s.abortWhenPrerequisiteMissing(ctx, "purchase") {
		return
	}

	term := ctx.Query("search")
	kind := resolvePurchaseTransactionKind(ctx)

	var v *vendor
	var err error
	vendorID := ctx.Query("vendor_id")
	if strings.TrimSpace(vendorID) != "" {
		v, err = s.findVendorByUUID(ctx.Request.Context(), vendorID)
		if err != nil {
			ctx.Error(err)
			return
		}
	}

	props := map[string]any{
		"translations": trans("purchases"),
		"kind":         kind,
		"vendor":       v,
		"vendors": inertia.Optional(func() (any, error) {
			vendors, err := s.findVendorsBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return vendors, nil
		}),
		"item": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByReference(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return item, nil
		}),
		"items": inertia.Optional(func() (any, error) {
			items, err := s.findItemsByCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return items, nil
		}),
	}

	ctx.Render("Purchases/Create", props)
}

func (s *Server) editPurchaseHandler(ctx *routing.Context) {
	company := CurrentCompany(ctx.Request.Context())
	uuid := ctx.Param("id")
	purchase, err := s.findPurchaseByUUID(ctx.Request.Context(), company.ID, uuid)
	if err != nil {
		ctx.Error(err)
		return
	}

	lines, err := s.findPurchaseLines(ctx.Request.Context(), purchase.CompanyID, purchase.ID)
	if err != nil {
		ctx.Error(err)
		return
	}

	term := ctx.Query("search")
	props := map[string]any{
		"translations": trans("purchases"),
		"kind":         purchase.Kind,
		"purchase": map[string]any{
			"header": purchase,
			"lines":  lines,
		},
		"vendors": inertia.Optional(func() (any, error) {
			vendors, err := s.findVendorsBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return vendors, nil
		}),
		"item": inertia.Optional(func() (any, error) {
			item, err := s.findItemsByReference(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return item, nil
		}),
		"items": inertia.Optional(func() (any, error) {
			items, err := s.findItemsByCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return items, nil
		}),
	}

	ctx.Render("Purchases/Edit", props)
}

func (s *Server) storePurchaseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StorePurchaseForm) {
		uuid, err := s.storePurchase(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating purchase: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.purchase"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.purchase"}))
		ctx.Flash("redirectTo", purchasesPathForKind(form.Kind))
		ctx.Flash("purchase_uuid", uuid)
		ctx.Back()
	})
}

func (s *Server) updatePurchaseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *UpdatePurchaseForm) {
		uuid := ctx.Param("id")
		err := s.updatePurchase(ctx.Request.Context(), uuid, form)
		if err != nil {
			log.Printf("Error updating purchase: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotUpdated", i18n.Replacements{"subject": "@global.purchase"}))
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:purchase:%s", uuid)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.purchase"}))
		ctx.Redirect(purchasesPathForKind(form.Kind))
	})
}

func (s *Server) destroyPurchaseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		uuid := ctx.Param("id")
		err := s.destroyPurchase(ctx.Request.Context(), uuid)
		if err != nil {
			log.Printf("Error deleting purchase: %v", err)
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.purchase"}))
		ctx.Back()
	})
}

func (s *Server) confirmPurchaseHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		uuid := ctx.Param("id")
		err := s.confirmPurchase(ctx.Request.Context(), uuid)
		if err != nil {
			log.Printf("Error confirming purchase: %v", err)
			ctx.BackWith("status", err.Error())
			return
		}

		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:purchase:%s", uuid)
		if err = c.Delete(ctx.Request.Context(), key); err != nil {
			log.Printf("Error deleting cache: %v", err)
		}

		ctx.Flash("success", s.trans("global.wasUpdated", i18n.Replacements{"subject": "@global.purchase"}))
		ctx.Back()
	})
}

func (s *Server) purchaseOrdersHandler(ctx *routing.Context)      { s.purchasesHandler(ctx) }
func (s *Server) purchaseReceiptsHandler(ctx *routing.Context)    { s.purchasesHandler(ctx) }
func (s *Server) purchaseVendorBillsHandler(ctx *routing.Context) { s.purchasesHandler(ctx) }
