# Wave 1 Extraction Plan — Clean Packages → `forge`

> Goal: move low-risk, business-free packages into a separate `forge` Go module, proving the
> multi-module split mechanics before touching the `foundation` keystone. Plan only — no code yet.

## Reality check vs Phase 1

Phase 1 listed 8 "immediate" packages. Verified imports show **2 are not clean**:

| Package | Internal imports | Verdict |
| --- | --- | --- |
| `store` | none | ✅ clean |
| `mailer` | none | ✅ clean |
| `inertia` | none | ✅ clean |
| `database` | none | ✅ clean |
| `console` | none | ✅ clean |
| `cache` | `database` | ✅ clean (Tier-A only) |
| `session` | `foundation` (`Authenticatable`, `ErrorBag`, `GetIpAddress`) | ⚠️ needs primitive carve |
| `validator` | `auth`, `database` | ❌ drags `auth → foundation` |

### Why `session` is cheaper than feared
`session` only touches foundation **primitives** — `Authenticatable` (interface), `ErrorBag`
(`types.go`), `GetIpAddress` (`helpers.go`). It never touches the `User` struct. foundation is only
483 LOC; the business-coupled part is the ~30-line `User` struct + its inline SQL in
`pkg/foundation/user.go`. Pull `User` out and the rest of foundation is already framework-grade
(hash, helpers, flash, fs, types, money, the `Authenticatable`/`MustVerifyPassword` interfaces).

### Why `validator` is deferred
`pkg/validator/validates-attributes.go:150-152` calls `auth.User(ctx)` / `auth.NewAuth(ctx)` for an
authenticated-context rule, and `database-rule.go:130` reads `database.ConnectionKey{}`. The auth
dependency chains `validator → auth → foundation`. Defer to Wave 2 (after auth) **or** move the
single auth-dependent rule into the app and extract the rest of validator now.

## Tiers (execution order within Wave 1)

| Tier | Packages | Prerequisite | Risk |
| --- | --- | --- | --- |
| A | `store`, `mailer`, `inertia`, `database`, `console` | none | Low |
| B | `cache` | Tier A (`database`) extracted | Low |
| C | `session` | carve `foundation` primitives (move `User` → `app`) | Medium |
| (defer) | `validator` | Wave 2 (`auth`) or relocate auth-rule | — |

## Status

- **Tier A + B: DONE.** `store mailer inertia database console cache` moved to
  `github.com/martin3zra/forge` (sibling repo, `replace => ../forge` in `acme/go.mod`). 29 acme
  files import-rewritten. `forge` + `acme` build (incl. `-tags prod`), `go vet` clean, zero
  `forge → acme` import edges. gonertia pinned `v2.0.3` in both. `mailer_test` fails only on missing
  local SMTP (`:1025`) — environmental, pre-existing.
- **Wave 2 decouple: DONE (in-place, not yet moved to forge).** Foundation carve + auth/routing/support
  decouple landed as one unit:
  - `foundation` lost the `User` struct + users-table SQL; keeps primitives + the
    `Authenticatable` / `MustVerifyPassword` interfaces (`contracts.go`, `GetRole()` added). Verified
    zero `User`/`users` refs remain.
  - `auth` holds no user schema — app registers `CredentialResolver` / `PasswordResolver` /
    `UserDecoder` seams (`app/auth-user.go` `init()`).
  - `routing.Context.User()` and `support.FormRequest.User()` return `foundation.Authenticatable`.
  - `app.AuthUser` now owns the identity struct + SQL; `app.User` embeds it; `Can(Authenticatable)`.
  - acme builds (`-tags prod`), `go vet` clean. Only `pkg/routing` `TestWithRequest_Success` fails —
    confirmed pre-existing on clean `HEAD` (unrelated: `ParseRequest` returns `UnprocessableEntity`
    for non-`FormRequestContract` bodies → session panic in test without a session).
- **Clean cluster extracted: DONE.** `foundation`, `auth`, `session` moved to forge (commits: forge
  `99abe7f`, acme `3a406bd`). forge now holds 9 packages: `store mailer inertia database console cache
  foundation auth session`. Zero forge→acme import edges; both modules build (`-tags prod`), vet
  clean, pinned go 1.23.6 / x-crypto v0.1.0 / x-text v0.26.0.
- **`validator` + `support` + `routing` extracted: DONE** (commits: forge `32e8b76`, acme `7f365c3`).
  Unblocked by moving the tenant context keys `AccountKey`/`CompanyKey` out of `support` into the app
  (`app/context-keys.go`) — `support` no longer names business concepts, and `routing` (which used
  `support.ParseRequest`) followed. forge now holds **12 packages**: `store mailer inertia database
  console cache foundation auth session validator support routing`. Zero forge→acme edges; both
  modules build (`-tags prod`), vet clean, app tests pass. Pinned lib/pq v1.10.9.
