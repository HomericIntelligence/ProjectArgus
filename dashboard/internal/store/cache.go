package store

import (
	"sync"
	"time"

	"github.com/HomericIntelligence/atlas/internal/catalog"
	"github.com/HomericIntelligence/atlas/internal/tailscale"
)

// maxAgentEvents is the maximum number of raw events retained per agent.
const maxAgentEvents = 50

// Cache is a thread-safe in-memory store for dashboard state.
// It is extended with new fields as Atlas milestones add data sources.
type Cache struct {
	mu          sync.RWMutex
	devices     []tailscale.Device // added in #158
	probes      []catalog.ProbeResult
	probesAt    time.Time
	// probes extended in #159
	agents      []AgentRecord        // added in #161
	tasks       []TaskRecord         // added in #161
	natsStats   NATSStats            // added in #161
	agentEvents map[string][]RawEvent // added in #163: per-agent event history
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

// SetAgents replaces the cached Agamemnon agent list.
func (c *Cache) SetAgents(agents []AgentRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]AgentRecord, len(agents))
	copy(cp, agents)
	c.agents = cp
}

// GetAgents returns a snapshot of the cached agent list.
// The returned slice is a copy; mutations do not affect the cache.
func (c *Cache) GetAgents() []AgentRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.agents) == 0 {
		return nil
	}
	cp := make([]AgentRecord, len(c.agents))
	copy(cp, c.agents)
	return cp
}

// SetTasks replaces the cached Agamemnon task list.
func (c *Cache) SetTasks(tasks []TaskRecord) {
	c.mu.Lock()
	defer c.mu.Unlock()
	cp := make([]TaskRecord, len(tasks))
	copy(cp, tasks)
	c.tasks = cp
}

// GetTasks returns a snapshot of the cached task list.
// The returned slice is a copy; mutations do not affect the cache.
func (c *Cache) GetTasks() []TaskRecord {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.tasks) == 0 {
		return nil
	}
	cp := make([]TaskRecord, len(c.tasks))
	copy(cp, c.tasks)
	return cp
}

// SetNATSStats replaces the cached NATS statistics.
func (c *Cache) SetNATSStats(s NATSStats) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.natsStats = s
}

// GetNATSStats returns the cached NATS statistics.
func (c *Cache) GetNATSStats() NATSStats {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.natsStats
}

// SetAgentEvents replaces the full event history slice for agentID.
func (c *Cache) SetAgentEvents(agentID string, events []RawEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.agentEvents == nil {
		c.agentEvents = make(map[string][]RawEvent)
	}
	cp := make([]RawEvent, len(events))
	copy(cp, events)
	c.agentEvents[agentID] = cp
}

// GetAgentEvents returns a copy of the event history for agentID.
// Returns nil if no events have been recorded for the agent.
func (c *Cache) GetAgentEvents(agentID string) []RawEvent {
	c.mu.RLock()
	defer c.mu.RUnlock()
	evts := c.agentEvents[agentID]
	if len(evts) == 0 {
		return nil
	}
	cp := make([]RawEvent, len(evts))
	copy(cp, evts)
	return cp
}

// AppendAgentEvent appends e to the event history for agentID, capping at maxAgentEvents.
func (c *Cache) AppendAgentEvent(agentID string, e RawEvent) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.agentEvents == nil {
		c.agentEvents = make(map[string][]RawEvent)
	}
	evts := c.agentEvents[agentID]
	evts = append(evts, e)
	if len(evts) > maxAgentEvents {
		evts = evts[len(evts)-maxAgentEvents:]
	}
	c.agentEvents[agentID] = evts
}
