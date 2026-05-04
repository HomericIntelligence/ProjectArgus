package poller

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/HomericIntelligence/atlas/internal/config"
	"github.com/HomericIntelligence/atlas/internal/store"
)

// varzResponse is the relevant subset of the NATS /varz monitoring endpoint.
type varzResponse struct {
	Connections int   `json:"connections"`
	InMsgs      int64 `json:"in_msgs"`
	OutMsgs     int64 `json:"out_msgs"`
}

// jszResponse is the relevant subset of the NATS /jsz monitoring endpoint.
type jszResponse struct {
	NumStreams int `json:"num_streams"`
}

// NATSPoller polls the NATS monitoring endpoints for server statistics.
type NATSPoller struct {
	base
	cache *store.Cache
	url   string
}

// NewNATSPoller constructs a NATSPoller with a 3-second HTTP timeout.
func NewNATSPoller(cfg *config.Config, cache *store.Cache) *NATSPoller {
	return &NATSPoller{
		base: base{
			name:   "nats",
			client: &http.Client{Timeout: 3 * time.Second},
		},
		cache: cache,
		url:   cfg.NATSMonURL,
	}
}

// Start runs the poller in a ticker loop until ctx is cancelled.
func (p *NATSPoller) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Fetch immediately on start.
	p.fetch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.fetch(ctx)
		}
	}
}

// fetch retrieves stats from /varz and /jsz and updates the cache.
// On any error it logs a warning and leaves the cache unchanged.
func (p *NATSPoller) fetch(ctx context.Context) {
	var varz varzResponse
	if err := p.getJSON(ctx, p.url+"/varz", &varz); err != nil {
		slog.Warn("nats poller: failed to fetch /varz", "err", err)
		return
	}

	var jsz jszResponse
	if err := p.getJSON(ctx, p.url+"/jsz", &jsz); err != nil {
		slog.Warn("nats poller: failed to fetch /jsz", "err", err)
		return
	}

	p.cache.SetNATSStats(store.NATSStats{
		Connections: varz.Connections,
		Streams:     jsz.NumStreams,
		InMsgs:      varz.InMsgs,
		OutMsgs:     varz.OutMsgs,
	})
}
