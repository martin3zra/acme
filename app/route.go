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
					"translations": trans("verify"),
				}
				ctx.Render("Verify/Index", props)
			})
			route.GET("/verify-account/:uuid/:hash", s.verifyAccountHandler).Middleware(Signed)
			route.GET("/verify-email/:uuid/:hash", s.verifyEmailHandler).WithoutMiddleware(RedirectIfAuthenticated).Middleware(Signed)
			route.GET("/verify-email", s.verifyEmailPromptHandler).WithoutMiddleware(RedirectIfAuthenticated)
			route.POST("/email/verification-notification", s.sendVerificationEmail).WithoutMiddleware(RedirectIfAuthenticated)
		})

	s.route.
		WithMiddleware(AuthenticatedMiddleware, Verified, EnforceVerifiedUserAccess).
		WithoutGroupMiddleware(RedirectIfAuthenticated).
		Group(func(route *routing.Router) {
			route.GET("/onboarding", s.onboardingHandler)

			route.
				WithMiddleware(RestrictedAccess).
				Group(func(route *routing.Router) {
					route.GET("/home", s.homeHandler)

					route.POST("/companies", s.storeCompanyHandler)

					route.GET("/customers", s.customersHandler).Can("viewAny:customer")
					route.POST("/customers", s.storeCustomerHandler())
					route.PUT("/customers/:id", s.updateCustomerHandler())
					route.PUT("/customers/:id/change-status", s.changeStatusCustomerHandler())
					route.DELETE("/customers/:id", s.deleteCustomerHandler())

					route.GET("/items", s.itemsHandler).Can("viewAny:item")
					route.POST("/items", s.storeItemHandler())
					route.PUT("/items/:id", s.updateItemHandler())
					route.PUT("/items/:id/change-status", s.changeStatusItemHandler())
					route.DELETE("/items/:id", s.deleteItemHandler())

					route.GET("/invoices", s.invoicesHandler).Can("viewAny:invoice")
					route.POST("/invoices", s.storeInvoiceHandler())
					route.GET("/invoices/create", s.createInvoiceHandler).Can("create:invoice")
					route.GET("/invoices/:id/edit", s.editInvoiceHandler).Can("update:invoice")
					route.PUT("/invoices/:id/void", s.voidInvoiceHandler())
					route.PUT("/invoices/:id", s.updateInvoiceHandler())

					route.GET("/payments", s.paymentsHandler).Can("viewAny:payment")
					route.POST("/payments", s.storePaymentHandler())
					route.GET("/payments/create", s.createPaymentHandler).Can("create:payment")
					route.GET("/payments/:id/edit", s.editPaymentHandler).Can("update:payment")
					route.PUT("/payments/:id/void", s.voidPaymentHandler())
					route.PUT("/payments/:id", s.updatePaymentHandler())

					route.POST("/password", s.createPasswordHandler())

					route.GroupPrefix("/settings/:account", func(route *routing.Router) {
						route.GET("/profile", s.accountProfileHandler)
						route.PUT("/profile", s.updateAccountProfileHandler())

						route.GET("/companies", s.companyHandler).Can("viewAny:company")
						route.PUT("/companies/:id/sequences", s.companyUpdateSequences())
						route.POST("/users", s.storeUserHandler())
						route.PUT("/users/:id", s.updateUserHandler())
					})
				})

			route.
				WithoutGroupMiddleware(RestrictedAccess).
				Group(func(route *routing.Router) {
					route.POST("/logout", s.logoutHandler)

					route.GET("/awaiting-association", func(ctx *routing.Context) {
						ctx.Render("Restricted/AwaitingAssociation/Index", map[string]any{})
					})
				})
		})

	s.route.
		GET("/", func(ctx *routing.Context) {
			props := map[string]any{}
			ctx.Render("Welcome/Index", props)
		}).
		WithoutMiddleware(RedirectIfAuthenticated, Verified, EnforceVerifiedUserAccess)

	uiAssets := foundation.GetBuildAssets(s.assets, "public/build")
	s.route.FileServer("/build/", http.FS(uiAssets))
}
