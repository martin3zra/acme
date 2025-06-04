package app

import (
	"log"

	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) homeHandler(ctx *routing.Context) {
	log.Println("Home route accessed")
	ctx.Render("Home/Index", map[string]any{})
}
