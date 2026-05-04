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

// BuildHostViews joins devices from the Tailscale cache with probe results
// to produce a HostView slice suitable for rendering the hosts page.
func BuildHostViews(c *Cache) []HostView {
	devices := c.GetDevices()
	views := make([]HostView, len(devices))
	for i, d := range devices {
		views[i] = HostView{
			Hostname:    d.Hostname,
			TailscaleIP: d.TailscaleIP,
			Online:      d.Online,
			Services:    c.HostServices(d.Hostname),
		}
	}
	return views
}
