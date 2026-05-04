#!/usr/bin/env bash
# Usage: atlas-review-aggregate.sh <MILESTONE> <TEAM_ID> [AGAMEMNON_URL]
# Exits 0 when 6/6 dimensions are completed+approved, 1 otherwise.
set -euo pipefail

MILESTONE=${1:?Usage: atlas-review-aggregate.sh MILESTONE TEAM_ID [AGAMEMNON_URL]}
TEAM_ID=${2:?Usage: atlas-review-aggregate.sh MILESTONE TEAM_ID [AGAMEMNON_URL]}
AGAMEMNON_URL=${3:-http://localhost:8080}

RESP=$(curl -fsS "${AGAMEMNON_URL}/v1/teams/${TEAM_ID}/tasks")
TASKS=$(echo "$RESP" | jq -r '.tasks[]? // .[]? | "\(.dimension // "unknown"):\(.status // "pending"):\(.verdict // "pending")"')

if [ -z "$TASKS" ]; then
  echo "Review wave ${MILESTONE} (team ${TEAM_ID}): 0 tasks found" >&2
  exit 1
fi

TOTAL=$(echo "$TASKS" | wc -l | tr -d ' ')
APPROVED=$(echo "$TASKS" | grep -c ':completed:approved' || true)

echo "Review wave ${MILESTONE} (team ${TEAM_ID}): ${APPROVED}/${TOTAL} approved"
echo ""
echo "$TASKS" | while IFS=: read -r dim status verdict; do
  if [ "$status" = "completed" ] && [ "$verdict" = "approved" ]; then
    echo "  ✓ ${dim}: ${status} (${verdict})"
  else
    echo "  ✗ ${dim}: ${status} (${verdict})"
  fi
done

[ "$APPROVED" -eq 6 ] && exit 0 || exit 1
