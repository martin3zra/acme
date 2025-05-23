package app

import (
	"html/template"
	"log"
	"net/http"

	"github.com/justinas/alice"
	"github.com/martin3zra/acme/pkg/foundation"
	"github.com/martin3zra/acme/pkg/inertia"
	"github.com/romsar/gonertia/v2"
)

func (s *Server) bootRoutes() {
	guestMiddleware := s.registerGuestMiddlewares()
	authMiddleware := s.registerAuthMiddlewares()

	s.registerRoutes(
		guestMiddleware,
		authMiddleware,
		inertia.InitInertia(s.assets, s.resources, s.config.port),
	)
}

func (s *Server) registerRoutes(guest, auth alice.Chain, inertia *gonertia.Inertia) {

	s.get("/login", guest.Then(s.loginHandler(inertia)))
	s.post("/login", guest.Then(s.authHandler(inertia)))

	s.get("/verify-account/{uuid}/{hash}", guest.Then(s.verifyAccountHandler(inertia)))

	s.post("/logout", auth.Then(s.logoutHandler(inertia)))

	s.get("/home", auth.Then(s.homeHandler(inertia)))

	s.get("/customers", auth.Then(s.customersHandler(inertia)))
	s.post("/customers", auth.Then(s.storeCustomerHandler(inertia)))
	s.put("/customers/{id}", auth.Then(s.updateCustomerHandler(inertia)))
	s.put("/customers/{id}/change-status", auth.Then(s.changeStatusCustomerHandler(inertia)))
	s.delete("/customers/{id}", auth.Then(s.deleteCustomerHandler()))

	s.get("/items", auth.Then(s.itemsHandler(inertia)))
	s.post("/items", auth.Then(s.storeItemHandler(inertia)))
	s.put("/items/{id}", auth.Then(s.updateItemHandler(inertia)))
	s.put("/items/{id}/change-status", auth.Then(s.changeStatusItemHandler(inertia)))
	s.delete("/items/{id}", auth.Then(s.deleteItemHandler()))

	s.get("/invoices", auth.Then(s.invoicesHandler(inertia)))
	s.post("/invoices", auth.Then(s.storeInvoiceHandler(inertia)))
	s.get("/invoices/create", auth.Then(s.createInvoiceHandler(inertia)))
	s.get("/invoices/{id}/edit", auth.Then(s.editInvoiceHandler(inertia)))
	s.put("/invoices/{id}/void", auth.Then(s.voidInvoiceHandler(inertia)))
	s.put("/invoices/{id}", auth.Then(s.updateInvoiceHandler(inertia)))

	s.get("/payments", auth.Then(s.paymentsHandler(inertia)))
	s.post("/payments", auth.Then(s.storePaymentHandler(inertia)))
	s.get("/payments/create", auth.Then(s.createPaymentHandler(inertia)))
	s.get("/payments/{id}/edit", auth.Then(s.editPaymentHandler(inertia)))
	s.put("/payments/{id}/void", auth.Then(s.voidPaymentHandler(inertia)))
	s.put("/payments/{id}", auth.Then(s.updatePaymentHandler(inertia)))

	uiAssets := foundation.GetBuildAssets(s.assets, "public/build")
	s.mux.Handle("/", http.FileServer(http.FS(uiAssets)))

	s.mux.Handle("/build/", http.StripPrefix("/build/", http.FileServer(http.FS(uiAssets))))

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
