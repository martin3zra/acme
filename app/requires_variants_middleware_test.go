package app

import (
	"testing"

	"github.com/martin3zra/forge/routing"
)

// TestRequiresVariantsMiddleware: the attribute routes are gated on the company
// variants flag — blocked (next never runs) when off, passed through when on.
func TestRequiresVariantsMiddleware(t *testing.T) {
	s := newHandlerServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	called := false
	next := func(ctx *routing.Context) { called = true }
	guarded := s.RequiresVariants(next)

	// Flag defaults off → request blocked.
	ctx, _, _ := handlerCtx(t, s, f, "GET", "/attributes", nil)
	guarded(ctx)
	is.True(!called, "next must not run while variants flag is off")

	// Enable the flag → request passes through.
	is.NoErr(s.updateHandlesVariants(f.ctx, f.company.UUID, true))
	ctx, _, _ = handlerCtx(t, s, f, "GET", "/attributes", nil)
	guarded(ctx)
	is.True(called, "next must run once variants flag is on")
}
