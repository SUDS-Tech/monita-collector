# monita-collector

The server-side collector for [Monita](https://github.com/kasimlyee/monita-agent) — an open-source infrastructure monitoring platform. Agents running on your hosts ship metrics and logs here; users query, stream, and alert on them through the REST API.

Built with Go, PostgreSQL 15+, and the [Bast](https://github.com/bastion-framework/bast) web framework. Licensed under **AGPL v3**.

---

## How it fits together

```
[monita-agent]  ──HMAC-signed HTTP──►  [monita-collector]  ──►  PostgreSQL
  (your host)                             (this repo)
                                               │
                                    JWT-authed REST API
                                               │
                                         Dashboard / CLI
```

[**monita-agent**](https://github.com/kasimlyee/monita-agent) is a lightweight Go binary (~10 MB, <20 MB RSS idle) that runs as a systemd service or Docker sidecar on monitored hosts. It collects CPU, memory, disk, network, and load metrics, tails log files and Docker container outputs, and ships signed batches to this collector. Agents authenticate with a bearer token + HMAC-SHA256 signature. Users authenticate with a JWT. The collector stores everything in PostgreSQL and streams live events over SSE.

---

## Features

- **Metric ingestion** — batch time-series points per agent, query by name/window/labels
- **Log ingestion** — structured log entries with full-text search (PostgreSQL `tsquery`)
- **Alert rules** — define conditions on metrics or logs; fire and resolve events
- **Live streaming** — Server-Sent Events for real-time metrics and log tails
- **Agent auth** — HMAC-SHA256 per-request signatures, nonce replay protection, fingerprint drift detection, token expiry
- **OpenAPI docs** — auto-generated at `/docs`
- **Rate limiting** — per-IP token bucket (100 req/s, burst 200)
- **Input validation** — field-level 422 responses on bad requests
- **Graceful shutdown** — drains in-flight requests on SIGTERM/SIGINT

---

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- [`golang-migrate`](https://github.com/golang-migrate/migrate) CLI for running migrations
- [`sqlc`](https://sqlc.dev) v1.28+ if regenerating query code

---

## Quick start

```bash
git clone https://github.com/SUDS-Tech/monita-collector
cd monita-collector

# copy and fill in env vars
cp .env.example .env

# run migrations
migrate -path internal/db/migrations -database "$DB_URL" up

# start the server
go run .
```

Server listens on `:8080` by default. OpenAPI docs at `http://localhost:8080/docs`.

---

## Configuration

All configuration is read from environment variables (or a `.env` file in the working directory).

| Variable | Required | Default | Description |
|---|---|---|---|
| `PORT` | | `8080` | HTTP listen port |
| `APP_ENV` | | `development` | `development` or `production` |
| `DB_URL` | ✓ | — | PostgreSQL connection string |
| `DB_MAX_CONNS` | | `25` | pgxpool max connections |
| `JWT_SECRET` | ✓ | — | Secret for signing user JWTs |
| `STRIPE_SECRET_KEY` | | — | Stripe secret (billing, optional) |
| `STRIPE_WEBHOOK_SECRET` | | — | Stripe webhook verification |
| `STRIPE_PRICE_ID` | | — | Stripe price for paid plans |

---

## API reference

Base URL: `http://localhost:8080`

### Auth

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/register` | — | Create user account |
| `POST` | `/login` | — | Get JWT |
| `GET` | `/me` | Session | Current user |

### Agents

All routes require a valid user session (`Authorization: Bearer <jwt>`).

| Method | Path | Description |
|---|---|---|
| `POST` | `/agents` | Register a new agent; returns token + signing key (shown once) |
| `GET` | `/agents` | List agents for your org |
| `GET` | `/agents/:id` | Get agent detail |
| `POST` | `/agents/:id/revoke` | Revoke agent token |
| `DELETE` | `/agents/:id` | Delete agent |

### Metrics

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/metrics/ingest` | Agent | Batch ingest metric points |
| `GET` | `/metrics` | Session | Query points (`agent_id`, `metric_name`, `from`, `to`, `limit`) |
| `GET` | `/metrics/names` | Session | List distinct metric names for an agent |

### Logs

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/logs/ingest` | Agent | Batch ingest log entries |
| `GET` | `/logs` | Session | Query logs (`agent_id`, `from`, `to`, `level`, `limit`) |
| `GET` | `/logs/search` | Session | Full-text search (`agent_id`, `q`, `from`, `to`, `limit`) |

### Alerts

| Method | Path | Auth | Description |
|---|---|---|---|
| `POST` | `/alerts/rules` | Session | Create alert rule |
| `GET` | `/alerts/rules` | Session | List rules for your org |
| `GET` | `/alerts/rules/:id` | Session | Get rule |
| `PUT` | `/alerts/rules/:id` | Session | Update rule |
| `DELETE` | `/alerts/rules/:id` | Session | Delete rule |
| `GET` | `/alerts/events` | Session | List firing events for your org |
| `GET` | `/alerts/rules/:id/events` | Session | List events for a specific rule |
| `POST` | `/alerts/events/:id/resolve` | Session | Resolve an alert event |

### Streaming (SSE)

| Path | Auth | Description |
|---|---|---|
| `/stream/metrics` | Session | Live metric events for your agents |
| `/stream/logs` | Session | Live log tail for your agents |

### System

| Path | Description |
|---|---|
| `/health` | Liveness probe |
| `/ready` | Readiness probe (checks PostgreSQL) |
| `/docs` | OpenAPI UI |
| `/openapi.json` | OpenAPI JSON spec |

---

## Agent authentication

Agents use a two-factor scheme: a **bearer token** (for identity) plus a **per-request HMAC-SHA256 signature** (for integrity and replay protection).

```
Authorization: Bearer <token>
X-Timestamp: <unix seconds>
X-Nonce: <random string, used once>

HMAC-SHA256(signing_key, "<timestamp>\n<nonce>\n<fingerprint>\n<sha256(body)>")
```

- The signing key is returned once at agent creation and never stored in plaintext on the server.
- Requests outside ±120 s of server time are rejected (clock skew protection).
- Nonces are tracked per agent and cannot be reused within the replay window.
- The client fingerprint is recorded on first use; drift triggers a freeze.

To configure an agent to point at this collector, set its `collector_url`, `agent_id`, `token`, and `signing_key` (base64url-encoded HMAC key returned at agent creation). See [monita-agent](https://github.com/kasimlyee/monita-agent) for the full wire protocol and deployment guide.

---

## Project layout

```
.
├── main.go                    # wiring: app, modules, signal handling
├── internal/
│   ├── config/                # env var loading
│   └── db/
│       ├── migrations/        # golang-migrate SQL files
│       ├── queries/           # sqlc SQL query files
│       └── sqlc/              # generated Go query code
├── modules/
│   ├── users/                 # registration, login, JWT
│   ├── agents/                # agent lifecycle + token management
│   ├── metrics/               # metric ingest + query
│   ├── logs/                  # log ingest + query + search
│   ├── alerts/                # alert rules + events
│   └── stream/                # SSE hub + streaming endpoints
└── shared/
    ├── guards/                # session auth, agent HMAC auth, nonce cache
    ├── middleware/            # rate limiter
    ├── validate/              # go-playground/validator wrapper
    └── errors/                # error code constants
```

---

## Development

```bash
# generate sqlc query code after editing SQL
sqlc generate

# create a new migration
migrate create -ext sql -dir internal/db/migrations -seq <name>

# build
go build .

# vet
go vet ./...
```

---

## License

[GNU Affero General Public License v3.0](LICENSE) — SUDS-Tech
