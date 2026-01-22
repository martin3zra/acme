package app

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/martin3zra/acme/pkg/foundation"
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

var rolePermissionsCache = map[string]map[string]bool{}

var groupedPermissions = map[string]map[string][]string{
	"owner": {"*": {"*"}},
	"admin": {
		"view":    {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports", "setting"},
		"viewAny": {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports", "setting"},
		"create":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports", "setting"},
		"delete":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports", "setting"},
		"update":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports", "setting", "company:sequence"},
	},
	"supervisor": {
		"view":    {"dashboard", "customer", "item", "payment", "reports"},
		"viewAny": {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports"},
		"create":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports"},
		"delete":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports"},
		"update":  {"dashboard", "invoice", "estimate", "customer", "item", "payment", "reports"},
	},
	"standard": {
		"view":   {"invoice", "customer"},
		"create": {"invoice"},
	},
}

func permissions(role string) map[string]bool {
	if cached, exists := rolePermissionsCache[role]; exists {
		return cached
	}

	flatPermissions := make(map[string]bool)

	if rolePermissions, exists := groupedPermissions[role]; exists {
		for action, modules := range rolePermissions {
			for _, module := range modules {
				flatPermissions[action+":"+module] = true

				// If role has full module access ("view:*"), create general key
				if module == "*" {
					flatPermissions[action+":*"] = true
				}

				// If role has full action access ("*:invoice"), create wildcard key
				if action == "*" {
					flatPermissions["*:"+module] = true
				}
			}
		}

		// If role has full access ("*:*"), create a general wildcard key
		if _, exists := rolePermissions["*"]; exists {
			flatPermissions["*"] = true
		}
	}

	rolePermissionsCache[role] = flatPermissions
	return flatPermissions
}

func Can(user *foundation.User, actionModule string) bool {
	permissions := permissions(user.Role)

	// If the user requests "*" (full access check), return true if full access exists
	if actionModule == "*" {
		return permissions["*"]
	}

	// Standard permission checks
	if permissions[actionModule] {
		return true
	}

	// Action-wide wildcard (e.g., "view:*")
	action := actionModule[:strings.Index(actionModule, ":")]
	if permissions[action+":*"] {
		return true
	}

	// Module-wide wildcard (e.g., "*:invoice")
	module := actionModule[strings.Index(actionModule, ":")+1:]
	if permissions["*:"+module] {
		return true
	}

	// Check for complete wildcard "*:*"
	return permissions["*"]
}
