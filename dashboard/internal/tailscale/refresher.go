package tailscale

import (
	"context"
	"log/slog"
	"time"
)

// DeviceStore is the interface satisfied by any cache that can store Tailscale
// devices.  Using an interface here avoids a circular import between the
// tailscale and store packages.
type DeviceStore interface {
	SetDevices(d []Device)
}

// Refresher periodically polls a Source and writes the results into a
// DeviceStore.  It runs in a single goroutine and exits when the provided
// context is cancelled, so there are no goroutine leaks.
type Refresher struct {
	src      Source
	store    DeviceStore
	interval time.Duration
}

// NewRefresher creates a Refresher that polls src every interval and stores
// results in store.
func NewRefresher(src Source, store DeviceStore, interval time.Duration) *Refresher {
	return &Refresher{
		src:      src,
		store:    store,
		interval: interval,
	}
}

// Start begins the polling loop.  It performs an immediate poll on entry, then
// ticks every r.interval.  It returns when ctx is cancelled.
func (r *Refresher) Start(ctx context.Context) {
	// Perform an immediate poll before the first tick so the cache is populated
	// as quickly as possible after startup.
	r.poll(ctx)

	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.poll(ctx)
		}
	}
}

func (r *Refresher) poll(ctx context.Context) {
	devices, err := r.src.Devices(ctx)
	if err != nil {
		slog.Warn("tailscale refresher: failed to fetch devices", "err", err)
		return
	}
	r.store.SetDevices(devices)
}
