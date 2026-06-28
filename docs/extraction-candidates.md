# Extraction Candidates

> Phase 1 analysis. Each package/module is classified for extraction into `forge`,
> retention as `acme-skeleton`, or as `[BUSINESS-SPECIFIC]` application code. No code changed.

## Immediate Extraction Candidates

These are generic today and can move to `forge` with little or no change.

| Package | Reason | Dependencies | Risk | Recommendation |
| --- | --- | --- | --- | --- |
| `pkg/store` | Embedded `.sql` query loader, fully generic | embed only | Low | Extract as-is |
| `pkg/console` | Stdin prompts, no business coupling | stdlib | Low | Extract as-is |
| `pkg/cache` | `PgCache` over generic `preview_cache` table | DB | Low | Extract; document required table schema |
| `pkg/mailer` | SMTP/API transport + templates | net/smtp, templates | Low | Extract as-is |
| `pkg/inertia` | Vite/SSR bootstrap | gonertia, vite | Low | Extract as-is |
| `pkg/database` | Querier / transaction / bulk helpers | `lib/pq` | Low | Extract as-is |
| `pkg/session` | DB-backed session manager + GC | DB | Low | Extract; document `sessions` table schema |
| `pkg/validator` | Rule engine with en/es locales | embed | Low | Extract as-is |
| `pkg/routing` *(after decoupling)* | Laravel-style router/middleware/context | imports `pkg/foundation` (User) | Low once `foundation.User` removed | Extract after the foundation split below |

## Extraction Candidates Requiring Refactoring

### `pkg/support/form-request.go`
- **Why not today:** declares `AccountKey{}` and `CompanyKey{}` (business/tenant concepts) inside an
  otherwise generic FormRequest base.
- **Coupling:** tenant context keys leak into a generic request-validation primitive.
- **Hidden assumptions:** every consumer expects account/company to be resolvable from context.
- **DB deps:** none directly.
- **App deps:** business middleware populates the keys it declares.
- **Strategy:** move `AccountKey`/`CompanyKey` to an app/skeleton package; have FormRequest accept a
  generic context-key interface or leave tenant resolution to the app.
- **Complexity:** **Medium**

### `pkg/foundation` (split the grab-bag)
- **Why not today:** `foundation` bundles `User` + money + hash + fs + helpers. `User`
  (`foundation/user.go`) hardcodes `Role string` and embeds raw SQL against a `users` table
  (`UPDATE users SET password…`), coupling a "foundation" type to the app schema.
- **Coupling:** `routing`, `support`, `auth` all import `foundation` only for `User`/errors, dragging
  business shape into reusable packages.
- **Hidden assumptions:** a `users` table with `email`/`password`/`role` columns.
- **DB deps:** inline SQL in `foundation/user.go`.
- **App deps:** auth flow, middleware.
- **Strategy:** split into `framework` primitives (errors, hash, fs, helpers) vs. app-owned `User`
  (move the struct + its SQL into `app/`). Reusable packages depend on a small `Authenticatable`
  interface, not a concrete `User`.
- **Complexity:** **High** (touches routing/support/auth import edges)

### `pkg/auth`
- **Why not today:** generic mechanism but assumes `email` / `users` schema via `foundation.User`.
- **Coupling:** to `foundation.User`.
- **Hidden assumptions:** credential = email; user persistence in `users`.
- **DB deps:** via `foundation.User` SQL.
- **App deps:** login handlers.
- **Strategy:** depend on an `Authenticatable` / user-provider interface supplied by the app.
- **Complexity:** **Medium** (gated on the foundation split)

### `pkg/i18n`
- **Why not today:** only `locales/es.json` embedded; Spanish is the de-facto hardcoded locale.
- **Coupling:** loader generic, but shipped content is Spanish-only.
- **Hidden assumptions:** `es` always present as fallback.
- **DB deps:** none.
- **App deps:** `use-translation` frontend hook consumes injected `translations` prop.
- **Strategy:** ship the loader empty (or with `en` default); app provides locale files.
- **Complexity:** **Low–Medium**

### `pkg/foundation/money.go`
- **Why not today:** `numberToWords` is Spanish-only ("uno/dos/…"); locale-locked.
- **Coupling:** locale baked into formatting.
- **Strategy:** extract money math as `[REUSABLE]`; move `numberToWords` behind a locale strategy or
  into the app locale layer.
- **Complexity:** **Medium**

## Skeleton-Only Components

Retained in `acme-skeleton` — reusable *shape*, business wiring inside; not framework, not deleted.

- `app/server.go` — Boot/Start/Shutdown/SSE/scheduler lifecycle.
- `app/middleware.go` — auth/verified/shared-props chain (reusable patterns mixed with
  company/account session attrs).
- `app/main.go` & `cmd/cli/main.go` — dual-binary bootstrap.
- `build.sh` — vite build → manifest copy → `go build -tags prod` web+CLI → version zip.
- `foundation/fs-dev.go` / `fs-prod.go` — build-tag FS switching.
- `app/config.go` — env loading.
- Frontend: `app.jsx` / `ssr.jsx` Inertia+SSR bootstrap, `components/ui/*` kit, `data-table`,
  layouts shell, generic hooks (`use-debounced`, `use-local-storage`, `use-mobile`),
  `lib/utils.ts`, `lib/event-bus.ts`, `import-zone`.

## Business-Only Components

`[BUSINESS-SPECIFIC]` — never extract.

- All domain `*-handlers.go` / `*-repository.go` / `*-mapper.go`.
- `app/acl.go` — concrete `groupedPermissions` (invoice/customer/purchase/…).
- `app/route.go` — concrete route registrations.
- PDF generators, `app/invoice-scheduler.go`, `app/sequence-formatter.go`.
- Taxes / sales / expenses reporting.
- `resources/js/Pages/*` (Invoices, Customers, Purchases, Inventories, Payables, Reports).
- Domain components (`create-company-form`, money-input semantics, `nav-*`).
- Business `types/index.ts` (User/Company/Auth/Account).

## Recommendations

1. **Wave 1 (no refactor):** extract the 8 immediate packages + frontend `ui`/hooks/utils into
   `forge`. Lowest risk, immediate payoff.
2. **Wave 2 (unblock routing/auth):** perform the `foundation` split first — it is the keystone
   blocker. Then `routing`, `support`, `auth` follow.
3. **Wave 3 (locale):** generalize `i18n` and `money` locale handling.
4. **Skeleton:** template `server.go` / `middleware.go` / bootstrap / `build.sh` with business
   wiring stubbed; emit domain triads via generator, not as framework code.
5. **Sequence matters:** do not attempt `routing`/`auth`/`support` extraction before the
   `foundation` split — they share the same `foundation.User` coupling.
