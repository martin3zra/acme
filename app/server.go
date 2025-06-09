package app

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

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
}

func NewServer(assets, resources *embed.FS) *Server {

	qs, err := store.NewQueryStore(sqlQueriesFS, "sql/")
	if err != nil {
		panic(err)
	}
	translator := i18n.NewTranslator(loadTranslations("global"))

	server := &Server{
		qs:         qs,
		config:     LoadConfig(),
		translator: translator,
	}

	if assets != nil && resources != nil {
		server.assets = *assets
		server.resources = *resources
	}

	return server
}

func (s *Server) Boot() {

	s.registerPermissions()
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
	s.route.RegisterInertia(
		inertia.InitInertia(
			s.assets,
			s.resources,
			s.config.port,
		),
	)
	s.bootRoutes()
}

func (s *Server) Start() {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.port),
		Handler: s.sessionManager.Handle(s.BindMiddleware(s.route)),
	}

	server.ListenAndServe()
}

func (s *Server) Shutdown() {
	log.Print("Shutting down server")
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

func (s *Server) registerPermissions() {
	modules := []string{
		"company", "user", "customer", "invoice", "payment",
	}

	actions := []string{"delete", "update", "view", "create", "store", "viewAny"}

	// Grouped by action
	var permissions []string

	for _, action := range actions {
		for _, module := range modules {
			permissions = append(permissions, fmt.Sprintf("%s:%s", action, module))
		}
	}

	fmt.Println("✅ permissions.", permissions)
}
