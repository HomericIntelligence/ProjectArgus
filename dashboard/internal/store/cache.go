package store

import (
	"sync"
	"time"

	"github.com/HomericIntelligence/atlas/internal/catalog"
	"github.com/HomericIntelligence/atlas/internal/tailscale"
)

// Cache is a thread-safe in-memory store for dashboard state.
// It is extended with new fields as Atlas milestones add data sources.
type Cache struct {
	mu       sync.RWMutex
	devices  []tailscale.Device // added in #158
	probes   []catalog.ProbeResult
	probesAt time.Time
	// probes extended in #159
}

// NewCache returns an empty Cache.
func NewCache() *Cache { return &Cache{} }

// SetProbes replaces the stored probe results and records the timestamp.
func (c *Cache) SetProbes(p []catalog.ProbeResult) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.probes = p
	c.probesAt = time.Now()
}

// GetProbes returns a copy of the stored probe results.
func (c *Cache) GetProbes() []catalog.ProbeResult {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]catalog.ProbeResult, len(c.probes))
	copy(out, c.probes)
	return out
}

// ProbesAge returns the time elapsed since the last SetProbes call.
// Returns a zero duration if SetProbes has never been called.
func (c *Cache) ProbesAge() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if c.probesAt.IsZero() {
		return 0
	}
	return time.Since(c.probesAt)
}

// SetDevices replaces the cached Tailscale device list.
func (c *Cache) SetDevices(d []tailscale.Device) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]tailscale.Device, len(d))
	copy(cp, d)
	c.devices = cp
}

// GetDevices returns a snapshot of the cached Tailscale device list.
// The returned slice is a copy; mutations do not affect the cache.
func (c *Cache) GetDevices() []tailscale.Device {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.devices) == 0 {
		return nil
	}
	cp := make([]tailscale.Device, len(c.devices))
	copy(cp, c.devices)
	return cp
}
