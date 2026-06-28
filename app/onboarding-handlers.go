package app

import (
	"github.com/martin3zra/forge/routing"
)

func (s *Server) onboardingHandler(ctx *routing.Context) {
	ctx.Render("Onboarding/Index", map[string]any{
		"translations": trans("onboarding", "companies"),
		"status":       ctx.Query("status"),
	})
}
