package app

import "github.com/martin3zra/acme/pkg/routing"

func (s *Server) warehousesHandler(ctx *routing.Context) {
	props := map[string]any{}
	ctx.Render("Inventories/Warehouses/Index", props)
}

func (s *Server) stocksHandler(ctx *routing.Context) {
	props := map[string]any{}
	ctx.Render("Inventories/Stocks/Index", props)
}

func (s *Server) transfersHandler(ctx *routing.Context) {
	props := map[string]any{}
	ctx.Render("Inventories/Transfers/Index", props)
}

func (s *Server) adjustmentsHandler(ctx *routing.Context) {
	props := map[string]any{}
	ctx.Render("Inventories/Adjustments/Index", props)
}
