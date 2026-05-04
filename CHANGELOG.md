# Changelog

All notable changes to ProjectArgus will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Unit test suite for `exporter/exporter.py` covering `collect()`, `_fetch()`, `_health_check()`, and HTTP handler (`tests/test_exporter.py`)
- `# HELP` lines for all metrics emitted by the exporter
- Health checks for all five Docker Compose services
- Resource limits (`memory`, `cpus`) for all Docker Compose services
- `GF_ANALYTICS_REPORTING_ENABLED: "false"` and related Grafana analytics env vars to prevent startup hangs
- `tests/` and `lint/` pixi feature environments with `pytest` and `ruff`
- `security` pixi task backed by `bandit`
- CI jobs: `test`, `lint`, `security` in addition to existing `validate`
- `.pre-commit-config.yaml` with yamllint, ruff, and bandit hooks
- `.editorconfig` for consistent line endings and indentation

### Changed
- `dashboards/nats-events.json`: all `gnatsd_varz_*` metric references replaced with actual exporter metric names (`nats_in_msgs_total`, `nats_out_msgs_total`, `nats_in_bytes_total`, `nats_out_bytes_total`, `nats_jetstream_bytes`, `nats_connections`)
- "Active Subscriptions" stat panel renamed to "Active Connections" to match `nats_connections` semantics
- `exporter/exporter.py`: fixed mutable default argument (`labels={}` → `labels=None`)
- `exporter/exporter.py`: each metric's `# TYPE` line is now emitted exactly once (no duplicates)
- `docker-compose.yml`: `GF_SECURITY_ADMIN_PASSWORD` now reads from `${GRAFANA_ADMIN_PASSWORD}` env var (default `changeme`)
- `docker-compose.yml`: Loki, Promtail, and argus-exporter ports removed from host-level exposure; services communicate over the `argus` bridge network
- `docker-compose.yml`: `AGAMEMNON_URL`, `NESTOR_URL`, `NATS_URL` now use env-var substitution with defaults
- `.env.example`: updated default password from `admin` to `changeme`, documented all variables
- `justfile`: fixed `GRAFANA_PORT` from `3000` to `3001` (matches compose port mapping); `GRAFANA_AUTH` reads from `GRAFANA_ADMIN_PASSWORD` env var
- CI branch trigger expanded to include `feature/**`, `fix/**`, `chore/**` branches
- `.gitignore`: added Python cache dirs (`__pycache__/`, `*.pyc`, `.pytest_cache/`, `.ruff_cache/`)

## [0.1.0] - 2026-03-22

### Added
- Initial ProjectArgus observability stack: Prometheus, Loki, Promtail, Grafana, homeric-exporter
- Grafana dashboards: `agent-health`, `nats-events`, `task-throughput`
- Prometheus alert rules: `AgamemnonDown`, `NestorDown`, `ExporterScrapeStale`, `HighTaskFailureRate`
- `justfile` with `start`, `stop`, `status`, `logs`, `reload-prometheus`, `test-scrape`, `import-dashboards`
- `pixi.toml` project configuration
- `CLAUDE.md` AI agent guidance
- `LICENSE`, `SECURITY.md`, `CONTRIBUTING.md`, `CODE_OF_CONDUCT.md`