- **`i18n` generalized + extracted: DONE** (commits: forge `3d8311d`, acme `0caa421`). i18n no longer
  embeds locale files or hardcodes a language — `LoadTranslations` and `NewTranslator` take an
  `fs.FS` plus primary/fallback language from the caller. The app now owns its locales
  (`app/locales/es.json` via `app/locales.go`, `defaultLang`/`fallbackLang` constants).
- **EXTRACTION COMPLETE.** `acme/pkg` no longer exists; acme is purely the application on top of the
  `forge` module. forge holds all **13 packages**: `store mailer inertia database console cache
  foundation auth session validator support routing i18n`. Zero forge→acme edges; both modules build
  (`-tags prod`), vet clean, app tests pass.
  - *Residual (non-blocking):* the `Trans` action-suffix grammar still defaults to Spanish `"o"/"a"`,
    but it is data-driven from `global.nouns.*.gender` keys in the app's locale file — a content
    concern, not a dependency lock.
  - *Remaining optional follow-ups:* tag a forge version and drop the local `replace` for a pinned
    require; push branches / open PRs.
- **Known pre-existing test failures (now in forge):** `mailer` `TestSendWithAttachment` (needs local
  SMTP `:1025`) and `routing` `TestWithRequest_Success` (session panic for non-`FormRequestContract`
  body) — both reproduce on the original `HEAD`, unrelated to extraction.

## Topology decision

**Single-module monorepo.** One repo `github.com/martin3zra/forge`, one `go.mod`, packages as
subdirs (`forge/database`, `forge/cache`, …). No org, no repo-per-pkg.

- **Standalone use is free:** consumers `go get github.com/martin3zra/forge` and import only the
  packages they need (`import "github.com/martin3zra/forge/database"`). Go compiles only imported
  packages into the binary — no meta/aggregate package required for "use one piece without the rest".
- **One version tag** for the whole library; no cross-package version matrix.
- **Not** repo-per-pkg / multi-module: avoids N go.mod, N tag streams, tag-prefix friction
  (`database/v1.2.0`), and cross-repo bumps for the existing internal deps (`cache→database`,
  `session→foundation`, `validator→auth→foundation`).
- **Split later only if forced:** if a single package gains independent traction, mirror it to a
  read-only repo via `git subtree` / `splitsh`, keeping mono-development. Reversible; repo-per-pkg
  is not.
- The starter/skeleton (laravel/laravel role) stays a **separate** concern (`acme-skeleton`), not
  part of the `forge` library repo.

## Module mechanics (proof-of-concept on Tier A)

1. **New module.** `git init` the `forge` repo with `module github.com/martin3zra/forge` in its
   `go.mod` (locally a sibling dir to `acme` during transition).
2. **Move packages.** Relocate Tier-A package dirs under the new module: `pkg/store` → `forge/store`,
   etc., rewriting import paths `…/acme/pkg/store` → `…/forge/store`.
3. **Rewrite imports in `app/`** to the new module path (mechanical, `gofmt`-safe).
4. **`replace` directive** in app `go.mod` during transition:
   `replace github.com/martin3zra/forge => ../forge` so the app builds against the
   working copy before the framework is tagged/published.
5. **Tidy + build.** `go mod tidy` both modules; `go build ./...`; run existing package tests.
6. **`build.sh` impact:** none for Tier A/B/C — all are pure Go libraries; the dual-binary build and
   vite/SSR pipeline are unaffected. Only the module graph changes.

## Foundation primitive carve (gate for Tier C)

Minimal, scoped — **not** the full Wave-2 foundation split:

- **Move out of foundation → `app/`:** the `User` struct + its inline SQL (`user.go`).
- **Keep in foundation (framework):** `types.go`, `hash.go`, `helpers.go`, `flash.go`,
  `fs-dev/prod.go`, `money.go`, and the `Authenticatable` / `MustVerifyPassword` interfaces (move
  these interfaces from `user.go` into `types.go` or a new `contracts.go`).
- After the carve, `session` imports only framework-grade foundation → extract to `forge`.

## Verification per tier

- `go build ./...` green in both modules.
- Existing tests pass (`pkg/validator/validator_test.go`, session tests, etc.).
- No `forge → acme` (app) import edge introduced — check with
  `grep -r "martin3zra/acme/app" framework/` (must be empty).
- App still boots (web + CLI binaries build under `-tags prod`).

## Recommendations / sequencing

1. **Start Tier A** — 5 zero-dependency packages. Smallest blast radius; establishes module + `replace`
   + import-rewrite workflow.
2. **Tier B (`cache`)** immediately after — single dep on already-moved `database`.
3. **Tier C (`session`)** after the small foundation primitive carve (move `User` to `app`). This
   carve is independently valuable — it also unblocks Wave 2 (`routing`/`support`/`auth`).
4. **Hold `validator`** for Wave 2, or split its one auth-rule out if Wave 1 momentum justifies it.
5. Keep the `replace` directive until the framework module is tagged; only then pin a version.
