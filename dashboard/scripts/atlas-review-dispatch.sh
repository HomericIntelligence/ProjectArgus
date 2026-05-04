#!/usr/bin/env bash
# Usage: atlas-review-dispatch.sh <MILESTONE> <PR_URL> [AGAMEMNON_URL]
# Creates a review team for the given milestone, files 6 review tasks, prints TEAM_ID.
set -euo pipefail

MILESTONE=${1:?Usage: atlas-review-dispatch.sh MILESTONE PR_URL [AGAMEMNON_URL]}
PR_URL=${2:?Usage: atlas-review-dispatch.sh MILESTONE PR_URL [AGAMEMNON_URL]}
AGAMEMNON_URL=${3:-http://localhost:8080}

TEAM_RESP=$(curl -fsS -X POST "${AGAMEMNON_URL}/v1/teams" \
  -H 'Content-Type: application/json' \
  -d "{\"name\":\"atlas-${MILESTONE}-review\",\"purpose\":\"review-gate\",\"milestone\":\"${MILESTONE}\",\"pr\":\"${PR_URL}\"}")

TEAM_ID=$(echo "$TEAM_RESP" | jq -r '.team.id // .id // empty')
if [ -z "$TEAM_ID" ]; then
  echo "ERROR: failed to create team. Response: $TEAM_RESP" >&2
  exit 1
fi

echo "TEAM_ID=${TEAM_ID}"
echo "PR=${PR_URL}"
echo "MILESTONE=${MILESTONE}"

for DIM in arch code security ux ops docs; do
  TASK_RESP=$(curl -fsS -X POST "${AGAMEMNON_URL}/v1/teams/${TEAM_ID}/tasks" \
    -H 'Content-Type: application/json' \
    -d "{\"subject\":\"${DIM} review — ${MILESTONE}\",\"description\":\"Review dimension: ${DIM}. Milestone: ${MILESTONE}. PR: ${PR_URL}\",\"type\":\"atlas-review\",\"dimension\":\"${DIM}\"}")
  TASK_ID=$(echo "$TASK_RESP" | jq -r '.task.id // .id // empty')
  echo "  task created: ${TASK_ID} [${DIM}]"
done
