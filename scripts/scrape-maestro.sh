#!/usr/bin/env bash
# scrape-maestro.sh — Manually test ai-maestro health endpoints and print a summary.
# Usage: ./scripts/scrape-maestro.sh [MAESTRO_URL]

set -euo pipefail

MAESTRO_URL="${1:-http://172.20.0.1:23000}"

# Color helpers
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

ok()   { echo -e "${GREEN}[OK]${NC}    $*"; }
fail() { echo -e "${RED}[FAIL]${NC}  $*"; }
info() { echo -e "${YELLOW}[INFO]${NC}  $*"; }

echo ""
info "Scraping ai-maestro at ${MAESTRO_URL}"
echo "-------------------------------------------"

# --- /api/diagnostics ---
echo ""
info "GET ${MAESTRO_URL}/api/diagnostics"
DIAG_HTTP=$(curl -s -o /tmp/argus_diag.json -w "%{http_code}" --connect-timeout 5 "${MAESTRO_URL}/api/diagnostics" || echo "000")

if [[ "$DIAG_HTTP" == "200" ]]; then
    ok "HTTP ${DIAG_HTTP} — /api/diagnostics reachable"
    if command -v jq &>/dev/null; then
        echo ""
        echo "  Diagnostics summary:"
        jq -r 'to_entries[] | "  \(.key): \(.value)"' /tmp/argus_diag.json 2>/dev/null || cat /tmp/argus_diag.json
    else
        cat /tmp/argus_diag.json
    fi
else
    fail "HTTP ${DIAG_HTTP} — /api/diagnostics not reachable (check MAESTRO_URL and that ai-maestro is running)"
fi

# --- /api/agents/health ---
echo ""
info "GET ${MAESTRO_URL}/api/agents/health"
HEALTH_HTTP=$(curl -s -o /tmp/argus_health.json -w "%{http_code}" --connect-timeout 5 "${MAESTRO_URL}/api/agents/health" || echo "000")

if [[ "$HEALTH_HTTP" == "200" ]]; then
    ok "HTTP ${HEALTH_HTTP} — /api/agents/health reachable"
    if command -v jq &>/dev/null; then
        echo ""
        TOTAL=$(jq 'if type == "array" then length elif .agents then (.agents | length) else "?" end' /tmp/argus_health.json 2>/dev/null || echo "?")
        echo "  Agent count reported: ${TOTAL}"
        echo ""
        echo "  Raw response (truncated):"
        jq '.' /tmp/argus_health.json 2>/dev/null | head -40
    else
        cat /tmp/argus_health.json
    fi
else
    fail "HTTP ${HEALTH_HTTP} — /api/agents/health not reachable"
fi

# --- Prometheus target check ---
echo ""
info "Checking Prometheus scrape targets (http://localhost:9090)"
PROM_UP=$(curl -s --connect-timeout 3 "http://localhost:9090/api/v1/query?query=up" 2>/dev/null || echo "")

if [[ -n "$PROM_UP" ]]; then
    ok "Prometheus is reachable"
    if command -v jq &>/dev/null; then
        echo ""
        echo "  Scrape target status:"
        echo "$PROM_UP" | jq -r '.data.result[] | "  job=\(.metric.job) instance=\(.metric.instance) up=\(.value[1])"' 2>/dev/null
    fi
else
    fail "Prometheus not reachable at localhost:9090 — run 'just start' first"
fi

echo ""
echo "-------------------------------------------"
info "Done."
echo ""
