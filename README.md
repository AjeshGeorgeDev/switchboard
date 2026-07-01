# Switchboard

Internal App Launcher & Security Dashboard — one place to launch internal tools and stop pretending Harbor will email anyone.

## Why Switchboard?

You installed **Harbor**, enabled **Trivy**, and posted in `#devops`: *"we're secure now."* The registry looked great. The green checkmarks looked great. You went for coffee. Life was good.

Then Friday, 4:47 PM:

> **Trivy:** critical CVE on `prod/api:latest`  
> **Your phone:** 📵  
> **Your inbox:** 📭  
> **Slack:** 🦗  
> **You:** refreshing Harbor like it's a sports scoreboard

Harbor doesn't email anyone. It *can* fire webhooks — which is Harbor's way of saying *"I told **some** server; not my problem after that."* So now you're knee-deep in endpoint URLs, HMAC secrets, and a Zapier trial you swore you'd cancel, instead of rebuilding the image.

And when someone asks *"what did we ship last week?"* you're not reviewing a timeline. You're an archaeologist, brushing dust off artifacts one tag at a time, praying Harbor still has that scan from Tuesday.

```
Stages of Harbor + Trivy adoption:

[1] We have a registry!          ████████████ 100%
[2] Scans are enabled!           ████████████ 100%
[3] Someone gets notified        ░░░░░░░░░░░░   0%
[4] You write custom glue code   ██████░░░░░░  50%  ← you are here
[5] You find Switchboard         ░░░░░░░░░░░░   0%  ← fix this
```

**Drake says no / Drake says yes:**

| | |
|---|---|
| ❌ Parse Harbor JSON in a Lambda, email Slack, forget it breaks in prod | ✅ Point webhooks at Switchboard, configure SMTP once, touch grass |
| ❌ "I'll check Harbor after standup" | ✅ Critical CVE hits your inbox before standup ends |

Switchboard is the part that actually tells people what happened. Point Harbor and Trivy at it, browse history in **Security → Reports** and **Security → CVEs**, and let **email**, **Teams**, and **in-app** alerts do the nagging (per-user prefs under **Profile**). Keep Harbor as your registry and scanner — Switchboard is where the crew gets the memo.

