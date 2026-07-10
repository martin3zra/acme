# Playwright / End-to-End Tests

> **Status: not established.** There are no browser/end-to-end tests today, and
> Playwright is not a dependency. This document records the gap and a concrete
> path to adopt it, so no one assumes the frontend is covered end-to-end.

## What exists today

- **Frontend gates only**: CI runs `eslint`, `prettier --check`, and `tsc
  --noEmit` on the React/Inertia code (`.github/workflows/ci.yml`). These catch
  type and style errors, **not** behaviour.
- No component tests, no e2e. The closest thing to running the whole app is the
  `/run` flow used during development (boot the Go server against the dev DB and
  hit routes with `curl`), which is manual and not a test.

## Why it's a real gap

Business-critical flows cross the browser boundary — invoice creation with line
editing, payment allocation, the CSV import with its SSE progress stream, the
variant matrix. The Go flow tests cover the server side of these, but nothing
asserts that the React UI drives them correctly.

## Recommended approach

Playwright against a locally-booted app, wired into CI as a separate job:

1. **Add Playwright**: `npm i -D @playwright/test`, `npx playwright install`.
2. **Boot the real stack for tests**: a dedicated test Postgres (reuse the
   `.env.test` DB / the CI `postgres:17` service), the Go server built with the
   embedded frontend (`npm run build` then `go build .`), seeded with a known
   tenant. A Playwright `webServer` config can start the binary.
3. **Seed deterministically**: expose a test-only seed path (a CLI subcommand or
   an env-guarded endpoint) that provisions the same tenant the Go
   `mkAccountCompany` fixture builds, so UI specs have a known starting state.
4. **First specs** — the highest-value flows:
   - log in → create an invoice with two lines → verify totals and the printed doc;
   - record a payment against that invoice → verify the balance updates;
   - run a CSV customer import → assert the SSE progress toast reaches "completed".
5. **CI job**: a `playwright` job that builds the frontend + binary, starts the
   Postgres service, runs `npx playwright test`, and uploads the HTML report /
   traces on failure. Start non-blocking, promote to blocking once stable.

## Notes specific to this app

- The SSE endpoint is on a **separate port** and its URL is delivered to the
  frontend as an Inertia shared prop — a Playwright run must set that env so the
  browser dials the right host (see the logging/SSE configuration).
- Inertia navigations are client-side; prefer Playwright's web-first assertions
  (`expect(locator).toBeVisible()`) over hard waits.

## Related files

- `.github/workflows/ci.yml` — current frontend gates (where a `playwright` job
  would slot in)
- `resources/js/` — the React/Inertia app under test
- `build.sh` — how the production binary bundles the frontend
