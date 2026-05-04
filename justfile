# === Variables ===

compose_cmd := if `which podman-compose 2>/dev/null` != "" { "podman-compose" } else { "docker compose" }

# Load .env if present so just commands share the same config as docker-compose.
# set dotenv-load silently no-ops when .env does not exist.
set dotenv-load := true

# Defaults mirror docker-compose.yml defaults so the justfile works without a .env.
AGAMEMNON_URL := env_var_or_default("AGAMEMNON_URL", "http://172.20.0.1:8080")
GRAFANA_PORT  := env_var_or_default("GRAFANA_PORT", "3001")
GRAFANA_URL   := "http://localhost:" + GRAFANA_PORT
_admin_pass   := env_var_or_default("GF_SECURITY_ADMIN_PASSWORD", "admin")
GRAFANA_AUTH  := "admin:" + _admin_pass
PROMETHEUS_PORT := env_var_or_default("PROMETHEUS_PORT", "9090")

# === Default ===

default:
    @just --list

# === Services ===

# Start all observability services
start:
    {{compose_cmd}} up -d

# Stop all services
stop:
    {{compose_cmd}} down

# Show running container status
status:
    {{compose_cmd}} ps

# Tail logs for a specific service (e.g. just logs prometheus)
logs SERVICE:
    {{compose_cmd}} logs -f {{SERVICE}}

# === Prometheus ===

# Hot-reload Prometheus configuration (no restart needed)
reload-prometheus:
    curl -s -X POST http://localhost:{{PROMETHEUS_PORT}}/-/reload && echo "Prometheus config reloaded."

# Query Prometheus to verify all scrape targets are up
test-scrape:
    @echo "Querying Prometheus for 'up' metric..."
    curl -s "http://localhost:{{PROMETHEUS_PORT}}/api/v1/query?query=up" | jq '.data.result[] | {job: .metric.job, instance: .metric.instance, up: .value[1]}'

# Manually test Agamemnon and Nestor health endpoints
scrape-agamemnon:
    ./scripts/scrape-agamemnon.sh {{AGAMEMNON_URL}}

# === Grafana ===

# Import all JSON dashboards from dashboards/ into Grafana via API
import-dashboards:
    #!/usr/bin/env bash
    set -euo pipefail
    for f in dashboards/*.json; do
        echo "Importing $f ..."
        payload=$(jq -n --slurpfile dash "$f" '{"dashboard": $dash[0], "overwrite": true, "folderId": 0}')
        curl -s -u {{GRAFANA_AUTH}} \
            -H "Content-Type: application/json" \
            -d "$payload" \
            "{{GRAFANA_URL}}/api/dashboards/db" | jq '.status'
    done
    echo "Dashboard import complete."
