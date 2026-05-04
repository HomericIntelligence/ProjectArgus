package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/HomericIntelligence/atlas/internal/store"
)

// Hosts is the handler for GET /api/hosts.
// It returns a JSON array of HostView, one entry per distinct hostname in the
// probe cache. tailscale_ip and online are placeholder fields pending the
// merge of feat/issue-158-tailscale-source.
type Hosts struct {
	cache *store.Cache
}

// NewHosts constructs a Hosts handler backed by the given cache.
func NewHosts(cache *store.Cache) *Hosts {
	return &Hosts{cache: cache}
}

// ServeHTTP implements http.Handler.
func (h *Hosts) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	probes := h.cache.GetProbes()

	// Build a map of hostname → HostView, preserving insertion order via a
	// separate slice.
	order := make([]string, 0)
	views := make(map[string]*store.HostView)

	for _, p := range probes {
		if _, exists := views[p.Host]; !exists {
			order = append(order, p.Host)
			views[p.Host] = &store.HostView{
				Hostname:    p.Host,
				TailscaleIP: "", // wired after merging #158
				Online:      false,
			}
		}
		views[p.Host].Services = append(views[p.Host].Services, p)
	}

	out := make([]store.HostView, 0, len(order))
	for _, hostname := range order {
		out = append(out, *views[hostname])
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(out); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}
