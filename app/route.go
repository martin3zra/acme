package app

import (
	"net/http"

	"github.com/martin3zra/forge/foundation"
	"github.com/martin3zra/forge/routing"
)

func (s *Server) bootRoutes() {

	s.route.WithMiddleware(s.RememberMe, s.SharedProps)

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
			route.GET("/onboarding", s.onboardingHandler).WithoutMiddleware(RestrictedAccess)
			route.POST("/companies", s.storeCompanyHandler).WithoutMiddleware(RestrictedAccess)

			route.
				WithMiddleware(RestrictedAccess, AutoResourcePrerequisiteMiddleware).
				Group(func(route *routing.Router) {
					route.GET("/home", s.homeHandler)

					route.GET("/customers", s.customersHandler).Can("viewAny:customer")
					route.POST("/customers", s.storeCustomerHandler())
					route.PUT("/customers/:id", s.updateCustomerHandler())
					route.PUT("/customers/:id/change-status", s.changeStatusCustomerHandler())
					route.DELETE("/customers/:id", s.deleteCustomerHandler())

					route.GET("/vendors", s.vendorsHandler).Can("viewAny:vendor")
					route.POST("/vendors", s.storeVendorHandler())
					route.PUT("/vendors/:id", s.updateVendorHandler())
					route.PUT("/vendors/:id/change-status", s.changeStatusVendorHandler())
					route.DELETE("/vendors/:id", s.deleteVendorHandler())

					route.GET("/items", s.itemsHandler).Can("viewAny:item")
					route.POST("/items", s.storeItemHandler())
					route.POST("/items/variants", s.storeItemWithVariantsHandler()).Can("create:item")
					route.PUT("/items/:id", s.updateItemHandler())
					route.PUT("/items/:id/change-status", s.changeStatusItemHandler())
					route.DELETE("/items/:id", s.deleteItemHandler())

					route.
						WithMiddleware(s.RequiresVariants).
						Group(func(route *routing.Router) {
							route.GET("/attributes", s.attributesHandler).Can("viewAny:attribute")
							route.POST("/attributes", s.storeAttributeHandler()).Can("create:attribute")
							route.PUT("/attributes/:id", s.updateAttributeHandler()).Can("update:attribute")
							route.DELETE("/attributes/:id", s.deleteAttributeHandler()).Can("delete:attribute")
							route.GET("/attributes/:id/values", s.attributeValuesHandler).Can("viewAny:attribute")
							route.POST("/attributes/:id/values", s.storeAttributeValueHandler()).Can("create:attribute")
							route.PUT("/attribute-values/:uuid", s.updateAttributeValueHandler()).Can("update:attribute")
							route.DELETE("/attribute-values/:uuid", s.deleteAttributeValueHandler()).Can("delete:attribute")
						})

					route.GET("/invoices", s.invoicesHandler).Can("viewAny:invoice")
					route.POST("/invoices", s.storeInvoiceHandler())
					route.GET("/invoices/create", s.createInvoiceHandler).Can("create:invoice")
					route.GET("/invoices/:id/edit", s.editInvoiceHandler).Can("update:invoice")
					route.PUT("/invoices/:id/void", s.voidInvoiceHandler())
					route.POST("/invoices/:id/recurrence", s.markInvoiceAsRecurrentHandler())
					route.GET("/invoices/:id/print/:hash", s.printInvoiceHandler).Middleware(Signed)
					route.PUT("/invoices/:id", s.updateInvoiceHandler())
					route.GET("/invoices/:id", s.showInvoiceHandler)

					route.GET("/estimates", s.invoicesHandler).Can("viewAny:estimate")
					route.GET("/estimates/create", s.createInvoiceHandler).Can("create:estimate")
					route.POST("/estimates", s.storeInvoiceHandler())
					route.GET("/estimates/:id/edit", s.editInvoiceHandler).Can("update:estimate")
					route.PUT("/estimates/:id/void", s.voidInvoiceHandler())
					route.GET("/estimates/:id/print/:hash", s.printInvoiceHandler).Middleware(Signed)
					route.PUT("/estimates/:id", s.updateInvoiceHandler())

					route.GET("/orders", s.invoicesHandler).Can("viewAny:order")
					route.GET("/orders/create", s.createInvoiceHandler).Can("create:order")
					route.POST("/orders", s.storeInvoiceHandler())
					route.GET("/orders/:id/edit", s.editInvoiceHandler).Can("update:order")
					route.PUT("/orders/:id/void", s.voidInvoiceHandler())
					route.GET("/orders/:id/print/:hash", s.printInvoiceHandler).Middleware(Signed)
					route.PUT("/orders/:id", s.updateInvoiceHandler())

					route.GET("/purchases", s.purchasesHandler).Can("viewAny:purchase")
					route.GET("/purchases/orders", s.purchasesHandler).Can("viewAny:purchase")
					route.GET("/purchases/receipts", s.purchasesHandler).Can("viewAny:purchase")
					route.GET("/purchases/vendor-bills", s.purchasesHandler).Can("viewAny:purchase")

					route.POST("/purchases", s.storePurchaseHandler())
					route.GET("/purchases/create", s.createPurchaseHandler).Can("create:purchase")
					route.GET("/purchases/orders/create", s.createPurchaseHandler).Can("create:purchase")
					route.GET("/purchases/receipts/create", s.createPurchaseHandler).Can("create:purchase")
					route.GET("/purchases/vendor-bills/create", s.createPurchaseHandler).Can("create:purchase")

					route.GET("/purchases/:id/edit", s.editPurchaseHandler).Can("update:purchase")
					route.PUT("/purchases/:id/confirm", s.confirmPurchaseHandler()).Can("confirm:purchase")
					route.PUT("/purchases/:id", s.updatePurchaseHandler())
					route.DELETE("/purchases/:id", s.destroyPurchaseHandler())
					route.GET("/purchases/:id", s.showPurchaseHandler)

					route.GET("/inventories/warehouses", s.warehousesHandler).Can("viewAny:inventory")
					route.POST("/inventories/warehouses", s.storeWarehouseHandler())
					route.PUT("/inventories/warehouses/:id", s.updateWarehouseHandler())
					route.PUT("/inventories/warehouses/:id/change-status", s.changeStatusWarehouseHandler())
					route.DELETE("/inventories/warehouses/:id", s.deleteWarehouseHandler())

					route.GET("/inventories/stocks", s.stocksHandler).Can("viewAny:inventory")
					route.GET("/inventories/movements", s.movementsHandler).Can("viewAny:inventory")
					route.GET("/inventories/transfers", s.transfersHandler).Can("viewAny:inventory")
					route.GET("/inventories/transfers/create", s.createTransferHandler).Can("create:transfer")
					route.POST("/inventories/transfers", s.storeTransferHandler()).Can("create:transfer")
					route.PUT("/inventories/transfers/:id/dispatch", s.transferTransitionHandler(s.dispatchTransfer)).Can("create:transfer")
					route.PUT("/inventories/transfers/:id/receive", s.transferTransitionHandler(s.receiveTransfer)).Can("create:transfer")
					route.PUT("/inventories/transfers/:id/cancel", s.transferTransitionHandler(s.cancelTransfer)).Can("create:transfer")
					route.GET("/inventories/transfers/:id", s.showTransferHandler).Can("viewAny:inventory")
					route.GET("/inventories/adjustments", s.adjustmentsHandler).Can("viewAny:inventory")
					route.POST("/inventories/adjustments", s.storeAdjustmentHandler()).Can("create:adjustment")

					route.GET("/payments", s.paymentsHandler).Can("viewAny:payment")
					route.POST("/payments", s.storePaymentHandler())
					route.GET("/payments/create", s.createPaymentHandler).Can("create:payment")
					route.GET("/payments/:id/edit", s.editPaymentHandler).Can("update:payment")
					route.PUT("/payments/:id/void", s.voidPaymentHandler())
					route.GET("/payments/:id/print/:hash", s.printPaymentHandler).Middleware(Signed)
					route.PUT("/payments/:id", s.updatePaymentHandler())

					route.GET("/payables", s.payablesHandler).Can("viewAny:payable")
					route.GET("/payables/create", s.createVendorPaymentHandler).Can("create:payable")
					route.POST("/payables", s.storeVendorPaymentHandler())
					route.PUT("/payables/:id/void", s.voidVendorPaymentHandler())
					route.GET("/payables/:id", s.showVendorPaymentHandler)

					route.GET("/expenses", s.expensesHandler).Can("viewAny:expense")
					route.POST("/expenses", s.storeExpenseHandler())
					route.DELETE("/expenses/:id", s.deleteExpenseHandler())
					route.PUT("/expenses/:id", s.updateExpenseHandler())

					route.GET("/reports/sales", s.reportSalesHandler).Can("viewAny:reports")
					route.POST("/reports/sales", s.generateSalesReportHandler())

					route.GET("/reports/profit-lost", s.reportProfitLostHandler).Can("viewAny:reports")
					route.POST("/reports/profit-lost", s.generateProfitLostReportHandler())

					route.GET("/reports/expenses", s.reportExpensesHandler).Can("viewAny:reports")
					route.POST("/reports/expenses", s.generateExpensesReportHandler())

					route.GET("/reports/taxes", s.reportTaxesHandler).Can("viewAny:reports")
					route.POST("/reports/taxes", s.generateTaxesReportHandler())

					route.POST("/taxes", s.storeTaxes())
					route.PUT("/taxes/:id", s.updateTaxes())

					route.POST("/expense-categories", s.storeExpenseCategoryHandler())
					route.PUT("/expense-categories/:id", s.updateExpenseCategoryHandler())

					route.POST("/units", s.storeUnitHandler())
					route.PUT("/units/:id", s.updateUnitHandler())

					route.POST("/uploads/init", s.startUploadChunkHandler())
					route.POST("/uploads/chunks", s.uploadChunkHandler())
					route.POST("/uploads/complete", s.completeUploadChunkHandler())

					route.POST("/imports", s.startImportHandler())

					route.GroupPrefix("/settings/:account", func(route *routing.Router) {
						route.GET("/profile", s.accountProfileHandler)
						route.PUT("/profile", s.updateAccountProfileHandler())

						route.GET("/companies", s.companyHandler).Can("viewAny:company")
						route.PUT("/companies/:id/sequences", s.companyUpdateSequences())
						route.PUT("/companies/:id/redirect-preferences", s.companyUpdateRedirectPreferences())
						route.PUT("/companies/:id/handles-variants", s.companyUpdateHandlesVariants())
						route.PUT("/companies/:id/tax-receipts", s.companyUpdateTaxReceipts())
						route.POST("/users", s.storeUserHandler())
						route.PUT("/users/:id", s.updateUserHandler())
					})
				})

			route.
				WithoutGroupMiddleware(RestrictedAccess).
				Group(func(route *routing.Router) {
					route.POST("/logout", s.logoutHandler)
					route.POST("/password", s.createPasswordHandler())

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
