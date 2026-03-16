package app

import "github.com/martin3zra/acme/pkg/routing"

func (s *Server) vendorsHandler(ctx *routing.Context) {
	props := map[string]any{}
	ctx.Render("Vendors/Index", props)
}
