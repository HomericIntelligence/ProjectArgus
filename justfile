# === Variables ===

MAESTRO_URL := "http://172.20.0.1:23000"
GRAFANA_PORT := "3000"
GRAFANA_URL  := "http://localhost:" + GRAFANA_PORT
GRAFANA_AUTH := "admin:admin"

# === Default ===

default:
    @just --list

# === Services ===

# Start all observability services
start:
    docker compose up -d

# Stop all services
stop:
    docker compose down

# Show running container status
status:
    docker compose ps

# Tail logs for a specific service (e.g. just logs prometheus)
logs SERVICE:
    docker compose logs -f {{SERVICE}}

# === Prometheus ===

# Hot-reload Prometheus configuration (no restart needed)
reload-prometheus:
    curl -s -X POST http://localhost:9090/-/reload && echo "Prometheus config reloaded."

# Query Prometheus to verify all scrape targets are up
test-scrape:
    @echo "Querying Prometheus for 'up' metric..."
    curl -s "http://localhost:9090/api/v1/query?query=up" | jq '.data.result[] | {job: .metric.job, instance: .metric.instance, up: .value[1]}'

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
