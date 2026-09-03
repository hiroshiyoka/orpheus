# Orpheus

Orpheus is a unified monitoring dashboard for multiple web projects. It brings uptime monitoring, error tracking, performance metrics, incident history, and Telegram alerting into one dashboard.

The name comes from the Greek myth of Orpheus, who descended into the underworld to bring back the dead. The monitoring system follows the same metaphor: it detects when a service dies and reports when it comes back to life.

## Why Orpheus?

Managing several deployed projects often means checking multiple Cloudflare dashboards and scattered logs. Orpheus centralizes the information needed to answer a few essential questions:

- Which services are currently available?
- How quickly are they responding?
- Has a service recently experienced downtime?
- Is an error rate increasing?
- When did an incident start, and has it been resolved?

The system retains local historical data so service health can be reviewed beyond the default retention period of the provider dashboards.

## Features

- Active HTTP health checks for every monitored project
- Response status and response-time tracking
- Historical uptime and performance data
- Cloudflare Analytics metrics for configured zones
- Incident detection for downtime and error spikes
- Recovery detection when a service becomes available again
- Telegram notifications for incidents and recoveries
- SQLite storage with no external database service required
- REST API for dashboards and external integrations
- React dashboard using Cloudflare Kumo UI
- Container deployment with Podman

## Architecture

```text
┌──────────────────┐     ┌──────────────────┐     ┌─────────────────┐
│ Cloudflare        │────▶│  Go Collector    │────▶│   SQLite        │
│ Analytics API     │     │  (goroutines)    │     │                 │
└──────────────────┘     └──────────────────┘     └─────────────────┘
┌──────────────────┐              │                        │
│ Active Ping      │──────────────┘                        ▼
│ (HTTP checker)   │                              ┌─────────────────┐
└──────────────────┘                              │  Go API Server  │
                                                    └────────┬────────┘
                                          ┌─────────────────┼─────────────────┐
                                          ▼                 ▼                 ▼
                                  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
                                  │ React +      │  │  Telegram    │  │   REST API   │
                                  │ Kumo UI      │  │  Bot API     │  │  (external)  │
                                  └──────────────┘  └──────────────┘  └──────────────┘
```

### Components

| Component | Responsibility | Technology |
|---|---|---|
| Collector | Runs health checks and retrieves Cloudflare Analytics data | Go, goroutines, `time.Ticker` |
| Storage | Persists checks, metrics, incidents, and sent alerts | SQLite, `database/sql` |
| Detector | Identifies downtime, error spikes, and recoveries | Go |
| Alerting | Sends incident notifications | Telegram Bot API |
| API server | Exposes monitoring data and project configuration endpoints | Go, `net/http` |
| Dashboard | Displays project status, charts, and incident history | React, TypeScript, Kumo UI |

## Monitoring Model

Orpheus uses two complementary data collection schedules:

- **Health checks** run at short intervals, such as every 60 seconds. They record whether a project is reachable, its HTTP status code, response time, and any connection error.
- **Analytics collection** runs less frequently, such as hourly. It stores aggregated Cloudflare data including request count, error count, and response-time percentiles.

An incident is created when a project fails a configured number of consecutive health checks. A successful check resolves an active downtime incident. Cloudflare error-rate thresholds can also create error-spike incidents.

## Project Structure

```text
orpheus/
├── cmd/
│   └── server/
│       └── main.go
├── internal/
│   ├── collector/
│   │   ├── pinger.go
│   │   └── cloudflare.go
│   ├── storage/
│   │   ├── db.go
│   │   ├── projects.go
│   │   ├── checks.go
│   │   ├── metrics.go
│   │   └── incidents.go
│   ├── alerting/
│   │   └── telegram.go
│   ├── detector/
│   │   └── detector.go
│   └── api/
│       ├── router.go
│       ├── handlers_projects.go
│       ├── handlers_status.go
│       └── handlers_incidents.go
├── migrations/
│   └── 0001_init.sql
├── dashboard/
├── go.mod
└── go.sum
```

## Data Storage

The SQLite database contains five primary tables:

- `projects`: monitored project registry and check configuration
- `checks`: individual health-check results
- `metrics`: aggregated Cloudflare Analytics data
- `incidents`: detected downtime and metric-related incidents
- `alerts_sent`: record of notifications already sent

SQLite keeps the deployment simple while preserving structured historical data for uptime calculations, charts, and incident reports.

## API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/projects` | Lists projects with their latest status |
| `GET` | `/api/projects/:id/checks?range=24h` | Returns health-check history |
| `GET` | `/api/projects/:id/metrics?range=7d` | Returns aggregated metrics |
| `GET` | `/api/incidents` | Lists active and resolved incidents |
| `GET` | `/api/incidents/:id` | Returns incident details |
| `POST` | `/api/projects` | Adds a project to monitoring |
| `PATCH` | `/api/projects/:id` | Updates project configuration |

Example project response:

```json
[
  {
    "id": 1,
    "name": "Example Project",
    "url": "https://example.com",
    "is_up": true,
    "last_checked_at": "2026-09-02T10:30:00Z",
    "response_time_ms": 142,
    "uptime_24h_percent": 99.8
  }
]
```

## Configuration

Copy `.env.example` to `.env` and set the required values:

```dotenv
TELEGRAM_BOT_TOKEN=
TELEGRAM_CHAT_ID=
CLOUDFLARE_API_TOKEN=
DB_PATH=./data/orpheus.db
CHECK_INTERVAL_SECONDS=60
```

Telegram credentials are used for incident and recovery notifications. The Cloudflare API token is used to retrieve Analytics data for projects with a configured zone ID.

## Running Locally

Requirements:

- Go 1.23 or later
- Node.js 20 or later for the dashboard
- npm

Run the backend:

```bash
go mod download
go run ./cmd/server
```

The backend opens the configured SQLite database and applies the migration on startup.

Create and run the dashboard:

```bash
npm create vite@latest dashboard -- --template react-ts
cd dashboard
npm install @cloudflare/kumo react react-dom @phosphor-icons/react
npm run dev
```

## Telegram Alerts

Orpheus sends short, actionable messages when an incident starts or recovers. A downtime alert is sent after three consecutive failures by default. Error-spike alerts are sent when the configured error-rate threshold is exceeded.

Example:

```text
🔴 [Example Project] Down since 10:32
URL: https://example.com
Status: Connection timeout
```

## Container Deployment

The backend uses `modernc.org/sqlite`, a pure-Go SQLite driver, so the container does not require CGO or `libsqlite3`.

Build the images:

```bash
podman build -t localhost/orpheus-backend:latest -f Containerfile.backend .
podman build -t localhost/orpheus-frontend:latest -f Containerfile.frontend ./dashboard
```

Create the persistent data volume and run the services:

```bash
podman volume create orpheus-data

podman run -d --name orpheus-backend \
  --env-file .env \
  -v orpheus-data:/app/data \
  -p 8080:8080 \
  localhost/orpheus-backend:latest

podman run -d --name orpheus-frontend \
  -p 8081:80 \
  localhost/orpheus-frontend:latest
```

For a compose-style setup:

```bash
podman-compose up -d
```

Podman Quadlet files can be used to manage both containers as systemd services in production.

## CI/CD

GitHub Actions runs on every push and pull request. The pipeline checks:

- Go formatting
- Dependency resolution
- `go vet`
- Race-enabled tests
- Application build
- Go vulnerability advisories
- Exposed secrets with Gitleaks
- Dependency changes in pull requests

## Documentation

Detailed technical documentation and the implementation roadmap are available locally in the `docs/` directory.

## License

License information will be added when the project license is selected.