**Why the name?** Old telephone switchboards patched callers to the right line. This one patches two things: your **app catalog** (click → land in the right internal tool) and your **security feeds** (scan event → land in the right human's inbox). No more alerts lost in webhook limbo. Operator standing by. ☎️

## Screenshots

**Deployment history**

![Deployment reports](docs/screenshots/deployment-reports.png)

**CVE dashboard**

![CVE dashboard](docs/screenshots/cve-dashboard.png)

**Notification preferences**

![Profile notification preferences](docs/screenshots/profile-notifications.png)

Harbor webhooks land in **Reports**; Trivy JSON lands in **CVEs**; alerts go out via **Profile** prefs (email, Teams, in-app).

To refresh screenshots locally (backend + frontend dev servers running):

```bash
make screenshot-data
make screenshots
```

## Stack

- **Backend:** Go (chi, Casbin, sqlc, asynq, robfig/cron)
- **Frontend:** Vue 3 + PrimeVue + Tailwind CSS (`tailwindcss-primeui`) + Pinia
- **Data:** PostgreSQL 17 + Redis

## Quick Start

### Prerequisites

- Docker & Docker Compose (Postgres 17 via `docker-compose.yml`)
- Go 1.25+
- Node.js 20+
- [pnpm](https://pnpm.io/) 10+ (`corepack enable && corepack prepare pnpm@10 --activate`)
- [golang-migrate](https://github.com/golang-migrate/migrate) and [sqlc](https://sqlc.dev/)

```bash
# Install CLI tools
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
go install github.com/air-verse/air@latest
```

### Local development

```bash
# Start Postgres + Redis
make dev-up

# Run migrations
make migrate-up

# Generate sqlc (after schema/query changes)
make sqlc-generate

# Install frontend deps and build
make frontend-install
make frontend-build

# Run the server (from repo root)
make run
```

The API listens on `http://localhost:8080`.

### Development with hot reload

Use two terminals for full-stack dev:

```bash
# Terminal 1 — backend (restarts on .go changes)
make backend-dev

# Terminal 2 — frontend (Vite HMR)
make frontend-dev
```

Open `http://localhost:5173` for the UI (API proxied to `:8080`). Use `make run` instead of `make backend-dev` if you don't need backend hot reload.

### First-run setup

On a fresh install (no administrator account yet), open the app and you'll be redirected to `/setup` to create the superadmin account. This step runs once — after that, login and OIDC flows are enabled.

### Upgrading Postgres 16 → 17 without losing data

Postgres major versions cannot reuse the same on-disk data directory. If you see:

`database files are incompatible with server ... initialized by PostgreSQL version 16`

**Easiest path (fresh PG17, keep old volume):** this project uses a separate Docker volume (`postgres_data_v17`). Just start the stack — Postgres 17 initializes cleanly and leaves the old `switchboard_postgres_data` volume untouched on disk.

```bash
docker compose up -d
make migrate-up
```

You'll go through `/setup` again for a new admin account.

**Migrate old data into PG17** (optional):

```bash
make postgres-upgrade-16-to-17
```

That dumps from the old PG16 volume, then restores into PG17. To back up without upgrading:

```bash
make db-backup
```

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | HTTP port (use `8081` if something else owns `:8080`, e.g. Tomcat) |
| `DATABASE_URL` | `postgres://switchboard:switchboard@localhost:5432/switchboard?sslmode=disable` | Postgres DSN |
| `REDIS_URL` | `redis://localhost:6379/0` | Redis for asynq |
| `JWT_SECRET` | `dev-secret-change-in-production` | JWT signing key |
| `APP_BASE_URL` | `http://localhost:8080` | Base URL for OIDC callbacks |
| `HARBOR_URL` / `HARBOR_TOKEN` | — | Harbor API credentials |
| `TRIVY_URL` / `TRIVY_TOKEN` | — | Trivy API credentials |
| `HARBOR_WEBHOOK_SECRET` / `TRIVY_WEBHOOK_SECRET` | — | Webhook HMAC secrets |
| `CVE_PULL_CRON` | `0 6 * * 0` | Weekly CVE pull schedule |
| `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` | — | Email notifications |

## Webhooks

Switchboard receives security events from Harbor and Trivy, and can send alerts to Microsoft Teams or email — so you can stop hand-rolling the webhook-to-human pipeline.

### Incoming (Harbor & Trivy → Switchboard)

Register these endpoints in Harbor/Trivy (URLs are also shown under **Admin → Configuration**):

| Endpoint | Purpose |
|----------|---------|
| `POST {APP_BASE_URL}/webhooks/harbor` | Deployment reports (Security → Reports) |
| `POST {APP_BASE_URL}/webhooks/trivy` | CVE scan results (Security → CVEs) |

**Local dev:** with Vite (`make frontend-dev`), use `APP_BASE_URL=http://localhost:5173` — the dev proxy forwards `/webhooks` to the backend. Harbor/Trivy on another machine must reach your host IP, not `localhost`.

**Harbor setup**

1. Ensure Switchboard is reachable from Harbor at `APP_BASE_URL` (not `localhost` from a remote Harbor host).
2. In Harbor admin, enable the **Trivy** interrogation service (Configuration → Interrogation Services).
3. Open your project → **Webhooks** → **+ New Webhook**:
   - **Endpoint URL:** `{APP_BASE_URL}/webhooks/harbor`
   - **Events:** `PUSH_ARTIFACT`, `SCANNING_COMPLETED` (and optionally `SCANNING_FAILED`)
4. Test the webhook from Harbor; expect `202` with `{"status":"accepted"}`.
5. Deployment and scan events appear under **Security → Reports**.
6. For production auth, set `HARBOR_WEBHOOK_SECRET` and send `X-Webhook-Signature` (HMAC-SHA256 hex of raw body). Harbor does not send this header natively — use a CI relay or leave the secret empty in dev.

**Trivy setup**

Harbor scan webhooks go to the **Harbor** endpoint. To populate **Security → CVEs** with full Trivy detail, POST standard Trivy JSON from CI:

```bash
# Scan
trivy image --format json --output trivy-report.json myregistry/myapp:v1.0.0

# Send to Switchboard (no auth, local dev)
curl -i -X POST "${APP_BASE_URL}/webhooks/trivy" \
  -H "Content-Type: application/json" \
  --data-binary @trivy-report.json

# With secret (production)
SIG=$(openssl dgst -sha256 -hmac "$TRIVY_WEBHOOK_SECRET" -hex trivy-report.json | awk '{print $2}')
curl -i -X POST "${APP_BASE_URL}/webhooks/trivy" \
  -H "Content-Type: application/json" \
  -H "X-Webhook-Signature: $SIG" \
  --data-binary @trivy-report.json
```

Set `TRIVY_WEBHOOK_SECRET` on the server. Critical CVEs trigger in-app, Teams, and email alerts. Optional scheduled pulls: set `TRIVY_URL`, `TRIVY_TOKEN`, and `CVE_PULL_CRON`.

**Signature header** (when secrets are configured):

```bash
# hex digest of raw body
echo -n '<request-body>' | openssl dgst -sha256 -hmac "$HARBOR_WEBHOOK_SECRET" | awk '{print $2}'
```

Send the result in the `X-Webhook-Signature` request header. Leave secrets empty in local dev to skip verification.

Successful requests return `202` with `{"status":"accepted"}`; processing is asynchronous via Redis/asynq.

### Outgoing (Switchboard → Teams / email)

- **Teams:** create an Incoming Webhook in your Teams channel, then add it under **Admin → Configuration**
- **Email:** set `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, and `SMTP_FROM` on the server

## Docker

```bash
docker build -t switchboard .
docker run -p 8080:8080 --env-file .env switchboard
```

## Architecture

See [architecture-plan.md](architecture-plan.md) for the full design document.

**Note:** The launcher opens catalog targets directly in the user's browser (client-side redirect). The user's machine must have network access to internal IP:port targets — the server does not proxy.
