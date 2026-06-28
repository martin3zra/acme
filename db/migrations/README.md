# Database migrations (Camel)

Schema changes are managed with [Camel](https://github.com/martin3zra/camel).
Config lives in [`camel.yaml`](../../camel.yaml) at the repo root.

## Connecting

The DSN is supplied at runtime (no credentials in git):

```bash
export DB_SOURCE="postgres://postgres:PASSWORD@localhost:5433/acme?sslmode=disable"
camel status
```

## The baseline

`00000000000000_schema_baseline.yaml` is the initial snapshot of the existing
production schema. It is a single `action: raw` migration because the schema
uses native Postgres **enum types**, **plpgsql functions**, and **triggers**
that Camel's column DSL cannot express. Statements are raw SQL so they apply
verbatim.

It was generated from a schema-only dump:

```bash
docker exec -e PGPASSWORD=$DB_PASSWORD postgres \
  pg_dump -U postgres --schema-only --no-owner --no-privileges -d acme \
  > acme_schema.sql

python3 scripts/pgdump_to_camel.py acme_schema.sql \
  db/migrations/00000000000000_schema_baseline.yaml
```

The generator ([`scripts/pgdump_to_camel.py`](../../scripts/pgdump_to_camel.py))
splits the dump dollar-quote- and string-literal-aware, so function bodies and
comments containing `;` are not shredded. A YAML `statements:` list is used
(not a `.sql` file) because Camel's `.sql` loader splits on bare semicolons and
would corrupt plpgsql bodies.

Verified faithful: applying the baseline to a fresh database and re-dumping it
yields a schema identical to production (modulo Camel's own ledger table).

## Adopting Camel on the existing production database

The baseline **creates** objects, so it must not run against the prod database
where they already exist. Mark it as applied without executing it:

```sql
-- one-time, on the existing prod DB
INSERT INTO camel_migrations (migration, batch, applied_at)
VALUES ('00000000000000_schema_baseline.yaml', 1, now());
```

Fresh/CI databases instead run `camel migrate` normally to build the schema
from the baseline.

## Going forward

New schema changes are authored as readable Camel YAML migrations
(`create` / `alter` actions) and applied with `camel migrate`. Use
`action: raw` only when a change needs SQL the DSL can't express (new enum
type, function, trigger, partial index, check constraint).

## Legacy

Pre-Camel ad-hoc SQL migrations are retained for history under
[`db/legacy/`](../legacy/). Their end-state is already folded into the baseline;
they are not run by Camel.
