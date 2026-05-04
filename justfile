# === Variables ===

compose_cmd := if `command -v podman-compose 2>/dev/null || true` != "" { "podman-compose" } else { "docker compose" }

AGAMEMNON_URL := "http://172.20.0.1:8080"
GRAFANA_PORT := "3000"
GRAFANA_URL  := "http://localhost:" + GRAFANA_PORT
GRAFANA_AUTH := "admin:admin"

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

# Restart all services (stop then start)
restart:
    {{compose_cmd}} down
    {{compose_cmd}} up -d

# Remove all containers and volumes (destructive — data loss!)
clean:
    {{compose_cmd}} down -v

# Validate docker-compose config and YAML files
validate:
    {{compose_cmd}} config --quiet
    @echo "Config is valid."

# Run local test suite
test:
    python3 -m pytest tests/ -v 2>/dev/null || python3 -m unittest discover -s tests -v

# Tail logs for a specific service (e.g. just logs prometheus)
logs SERVICE:
    {{compose_cmd}} logs -f {{SERVICE}}

# === Prometheus ===

# Hot-reload Prometheus configuration via SIGHUP (--web.enable-lifecycle is disabled)
reload-prometheus:
    {{compose_cmd}} kill -s HUP prometheus && echo "Prometheus config reloaded."

# Query Prometheus to verify all scrape targets are up
test-scrape:
    @echo "Querying Prometheus for 'up' metric..."
    curl -s "http://localhost:9090/api/v1/query?query=up" | jq '.data.result[] | {job: .metric.job, instance: .metric.instance, up: .value[1]}'

# Manually test Agamemnon and Nestor health endpoints
scrape-agamemnon:
    ./scripts/scrape-agamemnon.sh {{AGAMEMNON_URL}}

# === Backup & Restore ===

# Back up Prometheus and Loki data volumes to ./backups/
backup:
    ./scripts/backup.sh

# Restore a volume from a backup file: just restore <volume> <file>
restore VOLUME FILE:
    ./scripts/restore.sh {{VOLUME}} {{FILE}}

# === Grafana ===

# Import all JSON dashboards from dashboards/ into Grafana via API
import-dashboards:
    GRAFANA_PORT={{GRAFANA_PORT}} GRAFANA_ADMIN_PASSWORD=$(echo "{{GRAFANA_AUTH}}" | cut -d: -f2) ./scripts/import-dashboards.sh

# === Exporter ===

# Build argus-exporter image locally (tagged argus-exporter:local)
build-exporter:
    docker build -t argus-exporter:local ./exporter

# Start the stack using a locally built exporter (offline dev — bypasses registry pull)
run-exporter-local: build-exporter
    docker run --rm --network argus \
        -e AGAMEMNON_URL=${AGAMEMNON_URL} \
        -e NESTOR_URL=${NESTOR_URL} \
        -e NATS_URL=${NATS_URL} \
        -p 127.0.0.1:9100:9100 \
        argus-exporter:local
