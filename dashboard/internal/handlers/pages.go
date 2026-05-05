package handlers

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/HomericIntelligence/atlas/internal/store"
	"github.com/HomericIntelligence/atlas/web/templates"
)

// HostsHandler serves the /hosts page and the /partials/host/{name} fragment.
type HostsHandler struct {
	cache *store.Cache
}

// NewHostsHandler creates a HostsHandler backed by the given cache.
func NewHostsHandler(cache *store.Cache) *HostsHandler {
	return &HostsHandler{cache: cache}
}

// ServeHTTP renders the full hosts page.
func (h *HostsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	views := store.BuildHostViews(h.cache)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.HostsPage(views).Render(r.Context(), w) //nolint:errcheck
}

// Partial renders a single host card fragment for HTMX polling.
func (h *HostsHandler) Partial(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	views := store.BuildHostViews(h.cache)
	for _, v := range views {
		if v.Hostname == name {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			templates.HostCard(v).Render(r.Context(), w) //nolint:errcheck
			return
		}
	}
	http.NotFound(w, r)
}

// AgentsPage renders the full agents list page with optional filters.
func (h *HostsHandler) AgentsPage(w http.ResponseWriter, r *http.Request) {
	agents := h.cache.GetAgents()
	search := r.URL.Query().Get("search")
	statusF := r.URL.Query().Get("status")
	hostF := r.URL.Query().Get("host")
	filtered := filterAgents(agents, search, statusF, hostF)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.AgentsPage(filtered, search, statusF, hostF).Render(r.Context(), w) //nolint:errcheck
}

// AgentsTablePartial renders only the table rows for HTMX partial updates.
func (h *HostsHandler) AgentsTablePartial(w http.ResponseWriter, r *http.Request) {
	agents := h.cache.GetAgents()
	search := r.URL.Query().Get("search")
	statusF := r.URL.Query().Get("status")
	hostF := r.URL.Query().Get("host")
	filtered := filterAgents(agents, search, statusF, hostF)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	for _, a := range filtered {
		templates.AgentRow(a).Render(r.Context(), w) //nolint:errcheck
	}
}

// filterAgents applies case-insensitive substring search on Name+Host,
// exact status match, and exact host match.
func filterAgents(agents []store.AgentRecord, search, status, host string) []store.AgentRecord {
	if search == "" && status == "" && host == "" {
		return agents
	}
	searchLower := strings.ToLower(search)
	out := make([]store.AgentRecord, 0, len(agents))
	for _, a := range agents {
		if search != "" {
			nameLower := strings.ToLower(a.Name)
			hostLower := strings.ToLower(a.Host)
			if !strings.Contains(nameLower, searchLower) && !strings.Contains(hostLower, searchLower) {
				continue
			}
		}
		if status != "" && a.Status != status {
			continue
		}
		if host != "" && a.Host != host {
			continue
		}
		out = append(out, a)
	}
	return out
}
