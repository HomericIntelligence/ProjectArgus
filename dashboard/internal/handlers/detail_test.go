package handlers_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/HomericIntelligence/atlas/internal/handlers"
	"github.com/HomericIntelligence/atlas/internal/store"
)

// newCacheWithAgents builds a *store.Cache pre-populated with agents and tasks.
func newCacheWithAgents(agents []store.AgentRecord, tasks []store.TaskRecord) *store.Cache {
	c := store.NewCache()
	if len(agents) > 0 {
		c.SetAgents(agents)
	}
	if len(tasks) > 0 {
		c.SetTasks(tasks)
	}
	return c
}

// newDetailRouter wires up an AgentDetail and TaskDetail route via chi.
func newDetailRouter(cache *store.Cache) http.Handler {
	h := handlers.NewHostsHandler(cache)
	r := chi.NewRouter()
	r.Get("/agents/{id}", h.AgentDetail)
	r.Get("/tasks/{id}", h.TaskDetail)
	return r
}

// TestAgentDetail_NotFound asserts that GET /agents/{unknown-id} returns 404.
func TestAgentDetail_NotFound(t *testing.T) {
	t.Parallel()

	cache := newCacheWithAgents(nil, nil)
	r := newDetailRouter(cache)

	req := httptest.NewRequest(http.MethodGet, "/agents/nonexistent-id", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("status = %d; want 404", rr.Code)
	}
}

// TestAgentDetail_Found asserts that GET /agents/{known-id} returns 200 and
// includes the agent's name in the response body.
func TestAgentDetail_Found(t *testing.T) {
	t.Parallel()

	agent := store.AgentRecord{
		ID:        "agent-abc-123",
		Name:      "hermes-worker-1",
		Host:      "apollo.local",
		Status:    "idle",
		UpdatedAt: time.Now(),
	}
	cache := newCacheWithAgents([]store.AgentRecord{agent}, nil)
	r := newDetailRouter(cache)

	req := httptest.NewRequest(http.MethodGet, "/agents/agent-abc-123", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "hermes-worker-1") {
		t.Errorf("response body does not contain agent name %q\nbody: %s", "hermes-worker-1", body)
	}
}

// TestTaskDetail_Found asserts that GET /tasks/{known-id} returns 200 and
// includes the task's subject in the response body.
func TestTaskDetail_Found(t *testing.T) {
	t.Parallel()

	task := store.TaskRecord{
		ID:         "task-xyz-456",
		TeamID:     "team-alpha",
		Subject:    "Analyse quarterly metrics",
		Status:     "in_progress",
		AssigneeID: "agent-abc-123",
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		UpdatedAt:  time.Now(),
	}
	cache := newCacheWithAgents(nil, []store.TaskRecord{task})
	r := newDetailRouter(cache)

	req := httptest.NewRequest(http.MethodGet, "/tasks/task-xyz-456", nil)
	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d; want 200", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "Analyse quarterly metrics") {
		t.Errorf("response body does not contain task subject\nbody: %s", body)
	}
}
