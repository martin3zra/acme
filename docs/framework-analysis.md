# Framework Extraction Analysis — Phase 1

> **Scope:** Analysis only. No extraction, refactor, or scaffolding performed. This document records
> the verified state of the `acme` codebase to evaluate splitting it into three layers:
>
> ```
> forge  →  acme-skeleton  →  business application
> ```
>
> **`forge`** is the reusable framework layer — projects are forged from it.
>
> Every significant subsystem is tagged `[REUSABLE]`, `[BUSINESS-SPECIFIC]`, or
> `[NEEDS REFACTOR BEFORE EXTRACTION]`.

## Architecture Overview

Go + Inertia.js + React SSR business application (invoicing, customers, payments, expenses,
purchasing, inventory, reporting).

| Concern | Detail |
| --- | --- |
| Language/runtime | Go 1.23 |
| Database | PostgreSQL via `lib/pq` |
| Server↔client | gonertia (Inertia.js adapter) |
| Frontend | React 19 + Vite 6 + Tailwind v4 |
| Build target | Dual binary: web server + CLI |
| Deploy | systemd, version-tagged zip artifact |

### Code size map

| Layer | Size | Shape |
| --- | --- | --- |
| `pkg/` | ≈ 5,076 LOC, 13 packages | `auth cache console database foundation i18n inertia mailer routing session store support validator` |
| `app/` | ≈ 16,046 LOC | Flat domain files (`*-handlers.go`, `*-repository.go`, `*-mapper.go`) + `acl.go`, `route.go`, `middleware.go`, `server.go`, `types.go`, `config.go`, `model.go` |
| `resources/js` | 252 TS/TSX files | 139 Pages, 79 components (33 `ui/` primitives), plus `hooks/ composables/ layouts/ lib/ types/ even-channels/` |

### Documentation discrepancy

The only documented guidance lives in **`.github/copilot-instructions.md`**. No `CLAUDE.md` and no
`.claude/` directory exist. Claims were verified against source; most hold, several are overstated
(see Discrepancies).

## Framework Capabilities

| Capability | Location | Dependencies | Extraction difficulty | Framework suitability | Tag |
| --- | --- | --- | --- | --- | --- |
| Routing (Laravel-style) | `pkg/routing` | `pkg/foundation`, gonertia | Medium (decouple `foundation.User`) | High | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| Middleware chain | `app/middleware.go` | session, account/company attrs | Medium | Pattern reusable, wiring business | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| Lifecycle (Boot/Start/Shutdown/SSE/scheduler) | `app/server.go` | DB, config, scheduler | Medium | Shape reusable | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| FormRequest base | `pkg/support/form-request.go` | declares `AccountKey{}`, `CompanyKey{}` | Medium | High after key removal | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| Validation rule engine | `pkg/validator` | en/es locale embedded | Low | High | `[REUSABLE]` |
| Session manager (DB-backed + GC) | `pkg/session` | DB | Low | High | `[REUSABLE]` |
| Auth mechanism | `pkg/auth` | `foundation.User`, `users`/`email` schema | Medium | High after schema decouple | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| ACL / permissions | `app/acl.go` | concrete domain permissions | n/a (business) | None | `[BUSINESS-SPECIFIC]` |
| Context DI helpers | `pkg/*` context keys | — | Low | High | `[REUSABLE]` |
| Query store (`.sql` loader) | `pkg/store` | embed | Low | High | `[REUSABLE]` |
| Cache | `pkg/cache` (`PgCache`) | generic `preview_cache` table | Low | High | `[REUSABLE]` |
| Mailer | `pkg/mailer` | SMTP/API + templates | Low | High | `[REUSABLE]` |
| Config loader | `app/config.go` | env | Low | Skeleton | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| PDF generation | `app/*pdf*` | domain layout | n/a | None | `[BUSINESS-SPECIFIC]` |
| CLI | `pkg/console` + `cmd/cli/main.go` | stdin prompts | Low (console) | High | `[REUSABLE]` (console) |
| Inertia / SSR bootstrap | `pkg/inertia` | vite/SSR | Low | High | `[REUSABLE]` |
| Shared props / flash / response helpers | `app/middleware.go`, handlers | session | Medium | Pattern reusable | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| Pagination | `pkg/database` | DB | Low | High | `[REUSABLE]` |
| DB abstractions (querier/tx/bulk) | `pkg/database` | `lib/pq` | Low | High | `[REUSABLE]` |
| i18n | `pkg/i18n` | only `locales/es.json` embedded | Low–Med | High after locale generalization | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| Money / number-to-words | `pkg/foundation/money.go` | Spanish-only `numberToWords` | Medium | High after locale split | `[NEEDS REFACTOR BEFORE EXTRACTION]` |

## Business Capabilities

- **Stays app-specific (never extract):** every `*-handlers.go` / `*-repository.go` / `*-mapper.go`,
  `acl.go` concrete `groupedPermissions`, `route.go`, PDF generators, `invoice-scheduler.go`,
  `sequence-formatter.go`, taxes/sales/expenses reporting.
- **Could become scaffolding template:** the *shape* of a domain triad (handler+repository+mapper),
  a sample ACL group, a sample route registration — emitted by a generator, not shipped as framework.
- **Must never be extracted:** any concrete business rule (invoice recurrence math, sequence
  formatting, tax-receipt JSONB upserts).

## Infrastructure Capabilities

