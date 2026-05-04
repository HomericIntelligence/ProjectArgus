package store

import (
	"sync"

	"github.com/HomericIntelligence/atlas/internal/tailscale"
)

// Cache is a concurrency-safe in-memory cache for dashboard state.
// It is designed to be extended with additional fields as new Atlas milestones
// add data sources.
type Cache struct {
	mu      sync.RWMutex
	devices []tailscale.Device
	// probes added in #159
}

// NewCache returns an empty Cache.
func NewCache() *Cache {
	return &Cache{}
}

// SetDevices replaces the cached Tailscale device list.
func (c *Cache) SetDevices(d []tailscale.Device) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Copy the slice so callers cannot mutate the cache through the original.
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
