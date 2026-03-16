package app

import "github.com/martin3zra/acme/pkg/routing"

func (s *Server) purchaseOrdersHandler(ctx *routing.Context) {
	props := map[string]any{
		"translations": trans("purchases"),
	}
	ctx.Render("Purchases/Orders/Index", props)
}

func (s *Server) purchaseReceiptsHandler(ctx *routing.Context) {
	props := map[string]any{
		"translations": trans("purchases"),
	}
	ctx.Render("Purchases/Receipts/Index", props)
}

func (s *Server) purchaseVendorBillsHandler(ctx *routing.Context) {
	props := map[string]any{
		"translations": trans("purchases"),
	}
	ctx.Render("Purchases/VendorBills/Index", props)
}
