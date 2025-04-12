package app

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/martin3zra/acme/pkg/session"
	"github.com/martin3zra/acme/pkg/store"
)

//go:embed sql/*.sql
var sqlQueriesFS embed.FS

type Server struct {
	mux            *http.ServeMux
	qs             store.Query
	db             *sql.DB
	config         *Config
	sessionManager *session.SessionManager
	// Transient request data
	session   *session.Session
	assets    embed.FS
	resources embed.FS
}

func NewServer(assets, resources embed.FS) *Server {

	qs, err := store.NewQueryStore(sqlQueriesFS, "sql/")
	if err != nil {
		panic(err)
	}

	return &Server{
		mux:       http.NewServeMux(),
		qs:        qs,
		config:    LoadConfig(),
		assets:    assets,
		resources: resources,
	}
}

func (s *Server) Boot() {
	s.config.ensureHasBeenSet()

	s.openDatabaseConnection()
	s.configureSessionManager()
	s.bootRoutes()
}

func (s *Server) Start() {

	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", s.config.port),
		Handler: s.sessionManager.Handle(s.BindMiddleware(s.mux)),
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

func (s *Server) get(pattern string, handler http.Handler) {
	s.mux.Handle(fmt.Sprintf("GET %s", pattern), handler)
}

func (s *Server) post(pattern string, handler http.Handler) {
	s.mux.Handle(fmt.Sprintf("POST %s", pattern), handler)
}

func (s *Server) put(pattern string, handler http.Handler) {
	s.mux.Handle(fmt.Sprintf("PUT %s", pattern), handler)
}

func (s *Server) delete(pattern string, handler http.Handler) {
	s.mux.Handle(fmt.Sprintf("DELETE %s", pattern), handler)
}
