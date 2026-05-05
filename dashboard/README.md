# Atlas — HomericIntelligence Dashboard

Atlas is the unified observability dashboard for the HomericIntelligence distributed agent mesh.
It provides a real-time overview of agents, tasks, NATS streams, and hosts via a lightweight
Go/Chi HTTP server with a dark-themed UI built on htmx and SSE.

## Quick Start

```bash
# Run with defaults (listens on :3002)
go run ./cmd/argus-dashboard

# Health check
curl -fsS http://localhost:3002/healthz

# Custom listen address
ATLAS_LISTEN_ADDR=:8090 go run ./cmd/argus-dashboard
```

## Configuration

All configuration is via environment variables with the `ATLAS_` prefix:

| Variable | Default | Description |
|---|---|---|
| `ATLAS_LISTEN_ADDR` | `:3002` | HTTP listen address |
| `ATLAS_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `ATLAS_NATS_URL` | `nats://nats:4222` | NATS server URL |
| `ATLAS_NATS_MON_URL` | `http://nats:8222` | NATS monitoring URL |
| `ATLAS_AGAMEMNON_URL` | `http://agamemnon:8080` | Agamemnon API URL |
| `ATLAS_NESTOR_URL` | `http://nestor:8081` | Nestor API URL |
| `ATLAS_HERMES_URL` | `http://hermes:8080` | Hermes event bridge URL |
| `ATLAS_PROMETHEUS_URL` | `http://prometheus:9090` | Prometheus URL |
| `ATLAS_GRAFANA_URL` | `http://grafana:3000` | Grafana URL |
| `ATLAS_AUTH_MODE` | `none` | Auth mode (none/basic/bearer) |
| `ATLAS_TAILSCALE_SOURCE` | `static` | Device discovery: `static`, `cli`, `api`, `auto` |
| `ATLAS_WORKER_HOST_IP` | `127.0.0.1` | Static source: IP of worker host |
| `ATLAS_CONTROL_HOST_IP` | `127.0.0.1` | Static source: IP of control host |
| `ATLAS_TAILSCALE_API_KEY` | `` | API source: Tailscale API key |
| `ATLAS_TAILNET_NAME` | `` | API source: Tailnet name (e.g. `example.com`) |
| `ATLAS_POLL_AGAMEMNON_MS` | `5000` | Poll interval for Agamemnon in ms |
| `ATLAS_NATS_DASHBOARD_URL` | `` | Optional: URL of external nats-dashboard (linked on /nats page) |
| `ATLAS_NATS_TOP_URL` | `` | Optional: ttyd URL serving nats-top (embedded as iframe on /nats page) |
| `ATLAS_MNEMOSYNE_SKILLS_DIR` | `/mnt/mnemosyne/skills` | Path to Mnemosyne skills directory (read by /mnemosyne page) |
| `ATLAS_LOKI_URL` | `http://loki:3100` | Loki URL (included in CSP frame-src) |
| `ATLAS_EXPORTER_URL` | `http://argus-exporter:9100` | Homeric exporter URL |

## SSE Event Stream

Atlas exposes a real-time event stream at `/events` using Server-Sent Events.

```
GET /events?topics=agent,task&replay=20
```

| Parameter | Description |
|---|---|
| `topics` | Comma-separated topic filter. Omit to receive all topics. |
| `replay` | Number of buffered events to replay on connect (ring buffer, max 256). |

**Topics** (derived from NATS subject prefix):

| Topic | NATS stream | Subject pattern |
|---|---|---|
| `agent` | `homeric-agents` | `hi.agents.>` |
| `task` | `homeric-tasks` | `hi.tasks.>` |
| `myrmidon` | `homeric-myrmidon` | `hi.myrmidon.>` |
| `research` | `homeric-research` | `hi.research.>` |
| `pipeline` | `homeric-pipeline` | `hi.pipeline.>` |
| `log` | `homeric-logs` | `hi.logs.>` |

**Wire format** (per event):

```
event: {topic}
data: {json payload}

```

Keepalive comment frames are sent every 15 seconds:

```
: heartbeat

```

## HTTP Endpoints

| Method | Path | Description |
|---|---|---|
| `GET` | `/` | Overview page |
| `GET` | `/hosts` | Tailscale host grid — cards refresh every 5 s via htmx |
| `GET` | `/healthz` | Liveness probe — returns `ok` |
| `GET` | `/readyz` | Readiness probe — returns `ok` |
| `GET` | `/events` | SSE event stream (see below) |
| `GET` | `/api/hosts` | JSON array of hosts with per-service probe results |
| `GET` | `/partials/host/{name}` | htmx fragment — single host card (used by 5 s poll) |
| `GET` | `/agents` | Agents list page with filter bar and live SSE row-swap |
| `GET` | `/partials/agents/table` | htmx fragment — filtered agents tbody |
| `GET` | `/agents/{id}` | Agent detail page with live 50-event tail |
| `GET` | `/tasks/{id}` | Task detail page with live event tail |
| `GET` | `/grafana` | Grafana iframe panel matrix (8 panels, time-range selector) |
| `GET` | `/nats` | NATS monitoring page — JetStream streams, connections, external links |
| `GET` | `/partials/nats/streams` | htmx fragment — JetStream streams table (5 s poll) |
| `GET` | `/partials/nats/connections` | htmx fragment — NATS connections table (5 s poll) |
| `GET` | `/mnemosyne` | Mnemosyne skill registry browser with live search |
| `GET` | `/partials/mnemosyne/search` | htmx fragment — filtered skill list |
| `GET` | `/partials/mnemosyne/skill/{name}` | htmx fragment — rendered markdown body of a skill |
| `GET` | `/static/*` | Static assets (CSS, JS) |

## Authentication

Set `ATLAS_AUTH_MODE` to configure the auth gate:

| Mode | Behaviour |
|------|-----------|
| `none` (default) | No authentication required |
| `bearer` | `Authorization: Bearer <token>` header required; SSE endpoints also accept `?token=<token>` |
| `basic` | `Authorization: Basic <base64(user:pass)>` required |

Set `ATLAS_AUTH_BEARER_TOKEN`, `ATLAS_AUTH_USER`, `ATLAS_AUTH_PASS` accordingly.
`/healthz`, `/readyz`, and `/metrics` are always unauthenticated.

## Metrics

Atlas exposes Prometheus metrics at `/metrics` (always unauthenticated):

| Metric | Type | Description |
|--------|------|-------------|
| `atlas_build_info` | gauge | Build info (version, goversion labels) |
| `atlas_nats_connected` | gauge | 1 if NATS is connected, 0 otherwise |
| `atlas_sse_connected_clients` | gauge | Active SSE client connections |
| `atlas_poll_errors_total{source}` | counter | REST poller errors by source |
| `atlas_poll_duration_seconds{source}` | histogram | REST poll latency |
| `atlas_sse_dropped_total{subscriber}` | counter | SSE events dropped for slow clients |
| `atlas_event_parse_errors_total{stream}` | counter | NATS event parse errors |
| `atlas_nats_messages_processed_total{stream}` | counter | NATS messages processed |

## Building

```bash
go build -ldflags "-X github.com/HomericIntelligence/atlas/internal/version.Version=$(git describe --tags --always)" ./cmd/argus-dashboard
```

## Template Generation

Templates use [templ](https://templ.guide/). Generated `*_templ.go` files are committed. To regenerate:

```bash
templ generate ./...
```
