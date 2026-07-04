# Acme

Go + Inertia.js + React business app (invoicing, inventory, purchasing, AP/AR).

See **[docs/getting-started.md](docs/getting-started.md)** for the full
end-to-end user walkthrough.

---

## Prerequisites

- Go 1.23+
- Node.js + npm
- PostgreSQL
- [`camel`](https://github.com/martin3zra/camel) CLI on your `PATH` (migrations)
- [`air`](https://github.com/air-verse/air) for backend live-reload (optional)

## Environment

Copy the sample and fill in your values:

```shell
cp .env.sample .env
# set DB_NAME, DB_USERNAME, DB_PASSWORD, DB_HOST, DB_PORT, DB_SSLMODE, APP_KEY, ...
```

## Database — start fresh on a new database

Yes — you can spin up a brand-new database and build the whole schema from the
Camel migrations. On a **fresh/empty** DB, `camel migrate` runs the baseline
normally (no marking needed — that's only for an existing prod DB).

```shell
# 1. Create the database (adjust to your setup)
createdb acme           # or: psql -c "CREATE DATABASE acme"

# 2. Point .env at it (DB_NAME=acme, etc.)

# 3. Build the schema. Camel needs a DSN via DB_SOURCE; derive it from .env:
set -a; . ./.env; set +a
export DB_SOURCE="postgres://$DB_USERNAME:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=$DB_SSLMODE"
camel migrate           # creates every table/enum/function/trigger

# (later) check / roll back:
camel status
camel rollback
```

## Create the first account (CLI)

The CLI provisions an owner account and emails an activation invite. Run it
against the migrated database:

```shell
# dev (from source)
go run ./cmd/cli setup:account
# it asks: account name, owner email -> creates the owner user (temp password
# "password", status disabled) + account, and sends the verification email.
```

Other CLI commands:

```shell
go run ./cmd/cli resend:account-email-verification   # resend the invite
go run ./cmd/cli generate:key                        # new APP_KEY value
go run ./cmd/cli version
```

> **There is no CLI command to create a company.** `setup:account` only
> bootstraps the account + owner user. The company itself is created **in the
> web app** by the owner after activating and logging in.

The invited owner then follows **[docs/getting-started.md](docs/getting-started.md)**:
accept the invite → set password → create the company → run the cycle.

## Run the app (development)

```shell
npm install
npm run dev      # Vite dev server (frontend, HMR)
air              # Go backend with live reload  (or: go run .)
```

App serves on `APP_PORT` (default `http://localhost:8092`).

## Tests

DB-backed integration tests run against a **separate** database (never the dev
one). See [docs/getting-started.md](docs/getting-started.md) and the Makefile:

```shell
cp .env.test.sample .env.test   # points at acme_test
make test                       # TestMain creates + migrates acme_test, then runs
```

# Deploy as service
## Test your binary
```shell
cd /Users/amartinez/Developer/acme
./build.sh
chmod +x bin/acme
./acme
```

## Create system service file
Create a file `/etc/systemd/system/acme-service.service`

```
[Unit]
Description=Acme Service
ConditionPathExists=/home/<your-user>/acme
After=network.target

[Service]
Type=simple
User=<your-user>
Group=<your-user>

WorkingDirectory=/home/<your-user>/acme
Environment="YOUR_ENV_1=YOUR_ENV_1_VALUE"
ExecStart=/home/<your-user>/acme/acme
Restart=on-failure
RestartSec=10

StandardOutput=syslog
StandardError=syslog
SyslogIdentifier=acme-service

[Install]
WantedBy=multi-user.target
```

# Restart syslog:
```shell
systemctl restart rsyslog.service
```
# Active service
```
systemctl daemon-reload
service acme-service start
service acme-service status
```