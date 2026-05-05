package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/HomericIntelligence/atlas/internal/handlers"
	"github.com/HomericIntelligence/atlas/internal/store"
)

// TestAgentsPage_Empty verifies that GET /agents returns 200 and contains
// "No agents" when the cache is empty.
func TestAgentsPage_Empty(t *testing.T) {
	t.Parallel()

	c := store.NewCache()
	h := handlers.NewHostsHandler(c)

	req := httptest.NewRequest(http.MethodGet, "/agents", nil)
	rr := httptest.NewRecorder()

	h.AgentsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "No agents") {
		t.Errorf("body does not contain 'No agents'; got: %s", body)
	}
}

// TestAgentsPage_Filter verifies that search=foo filters the table correctly:
// only agents whose Name or Host contains "foo" appear in the response.
func TestAgentsPage_Filter(t *testing.T) {
	t.Parallel()

	agents := []store.AgentRecord{
		{ID: "aaa11111", Name: "foobar", Host: "host1", Status: "idle", UpdatedAt: time.Now()},
		{ID: "bbb22222", Name: "nomatch", Host: "other", Status: "running", UpdatedAt: time.Now()},
		{ID: "ccc33333", Name: "baz", Host: "foohost", Status: "idle", UpdatedAt: time.Now()},
	}
	c := newCacheWithAgents(agents, nil)
	h := handlers.NewHostsHandler(c)

	req := httptest.NewRequest(http.MethodGet, "/agents?search=foo", nil)
	rr := httptest.NewRecorder()

	h.AgentsPage(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rr.Code)
	}
	body := rr.Body.String()
	if !strings.Contains(body, "foobar") {
		t.Errorf("body should contain 'foobar'; got: %s", body)
	}
	if !strings.Contains(body, "foohost") {
		t.Errorf("body should contain 'foohost'; got: %s", body)
	}
	if strings.Contains(body, "nomatch") {
		t.Errorf("body should NOT contain 'nomatch'; got: %s", body)
	}
}

// TestAgentsTablePartial verifies that GET /partials/agents/table returns
// only <tr> rows (no full HTML page) for each matching agent.
func TestAgentsTablePartial(t *testing.T) {
	t.Parallel()

	agents := []store.AgentRecord{
		{ID: "aaa11111", Name: "alpha", Host: "host-a", Status: "idle", UpdatedAt: time.Now()},
		{ID: "bbb22222", Name: "beta", Host: "host-b", Status: "running", UpdatedAt: time.Now()},
	}
	c := newCacheWithAgents(agents, nil)
	h := handlers.NewHostsHandler(c)

	req := httptest.NewRequest(http.MethodGet, "/partials/agents/table", nil)
	rr := httptest.NewRecorder()

	h.AgentsTablePartial(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d; want 200", rr.Code)
	}
	body := rr.Body.String()

	// Partial should contain <tr> elements
	if !strings.Contains(body, "<tr") {
		t.Errorf("body should contain <tr> elements; got: %s", body)
	}
	// Should NOT contain full HTML structure
	if strings.Contains(body, "<html") {
		t.Errorf("partial body should not contain <html> tag; got: %s", body)
	}
	// Both agents should appear
	if !strings.Contains(body, "alpha") {
		t.Errorf("body should contain 'alpha'; got: %s", body)
	}
	if !strings.Contains(body, "beta") {
		t.Errorf("body should contain 'beta'; got: %s", body)
	}

	// Status filter: only return idle agents
	req2 := httptest.NewRequest(http.MethodGet, "/partials/agents/table?status=idle", nil)
	rr2 := httptest.NewRecorder()
	h.AgentsTablePartial(rr2, req2)

	body2 := rr2.Body.String()
	if !strings.Contains(body2, "alpha") {
		t.Errorf("filtered body should contain 'alpha' (idle); got: %s", body2)
	}
	if strings.Contains(body2, "beta") {
		t.Errorf("filtered body should NOT contain 'beta' (running); got: %s", body2)
	}
}
