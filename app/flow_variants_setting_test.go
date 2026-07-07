package app

import "testing"

// TestFlowHandlesVariantsSetting: the company variants flag defaults off, and the
// updater flips it (read back via both the current-company getter and the
// by-uuid getter the settings screen uses).
func TestFlowHandlesVariantsSetting(t *testing.T) {
	s := newTestServer(t)
	is := newIs(t)
	f := mkAccountCompany(t, s)

	enabled, err := f.s.companyHandlesVariants(f.ctx)
	is.NoErr(err)
	is.True(!enabled, "variants flag should default off")

	is.NoErr(f.s.updateHandlesVariants(f.ctx, f.company.UUID, true))

	enabled, err = f.s.companyHandlesVariants(f.ctx)
	is.NoErr(err)
	is.True(enabled, "flag on after enabling")

	got, err := f.s.findHandlesVariants(f.ctx, f.company.UUID)
	is.NoErr(err)
	is.True(got, "by-uuid getter reflects the flag")

	// and back off
	is.NoErr(f.s.updateHandlesVariants(f.ctx, f.company.UUID, false))
	enabled, err = f.s.companyHandlesVariants(f.ctx)
	is.NoErr(err)
	is.True(!enabled, "flag off after disabling")
}
