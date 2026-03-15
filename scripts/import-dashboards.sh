#!/usr/bin/env bash
# import-dashboards.sh — Import all Grafana dashboard JSON files via the HTTP API.
# Reads GRAFANA_ADMIN_PASSWORD from env (default: admin).
set -euo pipefail

GRAFANA_PORT="${GRAFANA_PORT:-3000}"
GRAFANA_URL="http://localhost:${GRAFANA_PORT}"
GRAFANA_ADMIN_PASSWORD="${GRAFANA_ADMIN_PASSWORD:-admin}"
GRAFANA_AUTH="admin:${GRAFANA_ADMIN_PASSWORD}"

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DASHBOARDS_DIR="${SCRIPT_DIR}/../dashboards"

if [[ ! -d "$DASHBOARDS_DIR" ]]; then
    echo "ERROR: dashboards directory not found at $DASHBOARDS_DIR" >&2
    exit 1
fi

shopt -s nullglob
files=("$DASHBOARDS_DIR"/*.json)

if [[ ${#files[@]} -eq 0 ]]; then
    echo "No dashboard JSON files found in $DASHBOARDS_DIR"
    exit 0
fi

for f in "${files[@]}"; do
    echo "Importing $(basename "$f") ..."
    payload=$(jq -n --slurpfile dash "$f" '{"dashboard": $dash[0], "overwrite": true, "folderId": 0}')
    result=$(curl -sf -u "$GRAFANA_AUTH" \
        -H "Content-Type: application/json" \
        -d "$payload" \
        "${GRAFANA_URL}/api/dashboards/db")
    status=$(echo "$result" | jq -r '.status // "unknown"')
    echo "  -> status: $status"
done

echo "Dashboard import complete."
