# ProjectArgus

Observability stack for the HomericIntelligence mesh. ProjectArgus provides centralized metrics collection, log aggregation, and dashboards for all components of the HomericIntelligence ecosystem.

## What Gets Monitored

- **ai-maestro**: Agent health (`/api/agents/health`), system diagnostics (`/api/diagnostics`)
- **NATS**: Message rates, JetStream storage, subject counts (port 8222)
- **Nomad**: Job and allocation metrics (port 4646)
- **Containers**: All Podman/Docker container logs via Promtail

## Stack

| Component  | Role                          | Port |
|------------|-------------------------------|------|
| Prometheus | Metrics scraping and storage  | 9090 |
| Loki       | Log aggregation               | 3100 |
| Promtail   | Log shipping to Loki          | —    |
| Grafana    | Dashboards and visualization  | 3000 |

## Quick Start

```bash
just start
```

Then access Grafana at http://localhost:3000 (default credentials: admin / admin).

## Dashboards

- **AI Maestro - Agent Health**: Agent count, active/hibernated agents, uptime
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
