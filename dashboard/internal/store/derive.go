package store

import "github.com/HomericIntelligence/atlas/internal/catalog"

// HostServices returns all probe results for the given hostname.
func (c *Cache) HostServices(host string) []catalog.ProbeResult {
	all := c.GetProbes()
	out := make([]catalog.ProbeResult, 0)
	for _, r := range all {
		if r.Host == host {
			out = append(out, r)
		}
	}
	return out
}