| Item | Location | Tag |
| --- | --- | --- |
| DB init / connection open | `Server.Boot()` | `[NEEDS REFACTOR BEFORE EXTRACTION]` (skeleton) |
| DI via `context.Context` keys | `database.ConnectionKey{}`, `ConfigKey{}`, `support.AccountKey{}`, `session.SessionContextKey{}` | mixed — generic mechanism, business keys |
| Startup / shutdown | `app/server.go`, `app/main.go` | `[NEEDS REFACTOR BEFORE EXTRACTION]` (skeleton) |
| Build / deploy | `build.sh` (vite → manifest copy → `go build -tags prod` web+CLI → version zip) + systemd | `[NEEDS REFACTOR BEFORE EXTRACTION]` (skeleton) |
| Env loading | `app/config.go` | skeleton |
| Binary generation | `main.go`, `cmd/cli/main.go` | skeleton |
| FS embedding | `//go:embed public/build` + `resources/views`; dev/prod switch via `foundation/fs-dev.go` / `fs-prod.go` build tags | skeleton |
| Background scheduler | invoice recurrence, started `10*time.Second` in `main.go` | `[BUSINESS-SPECIFIC]` (mechanism reusable) |
| SSE server | separate server on `:8090` | `[NEEDS REFACTOR BEFORE EXTRACTION]` (shape reusable) |

## Frontend Audit

| Module | Examples | Tag |
| --- | --- | --- |
| Bootstrap | `app.jsx`, `ssr.jsx` (Inertia+SSR) | `[REUSABLE]` |
| UI primitives | `components/ui/*` (33 Radix primitives), `data-table` | `[REUSABLE]` |
| Generic hooks | `use-debounced`, `use-local-storage`, `use-mobile` | `[REUSABLE]` |
| Utils | `lib/utils.ts`, `lib/event-bus.ts`, `import-zone` | `[REUSABLE]` |
| Layouts shell | `layouts/*` | `[REUSABLE]` |
| Gate / translation | `hooks/use-gate.ts`, `use-translation`, `types/inertia.d.ts` | `[NEEDS REFACTOR BEFORE EXTRACTION]` (duplicate backend contracts / depend on injected props) |
| Pages | `Pages/*` (Invoices, Customers, Purchases, Inventories, Payables, Reports) | `[BUSINESS-SPECIFIC]` |
| Domain components | `create-company-form`, money-input semantics, `nav-*` | `[BUSINESS-SPECIFIC]` |
| Business types | `types/index.ts` (User/Company/Auth/Account) | `[BUSINESS-SPECIFIC]` |

## JSON / JSONB / sparse-matrix audit

- **Real JSONB column:** `companies_settings.sequences` — manipulated via `jsonb_set` / `#>>` path
  queries in `app/sequence-formatter.go` (sparse nested config: per-doc-type `{prefix,padding,next}`).
- **App-level JSON columns** via `driver.Valuer` / `sql.Scanner` structs in `app/types.go` &
  `company-repository.go`: `ItemIdentifiers`, `TransactionSource`, `PurchaseSource`, `Recurrence`,
  `Discount`, `Payment` (cash/check/card/bt), `RedirectPreferences`, tax receipts
  (`upsert_tax_receipts($1,$2::jsonb)`).
- **Null pattern:** every `Scan()` does `if value == nil { return nil }` → leaves the Go struct at
  zero value; nothing coalesces NULL → `{}`. Session `attrs` map and `account` / `current_company`
  are nil-initialized then populated.
- **Standardization opportunity (document only):** default these columns to `'{}'::jsonb` at the DB
  level and/or initialize structs so callers never branch on null. No code change in this phase.

## Discrepancies vs `.github/copilot-instructions.md`

1. Claims `CLAUDE.md` / `.claude/` guidance — neither exists; only the copilot file.
2. "Store complex queries as `.sql` under `app/sql/`" — only 2 trivial files
   (`companies_find_by_id.sql`, `ping.sql`); reality is inline SQL in `*-repository.go`.
3. "Migrations: SQL in app/sql" — migrations actually in `db/migrations/` (1 file:
   `20260320_accounts_payable_split_status.sql`).
4. `types.go` "1560 lines" — close enough; holds all forms/DTOs/enums/JSON scanners.
5. Scheduler comment says "every 5 minutes"; `main.go` actually starts it at `10*time.Second`.

## Recommendations

- **Framework (extract first, low risk):** `store`, `console`, `cache`, `mailer`, `inertia`,
  `database`, `session`, `validator`. Plus frontend `ui/` kit, generic hooks, bootstrap, utils.
- **Framework (extract after refactor):** `routing` (drop `foundation.User`), `support`
  (remove `AccountKey`/`CompanyKey`), `auth` (decouple `users`/`email` schema), `i18n` (generalize
  locale loading), `foundation/money.go` (split Spanish `numberToWords`).
- **Skeleton:** `server.go` lifecycle, `middleware.go` chain, `main.go` / `cmd/cli/main.go`
  bootstrap, `build.sh`, `fs-dev`/`fs-prod`, config loader, Inertia bootstrap jsx, layouts.
- **Business (never extract):** all domain handlers/repos/mappers, `acl.go`, `route.go`, PDFs,
  reports, scheduler logic, `Pages/*`, domain components, business types.
- **Refactor-first blockers:** the `foundation` grab-bag (split User/money/hash/fs/helpers) and
  business concepts (`Account`/`Company`/`User`) named inside reusable packages. See
  `package-boundaries.md`.
