package app

import (
	"net/http"

	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) bootRoutes() {

	s.route.WithMiddleware(s.SharedProps)

	s.route.
		WithMiddleware(RedirectIfAuthenticated).
		Group(func(route *routing.Router) {
			route.GET("/login", s.login)
			route.POST("/login", s.authHandler)

			route.GET("/verify-account", func(ctx *routing.Context) {
				props := map[string]any{
					"translations": mergeTranslations(ctx.Request.Context(), loadTranslations("verify")),
				}
				ctx.Render("Verify/Index", props)
			})
			route.GET("/verify-account/:uuid/:hash", s.verifyAccountHandler)
			route.GET("/verify-email/:uuid/:hash", s.verifyEmailHandler).WithoutMiddleware(RedirectIfAuthenticated)
			route.GET("/verify-email", s.verifyEmailPromptHandler).WithoutMiddleware(RedirectIfAuthenticated)
			route.POST("/email/verification-notification", s.sendVerificationEmail).WithoutMiddleware(RedirectIfAuthenticated)
		})

	s.route.
		WithMiddleware(AuthenticatedMiddleware, Verified, EnforceVerifiedUserAccess).
		WithoutGroupMiddleware(RedirectIfAuthenticated).
		Group(func(route *routing.Router) {
			route.GET("/onboarding", s.onboardingHandler)

			route.GET("/home", s.homeHandler)
			route.POST("/logout", s.logoutHandler)

			route.POST("/companies", s.storeCompanyHandler)

			route.GET("/customers", s.customersHandler)
			route.POST("/customers", s.storeCustomerHandler)
			route.PUT("/customers/:id", s.updateCustomerHandler)
			route.PUT("/customers/:id/change-status", s.changeStatusCustomerHandler)
			route.DELETE("/customers/:id", s.deleteCustomerHandler)

			route.GET("/items", s.itemsHandler)
			route.POST("/items", s.storeItemHandler)
			route.PUT("/items/:id", s.updateItemHandler)
			route.PUT("/items/:id/change-status", s.changeStatusItemHandler)
			route.DELETE("/items/:id", s.deleteItemHandler)

			route.GET("/invoices", s.invoicesHandler)
			route.POST("/invoices", s.storeInvoiceHandler)
			route.GET("/invoices/create", s.createInvoiceHandler)
			route.GET("/invoices/:id/edit", s.editInvoiceHandler)
			route.PUT("/invoices/:id/void", s.voidInvoiceHandler)
			route.PUT("/invoices/:id", s.updateInvoiceHandler)

			route.GET("/payments", s.paymentsHandler)
			route.POST("/payments", s.storePaymentHandler)
			route.GET("/payments/create", s.createPaymentHandler)
			route.GET("/payments/:id/edit", s.editPaymentHandler)
			route.PUT("/payments/:id/void", s.voidPaymentHandler)
			route.PUT("/payments/:id", s.updatePaymentHandler)

			route.POST("/password", s.createPasswordHandler)

			route.GroupPrefix("/settings/:account", func(route *routing.Router) {
				route.GET("/profile", s.accountProfileHandler)
				route.PUT("/profile", s.updateAccountProfileHandler())

				route.GET("/companies", s.companyHandler)
				route.GET("/users", func(ctx *routing.Context) {
					ctx.Render("Settings/Users/Index", map[string]any{})
				})
				// route.GET("/preferences", func(ctx *routing.Context) {
				// 	ctx.Render("Settings/Preferences", map[string]any{})
				// })
				// route.GET("/taxes", func(ctx *routing.Context) {
				// 	ctx.Render("Settings/Taxes/Index", map[string]any{})
				// })
				route.GET("/settings/profile", func(ctx *routing.Context) {
					ctx.Render("Settings/Profile", map[string]any{})
				})
			})
		})

	uiAssets := foundation.GetBuildAssets(s.assets, "public/build")
	s.route.FileServer("/build/", http.FS(uiAssets))
}
