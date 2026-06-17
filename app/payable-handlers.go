package app

import (
	"fmt"
	"log"
	"net/http"

	"github.com/martin3zra/forge/cache"
	"github.com/martin3zra/forge/i18n"
	"github.com/martin3zra/forge/routing"
	inertia "github.com/romsar/gonertia/v2"
)

func (s *Server) payablesHandler(ctx *routing.Context) {
	payables, err := s.findPayables(ctx.Request.Context())
	if err != nil {
		ctx.Error(err)
		return
	}

	props := map[string]any{
		"translations": trans("payables"),
		"payables":     payables,
	}

	uuid := ctx.Query("id")
	if uuid != "" {
		c := cache.NewPgCache(s.db)
		key := fmt.Sprintf("preview:vendor_payment:%s", uuid)
		data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
			payment, err := s.findVendorPaymentByUUID(ctx.Request.Context(), uuid)
			if err != nil {
				return nil, err
			}
			lines, err := s.findVendorPaymentLines(ctx.Request.Context(), payment.ID)
			if err != nil {
				return nil, err
			}
			return map[string]any{
				"header": payment,
				"lines":  lines,
			}, nil
		})
		if err != nil {
			ctx.Error(err)
			return
		}
		props["vendorPayment"] = data
		props["showVendorPayment"] = true
	}

	ctx.Render("Payables/Index", props)
}

func (s *Server) createVendorPaymentHandler(ctx *routing.Context) {
	term := ctx.Query("search")
	vendorUUID := ctx.Query("vendor_id")

	props := map[string]any{
		"translations": trans("payables"),
		"vendors": inertia.Optional(func() (any, error) {
			vendors, err := s.findVendorsBySearchCriteria(ctx.Request.Context(), term)
			if err != nil {
				return nil, err
			}
			return vendors, nil
		}),
	}

	if vendorUUID != "" {
		v, err := s.findVendorByUUID(ctx.Request.Context(), vendorUUID)
		if err != nil {
			ctx.Error(err)
			return
		}
		props["vendor"] = v

		payables, err := s.findVendorPayables(ctx.Request.Context(), vendorUUID)
		if err != nil {
			ctx.Error(err)
			return
		}
		props["payables"] = payables
	}

	ctx.Render("Payables/Create", props)
}

func (s *Server) storeVendorPaymentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *StoreVendorPaymentForm) {
		err := s.storeVendorPayment(ctx.Request.Context(), form)
		if err != nil {
			log.Printf("Error creating vendor payment: %v", err)
			ctx.BackWith("status", s.trans("global.wasNotCreated", i18n.Replacements{"subject": "@global.vendorPayment"}))
			return
		}

		ctx.Flash("success", s.trans("global.wasCreated", i18n.Replacements{"subject": "@global.vendorPayment"}))
		ctx.Redirect("/payables")
	})
}

func (s *Server) voidVendorPaymentHandler() routing.HandlerFunc {
	return routing.WithRequest(func(ctx *routing.Context, form *ConfirmsPasswords) {
		uuid := ctx.Param("id")
		err := s.voidVendorPayment(ctx.Request.Context(), uuid)
		if err != nil {
			log.Printf("Error voiding vendor payment: %v", err)
			ctx.BackWithError(err)
			return
		}

		ctx.Flash("success", s.trans("global.wasDeleted", i18n.Replacements{"subject": "@global.vendorPayment"}))
		ctx.Back()
	})
}

func (s *Server) showVendorPaymentHandler(ctx *routing.Context) {
	uuid := ctx.Param("id")
	if uuid == "" {
		ctx.JSON(http.StatusBadRequest, map[string]any{
			"status": "The given vendor payment ID is not valid.",
		})
		return
	}

	c := cache.NewPgCache(s.db)
	key := fmt.Sprintf("preview:vendor_payment:%s", uuid)
	data, err := cache.Remember(ctx.Request.Context(), c, key, func() (map[string]any, error) {
		payment, err := s.findVendorPaymentByUUID(ctx.Request.Context(), uuid)
		if err != nil {
			return nil, err
		}
		lines, err := s.findVendorPaymentLines(ctx.Request.Context(), payment.ID)
		if err != nil {
			return nil, err
		}
		return map[string]any{
			"header": payment,
			"lines":  lines,
		}, nil
	})
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, map[string]any{
			"status": "An error retrieving the vendor payment.",
		})
		return
	}

	ctx.JSON(http.StatusOK, data)
}
