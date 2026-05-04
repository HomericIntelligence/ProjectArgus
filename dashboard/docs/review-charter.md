# Atlas Review Charter

All Atlas milestone PRs are gated by a 6-dimension myrmidon review wave dispatched via
Agamemnon before merge. No milestone PR merges until 6/6 dimensions return `approved`.

## Review Dimensions

| Dimension | What is reviewed | Authoritative source |
|-----------|-----------------|---------------------|
| `arch` | API contract fidelity (pollers/handlers match Agamemnon routes.cpp:59-392, Nestor routes.cpp:14-80), NATS subject schema (ADR-005), package layout | `control/ProjectAgamemnon/src/routes.cpp`, ADR-005 |
| `code` | Go idioms, `go vet`, `golangci-lint`, `go test -race`, no goroutine leaks | `.golangci.yml`, `go test -race ./...` |
| `security` | Auth paths, CSP `frame-src` allowlist, iframe sandbox (`allow-scripts allow-popups` only — never `allow-same-origin`), SSE token fallback, no secrets in code | `internal/server/{auth,middleware}.go` |
| `ux` | templ/htmx patterns (`sse-swap`, not `hx-trigger="sse:*"`), accessibility, Grafana kiosk panels, connection dot (green/amber/red), pause/resume | `web/templates/`, `web/static/` |
| `ops` | Compose healthcheck, `depends_on`, port 3002, resource limits, Tailscale socket mount commented out, env var docs | `infrastructure/ProjectArgus/docker-compose.yml` |
| `docs` | README, runbook reproducibility, ADR alignment, architecture.md updated | `docs/architecture.md`, `dashboard/README.md` |

## Pass/Fail Criteria

A dimension passes when the myrmidon reviewer PATCHes the task with:

```json
{"status": "completed", "verdict": "approved"}
```

A dimension fails when PATCHed with:

```json
{"status": "completed", "verdict": "rejected", "findings": "..."}
```

## Usage

```bash
# 1. Dispatch review wave (after PR is open and CI is green)
TEAM=$(just atlas-review-dispatch MILESTONE PR_URL | grep TEAM_ID | cut -d= -f2)

# 2. Wait for all 6 reviewers to complete (myrmidon agents pick up tasks automatically)
just atlas-review-aggregate MILESTONE "$TEAM"  # exits 0 when 6/6 approved

# 3. Post GitHub status check
just atlas-review-status MILESTONE "$TEAM" "$(git rev-parse HEAD)"
```

## Known Issues / Warnings

- `completed-approved` in issue specs refers to `status: completed AND verdict: approved` — these are two separate JSON fields in the Agamemnon task store.
- `ATLAS_HERMES_URL` default is `http://hermes:8080` (NOT 8085; canonical per `infrastructure/ProjectHermes/src/hermes/config.py:35`).
- Never call `goldmark.WithUnsafe()` — omitting it is already safe; calling it enables raw HTML.
- iframe sandbox: use `allow-scripts allow-popups` only — never combine `allow-scripts` with `allow-same-origin` (sandbox escape).
