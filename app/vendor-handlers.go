package app

import "github.com/martin3zra/acme/pkg/routing"

func (s *Server) vendorsHandler(ctx *routing.Context) {
	vendorType := VendorType(ctx.Query("vendorType"))
	if err := vendorType.Validate(); err != nil {
		vendorType = "all"
	}
	vendors, err := s.findVendors(ctx.Request.Context(), vendorType)
	if err != nil {
		ctx.Error(err)
		return
	}
	props := map[string]any{
		"openState":               ctx.Query("mode") == "creating",
		"translations":            trans("vendors"),
		"vendors":                 vendors,
		"currentVendorTypeFilter": vendorType,
	}
	ctx.Render("Vendors/Index", props)
}
