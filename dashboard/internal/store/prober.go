package store

import (
	"context"
	"log/slog"
	"time"

	"github.com/HomericIntelligence/atlas/internal/catalog"
)

// Prober runs periodic catalog.ProbeAll and stores results in Cache.
type Prober struct {
	cache    *Cache
	interval time.Duration
}

// NewProber creates a Prober that probes all devices in cache every interval.
func NewProber(cache *Cache, interval time.Duration) *Prober {
	return &Prober{cache: cache, interval: interval}
}

// Start begins the probe loop. It performs an immediate probe on entry, then
// ticks every p.interval. It returns when ctx is cancelled.
func (p *Prober) Start(ctx context.Context) {
	ticker := time.NewTicker(p.interval)
	defer ticker.Stop()
	p.runOnce(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.runOnce(ctx)
		}
	}
}

func (p *Prober) runOnce(ctx context.Context) {
	devices := p.cache.GetDevices()
	hosts := make([]catalog.HostAddr, len(devices))
	for i, d := range devices {
		hosts[i] = catalog.HostAddr{Hostname: d.Hostname, IP: d.TailscaleIP}
	}
	results := catalog.ProbeAll(ctx, hosts, nil)
	p.cache.SetProbes(results)
	slog.Debug("atlas: probe cycle complete", "results", len(results))
}
