package handlers

import (
	"net/http"

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
