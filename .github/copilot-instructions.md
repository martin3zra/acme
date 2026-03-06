# Acme Codebase Guide

## Architecture Overview

This is a **Go + Inertia.js + React SSR** business application combining a custom Go backend with TypeScript/React frontend rendered via Inertia.js. The app manages invoicing, customers, payments, expenses, and financial reporting.

### Stack
- **Backend**: Go 1.23+, PostgreSQL, custom routing/middleware packages
- **Frontend**: React 18, TypeScript, Inertia.js v2, Vite 6, Tailwind CSS v4
- **Deployment**: Dual binary build (web app + CLI), systemd service

## Project Structure

```
app/                      # Business logic (handlers, repositories, types)
  *-handlers.go          # HTTP handlers per domain
  *-repository.go        # Database operations per domain  
  *-mapper.go            # Transform data between layers
  types.go               # FormRequests, DTOs, enums (1560+ lines)
  acl.go                 # Role-based permission system
  route.go               # Route definitions
  middleware.go          # Request middleware
  sql/                   # Embedded SQL queries
  
pkg/                     # Reusable packages
  routing/               # Custom Laravel-inspired router
  session/               # Session management
  validator/             # Form validation
  i18n/                  # Internationalization
  auth/, cache/, mailer/, etc.
  
resources/
  js/                    # React frontend
    Pages/               # Inertia page components
    components/          # Reusable UI components
    app.jsx             # Client-side entry
    ssr.jsx             # Server-side entry
  css/app.css           # Tailwind styles
```

## Core Patterns

### 1. Handler → Repository → Mapper Flow

Each business domain follows this pattern (see [app/customer-handlers.go](app/customer-handlers.go), [app/customer-repository.go](app/customer-repository.go)):

```go
// Handler: HTTP request/response
func (s *Server) storeCustomerHandler() routing.HandlerFunc {
    return routing.WithRequest(func(ctx *routing.Context, form *StoreCustomerForm) {
        // Validation automatic via FormRequest.Rules()
        err := s.storeCustomer(ctx.Request.Context(), form)
        if err != nil {
            ctx.BackWithError(err)
            return
        }
        ctx.Flash("success", "Customer created")
        ctx.Redirect("/customers")
    })
}

// Repository: Database operations
func (s *Server) storeCustomer(ctx context.Context, form *StoreCustomerForm) error {
    _, err := s.db.ExecContext(ctx, "INSERT INTO customers (...) VALUES (...)")
    return err
}
```

### 2. FormRequest Pattern

All form inputs use `support.FormRequest` with automatic validation and authorization (see [app/types.go](app/types.go)):

```go
type StoreCustomerForm struct {
    support.FormRequest
    Name  string `json:"name"`
    Email string `json:"email"`
}

func (StoreCustomerForm) Rules() map[string]any {
    return map[string]any{
        "name": "required|min:3|max:120",
        "email": "required|email|unique:customers,email",
    }
}

func (form StoreCustomerForm) Authorize() bool {
    return Can(form.User(), "create:customer")
}
```

**Use `routing.WithRequest()` to bind and validate forms automatically.**

### 3. Custom Routing with Middleware

The routing package ([pkg/routing/](pkg/routing/)) provides Laravel-style fluent API:

```go
// Route groups with prefixes
s.route.GroupPrefix("api", func(api *routing.Router) {
    api.GET("status", statusHandler)
})

// Middleware chaining and exclusion
route.
    WithMiddleware(AuthMiddleware, VerifiedMiddleware).
    Group(func(route *routing.Router) {
        route.GET("/customers", s.customersHandler).Can("viewAny:customer")
        route.POST("/customers", s.storeCustomerHandler()).WithoutMiddleware(SomeMiddleware)
    })
```

**Key conventions:**
- Use `.Can("action:resource")` for permission checks
- `.WithoutMiddleware()` excludes specific middleware per route
- `.Middleware()` adds route-specific middleware

### 4. ACL System

Permission system in [app/acl.go](app/acl.go) uses `action:resource` format:

```go
// Check permissions
Can(user, "create:customer")    // Can user create customers?
Can(user, "viewAny:invoice")    // Can user list invoices?
Can(user, "*")                  // Full access?

// Roles: owner, admin, supervisor, standard
// Permissions defined in groupedPermissions map
```

### 5. Context-Based Dependency Injection

Server dependencies injected via `context.Context` (see [app/middleware.go](app/middleware.go)):

```go
// Access in handlers
db := ctx.Request.Context().Value(database.ConnectionKey{}).(*sql.DB)
config := ctx.Request.Context().Value(ConfigKey{}).(*Config)
session := ctx.Request.Context().Value(session.SessionContextKey{}).(*session.Session)

// Access current user/company
CurrentCompany(ctx.Request.Context())  // From session attrs
CurrentUser(ctx.Request.Context())     // From session
```

