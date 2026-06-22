# Bridgehead

A Go-based IoT downlink service that manages and delivers requests to physical border gateway devices. Handles intermittently online gateways through an event-driven dispatcher and robust state machine.

---

## Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [Project Structure](#project-structure)
- [Domain Model](#domain-model)
- [State Machine](#state-machine)
- [Tech Stack](#tech-stack)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Database](#database)
- [API](#api)
- [Testing](#testing)
- [Deployment](#deployment)
- [Infrastructure](#infrastructure)
- [Contributing](#contributing)

---

## Overview

Border gateways (BG) are physical IoT devices that are offline most of the time and come online for short windows. During these windows, the gateway accepts downlink requests — configuration, commands, firmware updates, or acknowledgements — for itself or for sensors registered to it.

Bridgehead solves the reliable delivery problem:

- Accepts downlink requests via REST API
- Tracks gateway liveness via MQTT uplinks
- Maps sensors to their registered gateway via Kinesis streams
- Syncs site/gateway/sensor topology via SQS
- Queues requests and dispatches them the moment a gateway comes online
- Retries failed requests with exponential backoff
- Expires requests that exceed their TTL

```
User / Service
      │
      ▼ REST API
┌─────────────────────────────────┐
│          Bridgehead              │
│                                 │
│  Downlink CRUD  State Machine   │
│  Gateway Liveness  Scheduler   │
│  Sensor Mapping  Topology      │
└────────┬──────────┬─────────────┘
         │          │
         ▼          ▼
      Single     Border Gateway
        DB       (MQTT publish)
         ▲
    ┌────┴────┬──────────┐
    │         │          │
   MQTT    Kinesis      SQS
  uplinks   sensor    topology
            data      updates
```

---

## Architecture

Bridgehead is a **Modular Monolith** following **Ports and Adapters** (Hexagonal Architecture).

### Why Modular Monolith

- Integrations (MQTT, Kinesis, SQS) are fixed and known — not runtime-swappable
- Single database with shared state makes microservices harmful, not helpful
- The dispatcher needs tight coupling between gateway liveness and request state
- Go's goroutine model handles all concurrency needs in a single binary

### Why Not Microkernel

Microkernel assumes plugins are unknown at design time and independent of each other. Every integration here has a fixed role and they depend on each other — MQTT feeds the dispatcher which reads downlink requests created by the REST API. That's a pipeline, not a plugin system.

### Ports and Adapters

```
Domain (business logic)          Infrastructure (adapters)
────────────────────────         ─────────────────────────
internal/downlink/               infra/mqtt/       ← MQTT client
internal/gateway/                infra/kinesis/    ← Kinesis consumer
internal/sensor/                 infra/sqs/        ← SQS poller
internal/topology/               infra/db/         ← PostgreSQL pool
internal/scheduler/              infra/http/       ← Gin HTTP server
                                 infra/cache/      ← Redis client
```

Domain modules never import infrastructure. Infrastructure adapters call domain services. `cmd/server/main.go` is the sole composition root — the only file that knows about everything.

### Cross-Domain Communication

Domains never call each other's repositories. They communicate through interfaces defined by the consumer and implemented by the provider:

```go
// sensor/service.go defines what it needs
type GatewayLookup interface {
    IsOnline(ctx context.Context, gatewayID string) bool
}

// gateway/service.go implements it — without knowing sensor exists
func (s *Service) IsOnline(ctx context.Context, gatewayID string) bool { ... }

// main.go wires them together — only place that knows both exist
sensorSvc := sensor.NewService(sensorRepo, gatewayService)
```

---

## Project Structure

```
bridgehead/
├── app/                              # All Go application code
│   ├── cmd/
│   │   ├── server/
│   │   │   └── main.go              # Entry point — composition root
│   │   └── migrate/
│   │       └── main.go              # Standalone migration binary
│   │
│   ├── internal/                    # Domain modules
│   │   ├── downlink/
│   │   │   ├── model.go             # Domain model — no DB/JSON tags
│   │   │   ├── errors.go            # Domain errors
│   │   │   ├── state.go             # State machine transitions
│   │   │   ├── service.go           # Business logic
│   │   │   ├── service_test.go      # Unit tests (mocked deps)
│   │   │   ├── repository.go        # DB queries (sqlx)
│   │   │   ├── repository_test.go   # Repo tests (real DB)
│   │   │   └── integration_test.go  # Integration tests
│   │   ├── gateway/                 # Gateway liveness tracking
│   │   ├── sensor/                  # Sensor-BG mapping
│   │   ├── topology/                # Site/gateway/sensor relations
│   │   └── scheduler/               # Dispatcher — retry engine
│   │
│   ├── infra/                       # Infrastructure adapters
│   │   ├── db/
│   │   │   └── postgres.go          # sqlx connection pool
│   │   ├── cache/
│   │   │   └── redis.go             # Redis client
│   │   ├── mqtt/
│   │   │   └── consumer.go          # MQTT subscriber
│   │   ├── kinesis/
│   │   │   └── consumer.go          # Kinesis shard consumer
│   │   ├── sqs/
│   │   │   └── consumer.go          # SQS long poller
│   │   └── http/
│   │       ├── server/
│   │       │   └── server.go        # Gin engine setup
│   │       ├── handlers/
│   │       │   ├── dependencies.go  # Handler dependency struct
│   │       │   ├── downlink.go      # Downlink HTTP handlers
│   │       │   ├── health.go        # Health check handler
│   │       │   └── errors.go        # Domain error → HTTP status mapping
│   │       ├── routes/
│   │       │   ├── routes.go        # Route registration
│   │       │   ├── downlink.go      # Downlink routes
│   │       │   └── health.go        # Health routes
│   │       ├── middleware/
│   │       │   ├── logger.go        # Request logging middleware
│   │       │   └── auth.go          # JWT validation middleware
│   │       ├── dto/
│   │       │   └── downlink.go      # Request/response structs (json tags)
│   │       └── swagger/
│   │           └── swagger.go       # Swagger UI registration
│   │
│   ├── config/
│   │   └── config.go                # Env var loading — mustEnv panics fast
│   │
│   ├── pkg/
│   │   └── authctx/
│   │       └── context.go           # Shared auth context helpers
│   │
│   ├── migrations/
│   │   └── versioned/               # Hand-written SQL, never edited after apply
│   │       ├── 20240601000001_init_downlink_requests.sql
│   │       └── 20240601000002_init_gateway.sql
│   │
│   ├── scripts/
│   │   ├── setup_db.sh              # DB setup script runner
│   │   └── postgres_setup.sql       # Least-privilege DB setup
│   │
│   ├── go.mod
│   └── Makefile                     # Delegated from root Makefile
│
├── terraform/
│   ├── environments/
│   │   ├── dev/
│   │   ├── staging/
│   │   └── prod/
│   └── modules/
│       ├── ecs/                     # ECS service definition
│       ├── rds/                     # PostgreSQL RDS
│       ├── sqs/                     # SQS queue
│       ├── kinesis/                 # Kinesis stream
│       └── vpc/                     # Networking
│
├── .github/
│   └── workflows/
│       ├── ci.yml                   # Test + lint on PR
│       ├── cd-staging.yml           # Deploy on merge to main
│       └── cd-prod.yml              # Deploy on version tag
│
├── .envrc.example
├── .gitignore
├── Makefile                         # Root — delegates to app/ and terraform/
└── README.md
```

---

## Domain Model

### Downlink Request

| Field         | Type      | Description                                         |
| ------------- | --------- | --------------------------------------------------- |
| `id`          | UUID      | Auto-generated or caller-provided (idempotency key) |
| `device_eui`  | string    | Unique EUI of target gateway or sensor              |
| `device_type` | enum      | `gateway` or `sensor`                               |
| `payload`     | bytes     | Binary payload — base64 in API, raw in DB           |
| `type`        | enum      | `config`, `command`, `firmware`, `ack`              |
| `status`      | enum      | See state machine below                             |
| `retry_count` | int       | Number of dispatch attempts                         |
| `expires_at`  | timestamp | TTL — defaults to 24h from creation                 |

### Device Types

A downlink targets either a **gateway** or a **sensor** by their EUI — never both simultaneously. Sensors can only receive downlinks through their registered border gateway.

---

## State Machine

```
                  BG offline
  ┌────────────────────────────────────────┐
  │                                        │
CREATE      BG online     sent to BG    BG ack
PENDING ──▶ QUEUED ──────▶ DISPATCHED ──▶ DELIVERED
              │                │
              │ TTL exceeded    │ max retries hit
              ▼                ▼
           EXPIRED           FAILED
```

| Transition               | Trigger                                       |
| ------------------------ | --------------------------------------------- |
| `PENDING → QUEUED`       | Gateway comes online (MQTT uplink transition) |
| `QUEUED → DISPATCHED`    | Dispatcher sends request to BG                |
| `DISPATCHED → DELIVERED` | BG acknowledges receipt                       |
| `DISPATCHED → QUEUED`    | BG went offline mid-flush — re-queue          |
| `DISPATCHED → FAILED`    | Max retries (`5`) exceeded                    |
| `QUEUED → EXPIRED`       | Request TTL exceeded before dispatch          |

### Dispatcher Logic

The dispatcher is **event-driven**, not polling-based. It wakes on `OFFLINE → ONLINE` gateway transitions detected by the MQTT liveness watcher — not on a timer. A 60-second safety poller runs as fallback for requests missed during service restarts.

```
MQTT uplink received
       │
gateway.Service.RecordUplink()
       │
  ┌────▼────┐
  │ already │ yes → update last_seen, no event
  │ online? │
  └────┬────┘
       │ no (transition)
       ▼
fire GatewayOnlineEvent → channel
       │
scheduler.Dispatcher (goroutine)
       │
  re-verify liveness ──▶ offline → return
       │ still online
       ▼
fetch QUEUED requests for gateway's sensors
       │
  for each request:
    re-check liveness → offline → break
    send to BG
    on success → DISPATCHED
    on failure → QUEUED (retry) or FAILED (max retries)
```

---

## Tech Stack

| Concern    | Choice            | Reason                                                                     |
| ---------- | ----------------- | -------------------------------------------------------------------------- |
| Language   | Go 1.22           | Native goroutines for concurrent consumers, single binary, IoT suitability |
| HTTP       | Gin               | Fast router, middleware support                                            |
| Database   | PostgreSQL + sqlx | `FOR UPDATE SKIP LOCKED` for dispatcher, no ORM abstraction overhead       |
| Migrations | Atlas (versioned) | Hand-written SQL, audit trail, no ORM dependency                           |
| Cache      | Redis             | Gateway liveness cache — IsOnline() cache-first                            |
| Logging    | zap               | Structured, zero-allocation in production                                  |
| MQTT       | paho.mqtt.golang  | Eclipse Paho — industry standard                                           |
| Kinesis    | AWS SDK v2        | Shard iterator, checkpointing                                              |
| SQS        | AWS SDK v2        | Long polling, visibility timeout                                           |
| API Docs   | swaggo/swag       | Swagger UI from annotations                                                |
| Infra      | Terraform         | ECS, RDS, SQS, Kinesis, VPC                                                |
| CI/CD      | GitHub Actions    | PR checks, staging/prod deploy                                             |

---

## Prerequisites

- Go 1.22+
- PostgreSQL 15+ (Homebrew: `brew install postgresql@18`)
- Redis (`brew install redis`)
- Atlas CLI (`curl -sSf https://atlasgo.sh | sh`)
- swag CLI (`go install github.com/swaggo/swag/cmd/swag@latest`)
- direnv (`brew install direnv`)
- AWS CLI (for Kinesis/SQS local development via LocalStack)
- Docker (for LocalStack and testcontainers)

---

## Getting Started

### 1. Clone and configure

```bash
git clone https://github.com/anoop-dryad/bridgehead.git
cd bridgehead

# copy env config
cp .envrc.example .envrc

# edit .envrc with your local values
# then allow direnv to load it
direnv allow
```

### 2. Set up the database

```bash
# start postgres (Homebrew)
brew services start postgresql@18

# create DB and app user (uses .envrc values)
bash app/scripts/setup_db.sh
```

### 3. Run migrations

```bash
make migrate
```

### 4. Start the server

```bash
make run
```

Server starts on `http://localhost:8080`

### 5. Open Swagger UI

```
http://localhost:8080/swagger/index.html
```

---

## Configuration

All configuration is via environment variables. Copy `.envrc.example` to `.envrc` and fill in values.

| Variable                | Required | Default          | Description                                   |
| ----------------------- | -------- | ---------------- | --------------------------------------------- |
| `ENV`                   | No       | `development`    | Set to `production` to enable prod mode       |
| `APP_NAME`              | No       | `bridgehead`     | Application name                              |
| `APP_VERSION`           | No       | `dev`            | Application version                           |
| `HTTP_ADDR`             | No       | `:8080`          | HTTP server address                           |
| `DB_DSN`                | **Yes**  | —                | PostgreSQL DSN                                |
| `REDIS_ADDR`            | No       | `localhost:6379` | Redis address                                 |
| `REDIS_PASSWORD`        | No       | ``               | Redis password                                |
| `MQTT_BROKER_URL`       | **Yes**  | —                | MQTT broker URL (e.g. `tcp://localhost:1883`) |
| `MQTT_CLIENT_ID`        | No       | `bridgehead`     | MQTT client ID                                |
| `MQTT_TOPIC`            | **Yes**  | —                | MQTT topic pattern (e.g. `gateways/#`)        |
| `KINESIS_STREAM_NAME`   | **Yes**  | —                | Kinesis stream name                           |
| `SQS_QUEUE_URL`         | **Yes**  | —                | SQS queue URL                                 |
| `AWS_REGION`            | No       | `eu-central-1`   | AWS region                                    |
| `AWS_ACCESS_KEY_ID`     | **Yes**  | —                | AWS credentials                               |
| `AWS_SECRET_ACCESS_KEY` | **Yes**  | —                | AWS credentials                               |

**Required variables cause an immediate panic at startup if missing** — fail fast, fail loudly.

In production, values are sourced from AWS Secrets Manager. Never commit `.envrc`.

---

## Database

### Setup

```bash
# run setup script (creates DB, user, schema with least-privilege access)
bash app/scripts/setup_db.sh
```

The setup script creates:

- Database: `bridgehead`
- User: `bridgehead_user` (no superuser, no createdb)
- Grants only what the app needs — connect, read, write

### Migrations

Migrations use **Atlas in versioned mode**. Hand-written SQL files, never edited after they are applied.

```bash
# apply all pending migrations
make migrate

# check migration status
atlas migrate status --env local

# create a new migration
touch app/migrations/versioned/$(date +%Y%m%d%H%M%S)_describe_change.sql
# write your ALTER TABLE / CREATE TABLE, then:
make migrate
```

**Never run migrations on server startup.** Migrations run as a separate binary (`cmd/migrate`) before the server starts — two instances racing on migrations corrupt schema state.

**Never edit an applied migration file.** Atlas checksums each file. Create a new file for every change.

### Schema overview

```sql
downlink_requests   -- core table: requests, state, payload, device targeting
gateways            -- gateway registry, liveness tracking
sensors             -- sensor registry, gateway mapping
sites               -- site topology (from SQS)
atlas_schema_revisions  -- managed by Atlas, tracks applied migrations
```

---

## API

Base URL: `http://localhost:8080/v1`

### Downlink Requests

| Method   | Path             | Description                 |
| -------- | ---------------- | --------------------------- |
| `POST`   | `/downlinks`     | Create a downlink request   |
| `GET`    | `/downlinks`     | List requests by device EUI |
| `GET`    | `/downlinks/:id` | Get a request by ID         |
| `DELETE` | `/downlinks/:id` | Delete a pending request    |

### Create Downlink Request

```bash
POST /v1/downlinks
Content-Type: application/json

{
  "id": "optional-idempotency-key",        # optional — caller-provided UUID
  "device_eui": "AA:BB:CC:DD:EE:FF:00:01",
  "device_type": "sensor",                  # gateway | sensor
  "payload": "base64encodedpayload==",      # binary payload, base64 encoded
  "type": "config",                         # config | command | firmware | ack
  "expires_at": "2024-12-31T23:59:59Z"      # optional — defaults to now + 24h
}
```

```bash
# Response 201
{
  "id": "550e8400-e29b-41d4-a716-446655440000",
  "device_eui": "AA:BB:CC:DD:EE:FF:00:01",
  "device_type": "sensor",
  "payload": "base64encodedpayload==",
  "type": "config",
  "status": "pending",
  "retry_count": 0,
  "created_at": "2024-06-01T10:00:00Z",
  "expires_at": "2024-06-02T10:00:00Z"
}
```

**Idempotency:** providing the same `id` twice returns `409 Conflict`. Safe to retry with the same ID — only one request is created.

**Note:** `PUT` is intentionally absent. A downlink's payload, device, and type must not change after creation. Only the dispatcher changes status.

### Health Check

```bash
GET /ping
# Response 200: { "message": "pong" }
```

### Swagger UI

```
http://localhost:8080/swagger/index.html
```

Only available when `ENV != production`.

---

## Testing

### Unit tests (fast — no external deps)

```bash
make test
```

All dependencies mocked via interfaces. Tests run in milliseconds. Run on every save.

### Repository tests (real DB via testcontainers)

```bash
make test
# repository_test.go files use testcontainers automatically
# spins up a postgres container, runs migrations, tears down after
```

### Integration tests (full domain wired)

```bash
make test-integration
```

Uses `//go:build integration` tag — not run by default, only in CI.

### Vulnerability scan

```bash
make vuln
# uses govulncheck — Google's vuln DB, no NVD rate limits
```

### Test structure

Tests live **next to the code they test** — no separate `tests/` folder. This is Go convention:

```
internal/downlink/
├── service.go
├── service_test.go        # unit — white box, mocked repo
├── repository.go
├── repository_test.go     # repo — real postgres via testcontainers
└── integration_test.go    # integration — full domain, build tag required
```

---

## Deployment

### Docker

```bash
# build server image
make docker-build

# images
deploy/docker/Dockerfile         # server binary
deploy/docker/Dockerfile.migrate # migration runner (includes Atlas CLI)
```

### CI/CD (GitHub Actions)

| Workflow         | Trigger                | What it does                                              |
| ---------------- | ---------------------- | --------------------------------------------------------- |
| `ci.yml`         | Pull request to `main` | lint, unit tests, integration tests, build                |
| `cd-staging.yml` | Merge to `main`        | build image → push ECR → run migrations → deploy ECS      |
| `cd-prod.yml`    | Push tag `v*`          | same as staging → prod cluster (requires manual approval) |

### Release process

```bash
# tag a release — triggers prod deploy
git tag v1.2.0
git push origin v1.2.0
```

Production deploys require manual approval in GitHub Environments.

---

## Infrastructure

Terraform manages all AWS resources. Each environment is isolated.

```
terraform/
├── environments/
│   ├── dev/        # development — smallest instance sizes
│   ├── staging/    # staging — mirrors prod config
│   └── prod/       # production
└── modules/
    ├── ecs/        # ECS Fargate service + task definition
    ├── rds/        # PostgreSQL RDS (multi-AZ in prod)
    ├── sqs/        # SQS queue for topology updates
    ├── kinesis/    # Kinesis stream for sensor uplinks
    └── vpc/        # VPC, subnets, security groups
```

### Terraform commands

```bash
# initialise
make tf-init ENV=staging

# plan — always review before apply
make tf-plan ENV=staging

# apply
make tf-apply ENV=staging
```

---

## Makefile Reference

```bash
make run              # generate swagger docs + start server
make swagger          # regenerate swagger docs only
make migrate          # run pending DB migrations
make test             # unit + repository tests
make test-integration # all tests including integration
make lint             # golangci-lint
make vuln             # govulncheck vulnerability scan
make build            # compile binary
make docker-build     # build Docker image
make tidy             # go mod tidy
make deps             # go mod download
make tf-init ENV=x    # terraform init for environment
make tf-plan ENV=x    # terraform plan for environment
make tf-apply ENV=x   # terraform apply for environment
```

---

## Contributing

### Branch naming

```
feature/description
fix/description
chore/description
```

### Commit style

```
feat: add firmware downlink type
fix: prevent double dispatch on rapid MQTT reconnect
chore: update go dependencies
```

### Pull request checklist

- [ ] Tests added or updated
- [ ] `make lint` passes
- [ ] `make test` passes
- [ ] New migration file created if schema changed (never edit existing)
- [ ] Swagger regenerated if API changed (`make swagger`)
- [ ] `.envrc.example` updated if new env vars added

### Adding a new domain module

1. Create `internal/<domain>/` with `model.go`, `errors.go`, `service.go`, `repository.go`
2. Define interfaces for any cross-domain dependencies in the consuming domain
3. Add infra adapter in `infra/<transport>/` if a new external integration is needed
4. Wire in `cmd/server/main.go` only
5. Add handler + routes + DTO if HTTP exposure is needed
6. Write migration for any new tables

### Key conventions

- Constants and errors live inside their owning domain — not in shared packages
- Repositories are private to their domain — no cross-domain repo access
- Logging only in services — not in repositories or handlers
- Use `Debug` for success paths (visible in dev), `Error` for failures (visible everywhere)
- `main.go` is the only composition root — the only file that imports everything
- Never run migrations on server startup

---

## Dependency Updates

Dependabot monitors dependencies weekly:

- Go modules (`app/go.mod`)
- Terraform providers and modules (per environment)
- GitHub Actions versions

PRs are auto-created. Review and merge — never auto-merge dependency updates to production.
