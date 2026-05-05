package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/HomericIntelligence/atlas/internal/mnemosyne"
	"github.com/HomericIntelligence/atlas/web/templates"
)

// MnemosyneSearch renders the skill list partial for HTMX search updates.
// GET /partials/mnemosyne/search?q=
func (h *HostsHandler) MnemosyneSearch(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	var skills []mnemosyne.Skill
	if h.mnemoReader != nil {
		skills, _ = h.mnemoReader.Skills() //nolint:errcheck
	}
	filtered := mnemosyne.Filter(skills, q)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.SkillList(filtered).Render(r.Context(), w) //nolint:errcheck
}

// MnemosyneSkillBody renders the markdown body of a skill as HTML.
// GET /partials/mnemosyne/skill/{name}
func (h *HostsHandler) MnemosyneSkillBody(w http.ResponseWriter, r *http.Request) {
	name := chi.URLParam(r, "name")
	if h.mnemoReader == nil {
		http.NotFound(w, r)
		return
	}
	skills, _ := h.mnemoReader.Skills() //nolint:errcheck
	for _, s := range skills {
		if s.Name == name {
			html, err := mnemosyne.RenderMarkdown(s.Body)
			if err != nil {
				http.Error(w, "render error", http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			// Safe: goldmark renders with Unsafe=false (raw HTML stripped)
			fmt.Fprint(w, html) //nolint:errcheck
			return
		}
	}
	http.NotFound(w, r)
}
