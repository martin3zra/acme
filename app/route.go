package app

import (
	"html/template"
	"log"
	"net/http"

	"github.com/martin3zra/acme/pkg/auth"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/routing"
)

func (s *Server) bootRoutes() {

	s.route.WithMiddleware(s.SharedProps)

	s.route.
		WithMiddleware(auth.RedirectIfAuthenticated).
		Group(func(route *routing.Router) {
			route.GET("/login", s.login)
			route.POST("/login", s.authHandler)

			route.GET("/verify-account/:uuid/:hash", s.verifyAccountHandler)
			route.POST("/email/verification-notification", s.sendVerificationEmail)
		})

	s.route.
		WithMiddleware(auth.Middleware).
		WithoutGroupMiddleware(auth.RedirectIfAuthenticated).
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
		})

	uiAssets := foundation.GetBuildAssets(s.assets, "public/build")
	s.route.FileServer("/build/", http.FS(uiAssets))
}

func (s *Server) handleError(w http.ResponseWriter, err error, callbacks ...func()) {
	var titleHttpCode = map[int]string{
		500: "Internal Error.",
		403: "Forbidden.",
		401: "Unauthorized.",
	}

	var statusCode = http.StatusInternalServerError
	formatterError, ok := err.(foundation.ErrorFormatter)
	if ok {
		statusCode = formatterError.Status()
	} else {
		if len(callbacks) > 0 {
			callbacks[0]()
			return
		}
	}

	title, ok := titleHttpCode[statusCode]
	if !ok {
		title = "Something went wrong."
	}
	// display errors when on dev mode. otherwise logged this error.
	data := make(map[string]any)
	data["title"] = title
	data["message"] = err.Error()
	data["status"] = statusCode
	if s.config.isProduction {
		data["message"] = "Something happened, please contact the developer for support."
	}
	errorViewFile := "./resources/views/error/500.html"

	tmpl, _ := template.ParseFiles(errorViewFile)
	tmplErr := tmpl.Execute(w, data)

	if tmplErr != nil {
		log.Println(err.Error())
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}
