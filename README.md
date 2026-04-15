# Kong Manager OSS

[Installation](#getting-started) | [Docker](#docker) | [Database (BFF)](#database-go-bff) | [Provide feedback](https://github.com/kong/kong-manager/issues/new/choose) | [Ask a question](https://join.slack.com/t/kongcommunity/shared_invite/zt-1s4nb74vo-jLdMEk8MoTm~uMWYMMLPWg) | [Contributing](#contributing) | [Blog](https://konghq.com/blog)

> **Go BFF fork:** This repository adds a Go backend (JWT login, Casbin RBAC, multi-cluster Kong Admin proxy). Roadmap and design notes are in [BACKEND_PLAN.md](BACKEND_PLAN.md). For day-to-day development with login and `/kong-admin` proxying, see [Running locally with Vite and the Go BFF](#running-locally-with-vite-and-the-go-bff).

![Kong Manager OSS - Plugin list](./media/Plugin%20list.png)

Kong Manager OSS, a **free** and **open-source** UI for [Kong](https://github.com/kong/kong), the world's most used open source API Gateway.

Built and maintained with ❤️ by the team at Kong.

## Features

Kong Manager OSS allows you to view and edit all Kong Gateway objects using the admin API. It interacts directly with the Kong admin API and does not require a separate database.

![Kong Manager OSS - Service edit](./media/Service%20edit.png)

> **Important:** Kong Manager OSS is only compatible in Kong Gateway 3.4+

Kong Manager OSS is bundled with Kong Gateway, see [Getting Started](#getting-started) for information on how to use it. To run Kong Manager OSS locally, please see the [contributing](#contributing) section.

## Getting Started

To use Kong Manager OSS you'll need a running Kong Gateway instance. This can be a local instance or running on a remote server.

### Local testing

The quickest way to get started is using the quickstart script:

```bash
curl -Ls https://get.konghq.com/quickstart | bash -s -- -i kong -t latest
```

Finally, visit https://localhost:8002 to view Kong Manager.

### Server usage

Kong Manager OSS is intended to be a local testing tool. However, you can also use it on a public server.

> If running Kong Manager OSS on a public server, ensure that ports `8001` and `8002` are only accessible to your IP address

To access Kong Manager OSS from a remote machine, ensure that `admin_listen` and `admin_gui_listen` are binding to `0.0.0.0` rather than `127.0.0.1` in `kong.conf` and restart your Kong Gateway instance.

## Running locally with Vite and the Go BFF

Use this flow when you need **authentication**, **Casbin RBAC**, and **same-origin Kong Admin** via the BFF (`/kong-admin/c/{cluster}/…`).

### Prerequisites

- **Go** (for `go run ./cmd/kong-manager`)
- **Node.js 18+** and **pnpm**
- A **reachable Kong Admin API** (HTTP or HTTPS). Default listen is often `http://127.0.0.1:8001`; adjust to your environment.

### 1. Frontend config (`kconfig.js`)

The UI reads runtime flags from `/kconfig.js` (`window.K_CONFIG`). For BFF mode:

1. Copy the example file to a local (gitignored) file:

   ```bash
   cp public/kconfig.bff.example.js public/kconfig.js
   ```

2. Typical values (already in the example):

   - `AUTH_REQUIRED: true`
   - `ADMIN_API_URL: '/kong-admin'` (browser calls the BFF; the BFF proxies to Kong using the cluster row in the database)

If `public/kconfig.js` is missing, the Vite dev server may proxy `/kconfig.js` to another host (see `vite.config.ts`), which can fail with **502** if that upstream is not running. Keeping `public/kconfig.js` in place avoids that.

### 2. Start the Go BFF (terminal 1)

Run the BFF on **port `8081`** so it does not collide with Vite on **`8080`**. The Vite dev server proxies `/api` and `/kong-admin` to `http://127.0.0.1:8081` by default (`KONG_MANAGER_BFF_URL` overrides this).

Minimal example (SQLite DB file created next to the process unless you set `DATABASE_URL`):

```bash
KONG_UPSTREAM_TLS_SKIP_VERIFY=true \
HTTP_ADDR=:8081 \
KONG_ADMIN_URL=http://127.0.0.1:8001 \
BOOTSTRAP_ADMIN_USERNAME=admin \
BOOTSTRAP_ADMIN_PASSWORD=changeme \
go run ./cmd/kong-manager
```

- **Database:** By default the BFF uses **SQLite** (`DATABASE_DRIVER=sqlite`, `DATABASE_URL=file:kong-manager.db?…`). For **PostgreSQL** or **MySQL**, set `DATABASE_DRIVER` and `DATABASE_URL` (see [Database (Go BFF)](#database-go-bff)). Schema is applied with GORM `AutoMigrate` on startup.
- **`KONG_ADMIN_URL`**: Used only to **seed** the first `default` Kong cluster row when the `kong_clusters` table is empty. Runtime routing uses each row’s `admin_base_url` (see **Admin → Kong clusters** in the UI, or update the DB). If you already have clusters in the DB, changing this env alone does not retarget an existing row.
- **`BOOTSTRAP_ADMIN_*`**: Creates the first admin user when no users exist.
- **`JWT_SECRET`**: Set a strong secret in any shared or production deployment (default is development-only).

**HTTPS upstream (corporate CA or self-signed):** If Kong Admin uses TLS that Go does not trust (e.g. `x509: certificate signed by unknown authority`), set:

```bash
KONG_UPSTREAM_TLS_SKIP_VERIFY=true
```

Use only in **development** or **trusted networks**. For production, prefer installing your organization’s CA in the OS trust store or using a publicly trusted certificate.

### 3. Start the Vite dev server (terminal 2)

From the repo root:

```bash
pnpm install
pnpm serve:bff
```

- **`serve:bff`** sets `VITE_AUTH_REQUIRED=true` at build time. With `public/kconfig.js` present, `AUTH_REQUIRED` also comes from `window.K_CONFIG`.
- Open the UI at **http://localhost:8080** (Vite default). Use the bootstrap username/password to sign in.

If the BFF listens elsewhere, point Vite at it:

```bash
KONG_MANAGER_BFF_URL=http://127.0.0.1:9090 pnpm serve:bff
```

### 4. Production-style build (optional)

To serve the built SPA from the Go binary (single origin, no Vite):

1. Build the frontend with a root base path and auth enabled, e.g.:

   ```bash
   DISABLE_BASE_PATH=true VITE_AUTH_REQUIRED=true pnpm build
   ```

2. Ensure `public/kconfig.js` exists before build so it is copied into `dist`, or rely on `VITE_AUTH_REQUIRED=true` as documented in `public/kconfig.bff.example.js`.

3. Run the binary with `STATIC_DIR` pointing at `dist` and `HTTP_ADDR` as needed (default `:8080`).

### Troubleshooting

| Symptom | Likely cause | What to do |
|--------|----------------|------------|
| **502** on `/kong-admin/c/default` (response may mention upstream in BFF logs) | BFF cannot connect to Kong Admin (wrong URL/port, Kong down, TLS verify failure) | Confirm Kong Admin URL for cluster `default` (DB or seed). Check BFF logs: `kong-admin proxy: upstream=…`. For TLS trust issues, use `KONG_UPSTREAM_TLS_SKIP_VERIFY=true` only where appropriate, or fix the certificate chain. |
| **401** on `/kong-admin/…` while on the login page | Unauthenticated `getInfo` call before JWT is present | Expected noise in the network tab; sign in and retry. |
| **502** on `/kconfig.js` in dev | Missing `public/kconfig.js` and Vite proxy target unreachable | Add `public/kconfig.js` from the example, or set `KONG_GUI_URL` to a host that serves `kconfig.js`. |
| UI on wrong port | Vite vs BFF port confusion | Vite: **8080**; BFF in README examples: **8081**. Align `KONG_MANAGER_BFF_URL` if you change the BFF port. |

## Database (Go BFF)

The BFF stores users, groups, Casbin policies, Kong cluster rows, SSO providers, notification channels, and audit logs in a relational database. Configure it with:

| Variable | Default (dev) | Values |
|----------|----------------|--------|
| `DATABASE_DRIVER` | `sqlite` | `sqlite` / `sqlite3`, `postgres` / `postgresql`, `mysql` |
| `DATABASE_URL` | `file:kong-manager.db?cache=shared&mode=rwc` | Driver-specific DSN |

**PostgreSQL example:**

```bash
export DATABASE_DRIVER=postgres
export DATABASE_URL="host=127.0.0.1 user=kongmanager password=secret dbname=kong_manager port=5432 sslmode=disable TimeZone=UTC"
```

**PostgreSQL URL form:**

```bash
export DATABASE_URL="postgres://kongmanager:secret@127.0.0.1:5432/kong_manager?sslmode=disable"
```

Create the database (empty) before the first run; migrations run automatically. Switching from SQLite to Postgres does **not** migrate existing data—you need a separate export/import or tooling if you must keep rows.

More context: [BACKEND_PLAN.md](BACKEND_PLAN.md), `internal/config/config.go`, `internal/db/db.go`.

## Docker

A multi-stage **`Dockerfile`** builds the Vite SPA into `dist/` and compiles the Go binary; the runtime image is **distroless** (non-root) with `STATIC_DIR=/app/dist` and `HTTP_ADDR=:8080`.

### Build an image

From the repository root:

```bash
docker build -t kong-manager:latest .
```

### Run with Docker Compose (PostgreSQL)

[`docker-compose.yml`](docker-compose.yml) builds the app and starts **Postgres 16** with a named volume. Copy and edit environment variables first:

```bash
cp docker.env.example .env
# Set at least: POSTGRES_PASSWORD, JWT_SECRET, BOOTSTRAP_ADMIN_PASSWORD
docker compose --env-file .env up --build
```

Then open **http://localhost:8080** (or the host port in `KONG_MANAGER_PORT`).

- **`KONG_ADMIN_URL`** defaults to `http://host.docker.internal:8001` so the BFF can reach Kong Admin on the **host** (Docker Desktop). On **Linux**, `host.docker.internal` may be unavailable unless you add it (e.g. `extra_hosts` or use the host’s LAN IP in `KONG_ADMIN_URL`).
- Treat **`JWT_SECRET`**, database passwords, and bootstrap credentials as secrets in real deployments.

### Run the container alone

Example with SQLite on a volume (Postgres recommended for production):

```bash
docker run --rm -p 8080:8080 \
  -e DATABASE_DRIVER=sqlite \
  -e 'DATABASE_URL=file:/data/kong-manager.db?cache=shared&mode=rwc' \
  -e JWT_SECRET=your-secret \
  -e KONG_ADMIN_URL=http://host.docker.internal:8001 \
  -e BOOTSTRAP_ADMIN_USERNAME=admin \
  -e BOOTSTRAP_ADMIN_PASSWORD=your-pass \
  -v km-data:/data \
  kong-manager:latest
```

## Why do I need this?

You've been using the admin API just fine for years. Why would you want to use a UI?

Kong Manager OSS is a great way to see your Kong Gateway configuration at glance. You can see the routes and plugins configured on a service and drill in to the configuration of each in a single click.

In addition, the plugin configuration UI provides instructions for each configuration option. You can configure a plugin using the UI with helpful tooltips before running `deck dump` to see the final configuration values.

![Kong Manager OSS - Plugin configuration tooltip](./media/Plugin%20configuration%20tooltip.png)

## Contributing

Kong Manager OSS is written in JavaScript. It uses Vue for its UI components, and `pnpm` for managing dependencies. To build Kong Manager OSS locally please ensure that you have `node.js 18+` and `pnpm` installed.

You'll also need a running Kong Gateway instance. See [local testing](#local-testing) for a one-line solution. Alternatively, you can [build Kong Gateway from source](https://github.com/kong/kong/tree/master/build).

**UI only (no BFF, Kong Admin directly):** once Kong Gateway is running with Admin API exposed (often port `8001`), start the dev server:

```bash
pnpm install
pnpm serve
```

Kong Manager OSS is available at http://localhost:8080.

**This fork (UI + Go BFF):** use [Running locally with Vite and the Go BFF](#running-locally-with-vite-and-the-go-bff): `pnpm serve:bff` with the BFF running separately.

Lint:

```bash
pnpm lint
```
