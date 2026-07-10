# API Tests

> **Status: not established.** There is no dedicated API-contract test suite
> today. This document records what exists and a recommended approach, so the
> gap is explicit rather than assumed-covered.

## Why there is no separate API layer

This app is **Inertia-server-rendered**: most "endpoints" return an Inertia page
(HTML + props), not JSON, and are already exercised through handler/flow tests
([05-integration-tests.md](05-integration-tests.md)). There is no public REST/JSON
API with a stable contract that an external client consumes, so there is nothing
to pin with a contract suite yet.

## The JSON surface that does exist

A handful of endpoints speak JSON (fetched by the React frontend, or streamed):

- **Item / import endpoints** — `app/item-handlers.go` uses `ctx.JSON(...)` for
  item lookups, CSV import preview/validation, and error envelopes.
- **Detail fetches** — a few routes return JSON when the request sets
  `Accept: application/json` (e.g. the invoice/purchase "convert" and detail
  fetches the frontend calls with `fetch`).
- **Recurrence** — `POST /invoices/:id/recurrence`.
- **SSE** — `GET /sse/imports/:id` (`text/event-stream`), the import progress
  stream. Served on the separate SSE port (see the logging/SSE work).

These are covered indirectly: the handlers behind them run in flow tests, and the
repositories they call are pinned by `flow_*_test.go`.

## Recommended approach (when this becomes worth it)

If a first-class JSON API emerges (mobile client, integrations), add contract
tests at the handler boundary using the existing harness — no new framework
needed:

1. Use `newHandlerServer(t)` + `handlerCtx(t, s, f, "GET", "/path", nil)` with an
   `Accept: application/json` header.
2. Call the handler, then assert on `rec.Body` (decode JSON) and `rec.Code`.
3. Keep tenant scoping in the assertions — every response must be
   company-scoped.

For the SSE stream, drive `importEventsHandler` with an `httptest` recorder that
implements `http.Flusher`, publish an event via `emit`, and assert the framed
`data:` payload.

## Related files

- `app/item-handlers.go` — the main JSON surface (`ctx.JSON`), plus
  `importEventsHandler` and `emit` (the SSE stream)
- `app/server.go` — registers the `/sse/imports/` mux on the SSE server
- `app/route.go` — route table
- [02-fixtures.md](02-fixtures.md) — the harness these tests would reuse
