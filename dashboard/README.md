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
| `ATLAS_TAILSCALE_SOURCE` | `static` | Tailscale source (static/api/socket) |
| `ATLAS_POLL_AGAMEMNON_MS` | `5000` | Poll interval for Agamemnon in ms |

## Building

```bash
go build -ldflags "-X github.com/HomericIntelligence/atlas/internal/version.Version=$(git describe --tags --always)" ./cmd/argus-dashboard
```

## Template Generation

Templates use [templ](https://templ.guide/). Generated `*_templ.go` files are committed. To regenerate:

```bash
templ generate ./...
```
