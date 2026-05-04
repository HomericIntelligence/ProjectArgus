# ProjectArgus

Observability stack for the HomericIntelligence mesh. ProjectArgus provides centralized metrics collection, log aggregation, and dashboards for all components of the HomericIntelligence ecosystem.

## What Gets Monitored

- **ProjectAgamemnon**: Agent health (`/v1/health`), agent list (`/v1/agents`)
- **ProjectNestor**: Research pipeline health (`/v1/health`)
- **NATS**: Message rates, JetStream storage, subject counts (port 8222)
- **Nomad**: Job and allocation metrics (port 4646)
- **Containers**: All Podman/Docker container logs via Promtail

## Stack

| Component  | Role                          | Port |
|------------|-------------------------------|------|
| Prometheus | Metrics scraping and storage  | 9090 |
| Loki       | Log aggregation               | 3100 |
| Promtail   | Log shipping to Loki          | —    |
| Grafana    | Dashboards and visualization  | 3001 |

## Quick Start

```bash
cp .env.example .env   # copy and edit for your environment
just start
```

Then access Grafana at http://localhost:3001 (default credentials: admin / admin).

## Environment Configuration

All environment-specific values live in a `.env` file that is **not** committed to version control. A fully-documented template is provided:

```bash
cp .env.example .env
```

Key variables:

| Variable | Default | Description |
|---|---|---|
| `AGAMEMNON_URL` | `http://172.20.0.1:8080` | Agamemnon API base URL |
| `NESTOR_URL` | `http://172.20.0.1:8081` | Nestor API base URL |
| `NATS_URL` | `http://172.24.0.1:8222` | NATS monitoring endpoint |
| `GF_SECURITY_ADMIN_PASSWORD` | `admin` | Grafana admin password |
| `GRAFANA_PORT` | `3001` | Host port for Grafana |
| `PROMETHEUS_PORT` | `9090` | Host port for Prometheus |
| `LOKI_PORT` | `3100` | Host port for Loki |
| `EXPORTER_PORT` | `9100` | Host port for argus-exporter |

`docker compose` and `just` both load `.env` automatically — no extra steps required.

## Dashboards

- **HomericIntelligence - Agent Health**: Agent count, active/hibernated agents, uptime
- **NATS Event Bus**: Message rate, JetStream storage, subject counts
- **Task Throughput**: Tasks created/completed/failed per hour, dispatch latency

## Configuration

All scrape targets and service configs live in `configs/`. Alert rules are in `rules/`. Grafana dashboards (JSON) are in `dashboards/` and auto-provisioned on startup.

## Common Commands

```bash
just start                   # Start all services
just stop                    # Stop all services
just status                  # Show running containers
just logs prometheus         # Tail Prometheus logs
just reload-prometheus       # Hot-reload Prometheus config
just test-scrape             # Verify all scrape targets are up
just import-dashboards       # Push dashboards to Grafana API
```
