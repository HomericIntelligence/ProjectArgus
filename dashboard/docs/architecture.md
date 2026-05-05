# Atlas Architecture

## Component Overview

Atlas (the dashboard binary) consists of five subsystems:

```
NATS (JetStream)
    └── nats.Subscriber  →  events.Bus  →  handlers.SSE  →  SSE clients
                                        └── (ring buffer, topic filter)

REST pollers (5 s interval)
    ├── poller.AgamemnonPoller  ─┐
    ├── poller.NATSPoller       ─┤→  store.Cache  →  handlers.HostsHandler  →  htmx pages
    └── poller.TailscalePoller  ─┘

Tailscale source
    └── tailscale.Source  →  store.Cache  →  /hosts, /api/hosts

Grafana embedding
    └── grafana.KnownPanels  →  handlers.GrafanaPage  →  iframe matrix

Mnemosyne browser
    └── mnemosyne.Reader (5-min TTL cache)  →  handlers.MnemosynePage + partials
```

## Concurrency model

- **Cache**: `sync.RWMutex`-protected; writers are poller goroutines, readers are HTTP handlers
- **Event Bus**: `sync.RWMutex`-protected subscriber map; each subscriber gets a `chan []byte` with capacity 1000
- **Ring Buffer**: `sync.RWMutex`; capacity 256 per topic; replayed on SSE connect
- **Metrics**: `sync/atomic` for counters and gauges; labeled maps protected by `sync.RWMutex`

## Security

- All HTML responses include `X-Content-Type-Options: nosniff`, `X-Frame-Options: SAMEORIGIN`, `Referrer-Policy: strict-origin-when-cross-origin`
- CSP `frame-src` is built statically at startup from `ATLAS_GRAFANA_URL`, `ATLAS_LOKI_URL`, and optionally
  `ATLAS_NATS_DASHBOARD_URL` — no user input reaches the CSP header
- Auth middleware (none/basic/bearer) wraps all routes except `/healthz`, `/readyz`, `/metrics`
- SSE/EventSource bearer fallback via `?token=` allows JS EventSource to authenticate
- Mnemosyne markdown rendering uses goldmark with `WithUnsafe()` disabled — raw HTML in skill files is stripped
- Grafana `from`/`to` query params validated against `^(now(-[0-9]+(s|m|h|d|w|y))?|[0-9]{13})$`
  before embedding in iframe URLs
- iframe sandbox: `allow-scripts allow-popups` only — never `allow-same-origin` alongside `allow-scripts`
