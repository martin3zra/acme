package app

import (
	"bytes"
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/martin3zra/acme/app/mail"
	"github.com/martin3zra/acme/pkg/i18n"
	"github.com/martin3zra/acme/pkg/inertia"
	"github.com/martin3zra/acme/pkg/mailer"
	"github.com/martin3zra/acme/pkg/routing"
	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/store"
)

//go:embed sql/*.sql
var sqlQueriesFS embed.FS

type Server struct {
	qs             store.Query
	db             *sql.DB
	config         *Config
	sessionManager *session.SessionManager
	mailer         mailer.Mailer
	// Transient request data
	session    *session.Session
	assets     embed.FS
	resources  embed.FS
	translator *i18n.Translator
	route      *routing.Router
	httpServer *http.Server
}

func NewServer(assets, resources embed.FS) *Server {

	qs, err := store.NewQueryStore(sqlQueriesFS, "sql/")
	if err != nil {
		panic(err)
	}
	translator := i18n.NewTranslator(trans("global", "companies"))

	server := &Server{
		qs:         qs,
		config:     LoadConfig(),
		translator: translator,
		assets:     assets,
		resources:  resources,
	}

	return server
}

func (s *Server) Boot() {

	s.config.ensureHasBeenSet()
	s.openDatabaseConnection()
	s.configureMailClient()

	isRunningInCli := os.Getenv("RUNNING_IN_CLI")
	if isRunningInCli == "YES" {
		return
	}

	s.configureSessionManager()
	s.configureRouting()
}

func (s *Server) configureRouting() {
	s.route = routing.New()
	s.route.RegisterResources(s.resources)
	s.route.RegisterInertia(
		inertia.InitInertia(
			s.assets,
			s.resources,
			s.config.port,
		),
	)
	s.bootRoutes()
}

func (s *Server) Start() error {

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.port),
		Handler: s.sessionManager.Handle(s.BindMiddleware(s.route)),
	}

	log.Printf("HTTP server listening on %s", s.httpServer.Addr)

	err := s.httpServer.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) StartSSE() error {
	mux := http.NewServeMux()

	mux.HandleFunc("/sse/imports/", s.importEventsHandler)

	server := &http.Server{
		Addr:    ":8090", // separate port = no interference
		Handler: mux,
	}

	log.Println("SSE server listening on :8090")
	err := server.ListenAndServe()
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}

	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Print("Shutting down server")
	// Stop accepting new connections
	s.httpServer.SetKeepAlivesEnabled(false)

	// Attempt graceful shutdown
	if err := s.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("http shutdown: %w", err)
	}

	if err := s.db.Close(); err != nil {
		return fmt.Errorf("db close: %w", err)
	}

	return nil
}

func (s *Server) configureSessionManager() {
	s.sessionManager = session.NewSessionManager(
		session.NewDatabaseStore(s.db),
		30*time.Minute,
		time.Duration(s.config.session.lifetime)*time.Minute,
		12*time.Hour,
		s.config.session.cookie,
		s.config.session.domain,
		s.config.session.secure,
		s.config.session.httpOnly,
	)
}

func (s *Server) configureMailClient() {
	s.mailer = mailer.New(s.config.mail.asMailConfig(), s.resources)
}

func (s *Server) trans(key string, replacements ...i18n.Replacements) string {
	return s.translator.Trans(key, replacements...)
}

func (s *Server) abortWhenPrerequisiteMissing(ctx *routing.Context, resource string) bool {
	cache, ok := ctx.Request.Context().Value(prereqCacheKey).(prereqCache)
	if !ok {
		return false
	}

	company := CurrentCompany(ctx.Request.Context())
	key := fmt.Sprintf("%s:%d", resource, company.ID)
	result, ok := cache[key]

	if result.Ok || len(result.Missing) == 0 {
		return false
	}
	ctx.Render("Error/Prerequisites", map[string]any{
		"resource": resource,
		"missing":  result.Missing,
	})
	return true
}

func (s *Server) StartScheduler(ctx context.Context, interval time.Duration) {

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		log.Println("scheduler started")
		running := false

		for {
			select {
			case <-ticker.C:
				if running {
					log.Println("scheduler already running, skipping tick")
					continue
				}

				running = true
				if err := s.runRecurrenceScheduler(); err != nil {
					log.Printf("Scheduler error: %v", err)
				}
				running = false

			case <-ctx.Done():
				log.Println("scheduler shutting down...")
				return
			}
		}
	}()
}

func (s *Server) enqueueInvoiceEmail(companyID int, invoiceUUID string) {
	// We need to log those invoices as sent or not?

	type invoiceData struct {
		Header *invoice `json:"header"`
		Lines  []*line  `json:"lines"`
	}
	ctx := context.Background()
	invoice, err := s.findInvoicesByUUID(ctx, TransactionKinds.Invoice, companyID, invoiceUUID)
	if err != nil {
		log.Println("error fetching invoice: ", invoiceUUID, err)
		return
	}

	lines, err := s.findInvoiceLines(ctx, companyID, invoice.ID)
	if err != nil {
		log.Println("error fetching invoice lines: ", invoiceUUID)
		return
	}

	data := invoiceData{
		Header: invoice,
		Lines:  lines,
	}

	invoicePDF, err := NewInvoicePDF(s.translator, data.Header, data.Lines)
	if err != nil {
		log.Println("error generating invoice PDF", invoiceUUID)
		return
	}

	company, err := s.findCompanyByID(ctx, invoice.CompanyID)
	if err != nil {
		log.Println("error fetching company while generating invoice PDF", invoiceUUID)
		return
	}
	invoicePDF.Header(company)
	invoicePDF.Lines()
	invoicePDF.Footer(s.config.appName)

	var buf bytes.Buffer
	err = invoicePDF.pdf.Output(&buf)
	if err != nil {
		log.Println("error sending PDF document to the writer ", err)
		return
	}

	s.mailer.
		To(
			invoice.Customer.Email,
			invoice.Customer.Name,
		).Send(mail.NewInvoiceMail(
		map[string]any{
			"header": invoice,
			"lines":  lines,
		}, buf.Bytes()))

}
