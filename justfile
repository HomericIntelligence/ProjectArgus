# === Variables ===

compose_cmd := if `command -v podman-compose 2>/dev/null || true` != "" { "podman-compose" } else { "docker compose" }

AGAMEMNON_URL := "http://172.20.0.1:8080"
GRAFANA_PORT := "3000"
GRAFANA_URL  := "http://localhost:" + GRAFANA_PORT

# === Default ===

default:
    @just --list

# === Services ===

# Start all observability services (validates .env exists first)
start:
    @test -f .env || { echo "ERROR: .env not found — copy .env.example to .env and fill in values"; exit 1; }
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
    {{compose_cmd}} exec prometheus wget -qO- "http://localhost:9090/api/v1/query?query=up" | python3 -c "import json,sys; [print(r['metric'].get('job','?'), r['metric'].get('instance','?'), r['value'][1]) for r in json.load(sys.stdin)['data']['result']]"

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
# Reads GRAFANA_ADMIN_PASSWORD from .env (required — never hardcoded)
import-dashboards:
    @test -f .env || { echo "ERROR: .env not found — copy .env.example to .env and fill in values"; exit 1; }
    GRAFANA_PORT={{GRAFANA_PORT}} GRAFANA_ADMIN_PASSWORD=$$(grep '^GF_ADMIN_PASSWORD=' .env | cut -d= -f2-) ./scripts/import-dashboards.sh