### 6. SQL Query Store

Embedded SQL files loaded into memory (see [pkg/store/query.go](pkg/store/query.go), [app/sql/](app/sql/)):

```go
// In Server initialization
qs, _ := store.NewQueryStore(sqlQueriesFS, "sql/")

// Usage
query := s.qs.Q("companies_find_by_id")  // Loads companies_find_by_id.sql
```

**Store complex queries as .sql files under [app/sql/](app/sql/)**

### 7. Inertia.js Integration

Frontend uses Inertia.js for SPA-like experience without API:

```go
// Backend: Render Inertia pages
ctx.Render("Customers/Index", map[string]any{
    "customers": customers,
    "translations": trans("customers"),
})

// Frontend: Pages in resources/js/Pages/
// Access props via Inertia's usePage() hook
```

**Shared props** (auth, csrf_token, flash, translations) injected globally via `SharedProps` middleware.

## Development Workflows

### Build & Run

```bash
# Frontend development (HMR enabled)
npm run dev

# Backend (auto-reload with Air or manual restart)
go run .

# Full production build (creates versioned binaries + zip)
./build.sh darwin amd64
# Outputs: bin/acme-darwin-amd64-v*.*.*, bin/acme-cli-darwin-amd64-v*.*.*
```

**Build process**:
1. Vite builds React app (client + SSR bundles)
2. Copies manifest.json to public/build/
3. Go embeds public/build and resources/views
4. Creates web + CLI binaries with version tags

### Environment Setup

Copy [.env.sample](.env.sample) to `.env`:
```bash
APP_KEY=base64:...  # Generate with openssl rand -base64 64
DB_NAME, DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT
SESSION_LIFETIME=120
TZ=America/Santo_Domingo  # Set in code, not .env
```

### Database

- **Driver**: PostgreSQL (`lib/pq`)
- **Migrations**: Not explicit; SQL in [app/sql/](app/sql/)
- **Connection**: Opened in `Server.Boot()`, context-injected

### Testing

- Unit tests: Minimal (see [app/invoice-scheduler_test.go](app/invoice-scheduler_test.go))
- **Convention**: Test files as `*_test.go`

## Key Conventions

1. **Domain files prefixed**: `customer-handlers.go`, `customer-repository.go`, `invoice-handlers.go`, etc.
2. **Types in one file**: [app/types.go](app/types.go) contains all FormRequests, DTOs, enums (1560 lines)
3. **Translations**: Use `trans("namespace")` helper; loads from `pkg/i18n/`
4. **Error handling**: `ctx.BackWithError(err)` for form errors, `ctx.Error(err)` for 500s
5. **Flash messages**: `ctx.Flash("success", "Message")` → shown via toast on frontend
6. **Frontend imports**: Use `@/` alias for `resources/js/` (configured in tsconfig.json)
7. **UTC offset**: Timezone hardcoded to `America/Santo_Domingo` in [main.go](main.go#L38)

## Common Tasks

### Add a new domain (e.g., "vendors")

1. Create `app/vendor-handlers.go`, `app/vendor-repository.go`, `app/vendor-mapper.go`
2. Add types to [app/types.go](app/types.go): `StoreVendorForm`, `vendor` struct
3. Define routes in [app/route.go](app/route.go)
4. Add permissions to [app/acl.go](app/acl.go) `groupedPermissions`
5. Create React page `resources/js/Pages/Vendors/Index.tsx`

### Add middleware

1. Define in [app/middleware.go](app/middleware.go):
```go
func MyMiddleware(next routing.HandlerFunc) routing.HandlerFunc {
    return func(ctx *routing.Context) {
        // Pre-processing
        next(ctx)
        // Post-processing
    }
}
```
2. Apply: `s.route.WithMiddleware(MyMiddleware)` or per-route `.Middleware(MyMiddleware)`

### Add permission check

1. Add to [app/acl.go](app/acl.go) `groupedPermissions` map
2. Use in routes: `.Can("action:resource")`
3. Use in handlers: `if !Can(user, "action:resource") { ... }`

### Add SQL query

1. Create `app/sql/query_name.sql`
2. Use: `s.qs.Q("query_name")`

## External Dependencies

- **gonertia**: Inertia.js Go adapter (`github.com/romsar/gonertia/v2`)
- **laravel-vite-plugin**: Vite integration for asset manifests
- **fpdf**: PDF generation (`codeberg.org/go-pdf/fpdf`)
- **Radix UI**: Headless UI components
- **Sonner**: Toast notifications

## Deployment

Service deployment via systemd (see [README.md](README.md)):
1. Build with `./build.sh linux amd64`
2. Copy binary to server
3. Create systemd service file
4. Enable and start service

The app includes:
- Graceful shutdown handling
- Background scheduler (invoice recurrence)
- SSE endpoint for real-time updates
- Logging to `acme.log`
