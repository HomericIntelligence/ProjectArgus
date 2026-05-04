# AGENTS.md — Multi-Agent Coordination for ProjectArgus

This file describes how AI agents and automated tooling should interact with this
repository. It follows the conventions established across HomericIntelligence projects.

## Repository Role

ProjectArgus is a **read-only observability consumer**. It scrapes metrics from
other HomericIntelligence services and does not write to or modify them. Agents
working in this repository should maintain this invariant.

## Scope of Automated Changes

Agents are permitted to make the following changes autonomously:

| Area | Permitted |
|------|-----------|
| `configs/prometheus.yml` — add/edit scrape jobs | Yes |
| `configs/loki.yml` — adjust retention, limits | Yes |
| `configs/promtail.yml` — add log scrape targets | Yes |
| `dashboards/*.json` — add or update Grafana dashboards | Yes |
| `rules/*.yml` — add or update alert and recording rules | Yes |
| `exporter/exporter.py` — add metrics, fix bugs | Yes |
| `tests/` — add or update tests | Yes |
| `docker-compose.yml` — add/remove services or env vars | **Review required** |
| `.github/workflows/` — CI changes | **Review required** |
| Credentials, secrets, `.env` | **Never** |

## Coordination Protocol

When multiple agents work on this repository simultaneously:

1. **One scrape target per PR** — Do not bundle unrelated Prometheus job additions.
2. **Metric names are stable API** — Never rename an existing metric without a
   deprecation period and dashboard update in the same PR.
3. **Alert rules must not regress** — `just validate` must pass before merging.
4. **Test coverage** — Any change to `exporter/exporter.py` must include or update
   tests in `tests/test_exporter.py`. `just test` must pass.

## Validation Gates

Before opening a PR, agents must verify:

```bash
just validate          # docker compose config + YAML lint
just test              # pytest unit tests
pixi run ruff check exporter/exporter.py
pixi run bandit -ll exporter/exporter.py
```

## Prohibited Actions

- Do not commit `.env` or any file containing credentials.
- Do not modify external service configurations (Agamemnon, Nestor, NATS).
- Do not remove existing alert rules without explicit instruction.
- Do not tighten the Prometheus scrape interval below 10s.
- Do not expose internal ports (Loki 3100, Prometheus 9090) beyond localhost.
