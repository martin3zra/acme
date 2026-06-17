# Package Boundaries

> Phase 1 analysis. Dependency direction, layering violations, and circular-dependency check for the
> `framework → skeleton → business` split. No code changed.

## Per-package dependency audit

| Package | Imports (notable) | Imported by | Direction | Classification |
| --- | --- | --- | --- | --- |
| `pkg/store` | embed, stdlib | `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/console` | stdlib | `cmd/cli`, `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/cache` | `pkg/database` | `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/mailer` | net/smtp, templates | `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/inertia` | gonertia, vite | `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/database` | `lib/pq` | `pkg/cache`, `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/session` | `pkg/database` | `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/validator` | embed (en/es) | `pkg/support`, `app/*` | app → pkg | `[REUSABLE]` |
| `pkg/i18n` | embed (es only) | `app/*` | app → pkg | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| `pkg/routing` | **`pkg/foundation`**, gonertia | `app/*` | pkg → pkg + app → pkg | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| `pkg/support` | **`pkg/foundation`**, `pkg/validator` | `app/*` | pkg → pkg + app → pkg | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| `pkg/auth` | **`pkg/foundation`** | `app/*` | pkg → pkg + app → pkg | `[NEEDS REFACTOR BEFORE EXTRACTION]` |
| `pkg/foundation` | stdlib, `lib/pq` | `routing`, `support`, `auth`, `app/*` | leaf, widely imported | `[NEEDS REFACTOR BEFORE EXTRACTION]` (grab-bag) |

## Dependency violations blocking extraction

These are **business concepts embedded in reusable packages**, not import cycles. Each must be
resolved before the host package can move to `forge`.

1. **`pkg/foundation` grab-bag** — bundles `User` + money + hash + fs + helpers. `foundation/user.go`
   hardcodes `Role string` and embeds raw SQL against a `users` table
   (`UPDATE users SET password…`). This is the **keystone violation**: three packages import
   `foundation` only to reach `User`, dragging app schema into the framework layer.
2. **`pkg/support` tenant keys** — `form-request.go` declares `AccountKey{}` and `CompanyKey{}`
   (business/tenant concepts) inside a generic FormRequest base.
3. **`pkg/auth` schema coupling** — generic auth mechanism assumes `email` / `users` schema via
   `foundation.User`.
4. **`pkg/i18n` / `pkg/foundation/money.go` locale lock** — i18n embeds only `locales/es.json`;
   `money.go` `numberToWords` is Spanish-only. Locale hardcoded into otherwise-generic code.

## Circular-dependency check

**Result: no import cycles.** All edges flow `app → pkg`; the only intra-`pkg` edges are
`routing/support/auth/cache/session → foundation/database/validator` (downward, acyclic). App never
imports back into itself in a cycle, and no `pkg` imports `app`. Confirmed: the dependency *direction*
is clean.

## Layering audit (`framework → skeleton → business`)

| Rule | Status | Detail |
| --- | --- | --- |
| framework must not import skeleton or business | ✅ holds | No `pkg → app` edges found |
| business may import framework | ✅ holds | `app/*` imports `pkg/*` freely |
| framework must not name business concepts | ❌ violated | `foundation.User`/`Role`, `support.AccountKey`/`CompanyKey`, `auth` email schema, `i18n`/`money` Spanish locale |
| business must not bypass framework abstractions | ⚠️ partial | `app/*` frequently uses **inline SQL** in `*-repository.go` instead of `pkg/store`; only 2 `.sql` files exist (`companies_find_by_id`, `ping`). Acceptable for the app layer, but blocks any future "repository framework" extraction. |

### App-bypasses-abstraction cases
- Most queries are inline in `*-repository.go`, not stored under `app/sql/` (contradicts
  `.github/copilot-instructions.md`).
- Direct `lib/pq` / Postgres coupling in app repositories (e.g. `jsonb_set`/`#>>` in
  `sequence-formatter.go`, `upsert_tax_receipts($1,$2::jsonb)`). Fine for app layer; not portable.

## Recommendations

1. **Split `foundation` first.** It is the single keystone unblocking `routing`, `support`, and
   `auth`. Separate framework primitives (errors, hash, fs, helpers) from the app-owned `User`
   (move struct + SQL into `app/`); reusable packages depend on a small `Authenticatable` interface.
2. **Relocate tenant keys.** Move `AccountKey`/`CompanyKey` out of `pkg/support` into an
   app/skeleton package.
3. **Generalize locale.** Make `i18n` ship empty/`en`; move `numberToWords` behind a locale strategy.
4. **No cycle remediation needed** — direction is already clean; effort is purely de-coupling
   business concepts from `pkg`.
5. **Defer "repository framework"** — inline-SQL prevalence means a generic repository layer is not
   a near-term extraction target; revisit only if `app/sql/` adoption grows.
