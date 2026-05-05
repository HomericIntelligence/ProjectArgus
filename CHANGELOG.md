# Changelog

All notable changes to ProjectArgus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased] — targeting v0.2.0

### Added

- Atlas service (M1–M3): Go module with Chi server, hello-world page, SSE endpoint,
  NATS/JetStream subscribers, Tailscale sources, service probe matrix, and hosts page
- `.dockerignore` to minimise Docker build context
- Atlas M4: Agamemnon poller, agents page, agent/task detail pages with live SSE event tail
- Atlas M5: Grafana iframe matrix (8 panels, time-range selector), NATS monitoring page
  (JetStream streams + connections, 5 s htmx poll), Mnemosyne skill registry browser
  (live search, lazy goldmark render)
- Atlas M6: pluggable auth middleware (none/basic/bearer) with SSE `?token=` fallback;
  `/metrics` Prometheus exposition (hand-rolled text format, no external dep);
  Prometheus alert rules for NATS disconnect, poller errors, SSE drops, upstream outages;
  Atlas scrape job wired into `configs/prometheus.yml`; golangci-lint config; e2e test stubs;
  architecture doc (`dashboard/docs/architecture.md`)

### Changed

- CI lint job consolidates typecheck steps; unified required-checks workflow added
- Python base image bumped to `3.14.0-slim`
- Atlas CI job extended: golangci-lint, templ generate no-op check, Docker build
- `X-Frame-Options` changed from `DENY` to `SAMEORIGIN` (Atlas embeds its own Grafana panels)
- CSP `frame-src` now includes `ATLAS_LOKI_URL` and optionally `ATLAS_NATS_DASHBOARD_URL`

### Fixed

- Atlas SSE heartbeat data race in tests
- Atlas M2 review-wave findings (healthz plaintext, hermes port, README line wrap)
- CI: valid job IDs in `_required.yml` (no slashes allowed in keys)
- Dependabot: removed invalid Docker `package-ecosystem` entry
- Atlas NATS poller: `consumer_count` JSON tag was `num_consumers` (stream consumers always read 0)
- Atlas aggregate review script: used `.verdict` field; Agamemnon stores verdict in `.result`
- Atlas Grafana handler: `from`/`to` query params now validated before embedding in iframe URLs
- Atlas Mnemosyne handlers: nil guards added for unconfigured `ATLAS_MNEMOSYNE_SKILLS_DIR`

## [0.1.0] - 2026-04-23

### Added

- Prometheus scrape stack with 15s global interval
- Loki log aggregation with 30-day retention
- Promtail log shipping from container stdout and host log files
- Grafana dashboards: agent-health, nats-events, task-throughput
- Custom Python exporter (`exporter/exporter.py`) scraping ProjectAgamemnon and
  NATS HTTP APIs, exposing metrics on port 9101
- Prometheus alerting rules in `rules/agent-alerts.yml` for agent health
- Grafana Alertmanager contact point provisioning
- Docker Compose orchestration with pinned image versions
- `justfile` task runner with targets for start, stop, reload, scrape, and dashboard import
- CI workflow (yamllint + Python lint) on GitHub Actions
- CLAUDE.md project conventions and development guidelines
- CONTRIBUTING.md, SECURITY.md, LICENSE at repository root
- `.gitignore` covering `.pixi/`, IDE files, OS files, and secrets
- `pixi.toml` with locked dependencies (`just`, `jq`)

[Unreleased]: https://github.com/HomericIntelligence/ProjectArgus/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/HomericIntelligence/ProjectArgus/releases/tag/v0.1.0
