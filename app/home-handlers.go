package app

import (
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) homeHandler(ctx *routing.Context) {
	ctx.Render("Home/Index", map[string]any{})
}
