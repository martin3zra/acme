package app

import (
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) onboardingHandler(ctx *routing.Context) {
	ctx.Render("Onboarding/Index", map[string]any{
		"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("onboarding")),
		"status":       ctx.Query("status"),
	})
}
