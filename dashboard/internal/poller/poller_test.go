package poller

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/HomericIntelligence/atlas/internal/config"
	"github.com/HomericIntelligence/atlas/internal/store"
)

// makeConfig returns a minimal config pointing at the given base URL.
func makeConfig(agamemnonURL, natsMon string) *config.Config {
	return &config.Config{
		AgamemnonURL: agamemnonURL,
		NATSMonURL:   natsMon,
	}
}

// -------------------------------------------------------------------
// AgamemnonPoller tests
// -------------------------------------------------------------------

func TestAgamemnonPoller_FetchUpdatesCache(t *testing.T) {
	agentsJSON := `[{"id":"a1","name":"worker","host":"host1","status":"online","updatedAt":"2024-01-01T00:00:00Z"}]`
	tasksJSON := `{"tasks":[{"id":"t1","teamId":"team1","subject":"do-work","status":"pending","assigneeAgentId":"a1","createdAt":"2024-01-01T00:00:00Z","updatedAt":"2024-01-01T00:00:00Z"}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agents":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(agentsJSON))
		case "/v1/tasks":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(tasksJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cfg := makeConfig(srv.URL, "")
	p := NewAgamemnonPoller(cfg, cache)
	p.fetch(context.Background())

	agents := cache.GetAgents()
	if len(agents) != 1 {
		t.Fatalf("expected 1 agent, got %d", len(agents))
	}
	if agents[0].ID != "a1" {
		t.Errorf("expected agent ID a1, got %s", agents[0].ID)
	}
	if agents[0].Status != "online" {
		t.Errorf("expected status online, got %s", agents[0].Status)
	}

	tasks := cache.GetTasks()
	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != "t1" {
		t.Errorf("expected task ID t1, got %s", tasks[0].ID)
	}
	if tasks[0].TeamID != "team1" {
		t.Errorf("expected teamId team1, got %s", tasks[0].TeamID)
	}
}

func TestAgamemnonPoller_HTTP500_CacheNotUpdated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal server error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cache := store.NewCache()
	// Pre-populate with known data.
	cache.SetAgents([]store.AgentRecord{{ID: "existing", Name: "old", Status: "online"}})
	cache.SetTasks([]store.TaskRecord{{ID: "t-existing"}})

	cfg := makeConfig(srv.URL, "")
	p := NewAgamemnonPoller(cfg, cache)
	p.fetch(context.Background())

	// Cache must be unchanged.
	agents := cache.GetAgents()
	if len(agents) != 1 || agents[0].ID != "existing" {
		t.Errorf("expected cache to be unchanged after HTTP 500, got %+v", agents)
	}
	tasks := cache.GetTasks()
	if len(tasks) != 1 || tasks[0].ID != "t-existing" {
		t.Errorf("expected task cache to be unchanged after HTTP 500, got %+v", tasks)
	}
}

func TestAgamemnonPoller_AgentsOK_TasksFail_AgentsCacheUpdated(t *testing.T) {
	// Agents endpoint returns 200, tasks endpoint returns 500.
	// Agents should be updated; tasks should remain from prior state.
	agentsJSON := `[{"id":"a2","name":"scout","host":"host2","status":"offline","updatedAt":"2024-01-02T00:00:00Z"}]`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/agents":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(agentsJSON))
		case "/v1/tasks":
			http.Error(w, "error", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cache.SetTasks([]store.TaskRecord{{ID: "old-task"}})

	cfg := makeConfig(srv.URL, "")
	p := NewAgamemnonPoller(cfg, cache)
	p.fetch(context.Background())

	// Agents should be updated.
	agents := cache.GetAgents()
	if len(agents) != 1 || agents[0].ID != "a2" {
		t.Errorf("expected agent a2 in cache, got %+v", agents)
	}
	// Tasks should be unchanged.
	tasks := cache.GetTasks()
	if len(tasks) != 1 || tasks[0].ID != "old-task" {
		t.Errorf("expected tasks to remain unchanged, got %+v", tasks)
	}
}

func TestAgamemnonPoller_StartAndStop(t *testing.T) {
	agentsJSON := `[{"id":"a1","name":"worker","host":"host1","status":"online","updatedAt":"2024-01-01T00:00:00Z"}]`
	tasksJSON := `{"tasks":[]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/agents":
			_, _ = w.Write([]byte(agentsJSON))
		case "/v1/tasks":
			_, _ = w.Write([]byte(tasksJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cfg := makeConfig(srv.URL, "")
	p := NewAgamemnonPoller(cfg, cache)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		p.Start(ctx, 50*time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("poller did not stop after context cancellation")
	}
}

// -------------------------------------------------------------------
// NATSPoller tests
// -------------------------------------------------------------------

func TestNATSPoller_FetchUpdatesCache(t *testing.T) {
	varzJSON := `{"connections":5,"in_msgs":1000,"out_msgs":800}`
	jszJSON := `{"num_streams":3}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/varz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(varzJSON))
		case "/jsz":
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(jszJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)
	p.fetch(context.Background())

	stats := cache.GetNATSStats()
	if stats.Connections != 5 {
		t.Errorf("expected 5 connections, got %d", stats.Connections)
	}
	if stats.Streams != 3 {
		t.Errorf("expected 3 streams, got %d", stats.Streams)
	}
	if stats.InMsgs != 1000 {
		t.Errorf("expected 1000 in_msgs, got %d", stats.InMsgs)
	}
	if stats.OutMsgs != 800 {
		t.Errorf("expected 800 out_msgs, got %d", stats.OutMsgs)
	}
}

func TestNATSPoller_VarzHTTP500_CacheNotUpdated(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	cache := store.NewCache()
	// Pre-populate.
	cache.SetNATSStats(store.NATSStats{Connections: 99, Streams: 7, InMsgs: 500, OutMsgs: 400})

	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)
	p.fetch(context.Background())

	stats := cache.GetNATSStats()
	if stats.Connections != 99 {
		t.Errorf("expected cache unchanged after HTTP 500, got connections=%d", stats.Connections)
	}
}

func TestNATSPoller_JszHTTP500_CacheNotUpdated(t *testing.T) {
	varzJSON := `{"connections":5,"in_msgs":1000,"out_msgs":800}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/varz":
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(varzJSON))
		case "/jsz":
			http.Error(w, "error", http.StatusInternalServerError)
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	// Pre-populate.
	cache.SetNATSStats(store.NATSStats{Connections: 99, Streams: 7, InMsgs: 500, OutMsgs: 400})

	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)
	p.fetch(context.Background())

	stats := cache.GetNATSStats()
	if stats.Connections != 99 {
		t.Errorf("expected cache unchanged after /jsz HTTP 500, got connections=%d", stats.Connections)
	}
}

func TestNATSPoller_StartAndStop(t *testing.T) {
	varzJSON := `{"connections":2,"in_msgs":50,"out_msgs":30}`
	jszJSON := `{"num_streams":1}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/varz":
			_, _ = w.Write([]byte(varzJSON))
		case "/jsz":
			_, _ = w.Write([]byte(jszJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	done := make(chan struct{})
	go func() {
		p.Start(ctx, 50*time.Millisecond)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("nats poller did not stop after context cancellation")
	}
}

// TestNATSPoller_FetchDetail_StreamsAndConns verifies that fetchDetail correctly
// parses stream list from /jsz?detail=1 and connection list from /connz and
// updates the cache with the correct values.
func TestNATSPoller_FetchDetail_StreamsAndConns(t *testing.T) {
	varzJSON := `{"connections":2,"in_msgs":10,"out_msgs":8}`
	jszJSON := `{"num_streams":2}`
	jszDetailJSON := `{"streams":[{"config":{"name":"homeric-agents","subjects":["hi.agents.>"]},"state":{"messages":42,"bytes":1024,"num_consumers":3},"created":"2024-01-01T00:00:00Z"},{"config":{"name":"homeric-tasks","subjects":["hi.tasks.>","hi.tasks.done"]},"state":{"messages":7,"bytes":256,"num_consumers":1},"created":"2024-02-01T00:00:00Z"}]}`
	connzJSON := `{"connections":[{"name":"agamemnon","ip":"10.0.0.1","subscriptions":5,"in_msgs":100,"out_msgs":80,"uptime":"1h0m0s"},{"name":"nestor","ip":"10.0.0.2","subscriptions":2,"in_msgs":20,"out_msgs":15,"uptime":"30m0s"}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/varz":
			_, _ = w.Write([]byte(varzJSON))
		case "/jsz":
			if r.URL.RawQuery == "detail=1" {
				_, _ = w.Write([]byte(jszDetailJSON))
			} else {
				_, _ = w.Write([]byte(jszJSON))
			}
		case "/connz":
			_, _ = w.Write([]byte(connzJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)
	p.fetch(context.Background())

	// Verify streams.
	streams := cache.GetNATSStreams()
	if len(streams) != 2 {
		t.Fatalf("expected 2 streams, got %d", len(streams))
	}
	if streams[0].Name != "homeric-agents" {
		t.Errorf("expected first stream name 'homeric-agents', got %q", streams[0].Name)
	}
	if len(streams[0].Subjects) != 1 || streams[0].Subjects[0] != "hi.agents.>" {
		t.Errorf("unexpected subjects for homeric-agents: %v", streams[0].Subjects)
	}
	if streams[0].Messages != 42 {
		t.Errorf("expected 42 messages, got %d", streams[0].Messages)
	}
	if streams[0].Consumers != 3 {
		t.Errorf("expected 3 consumers, got %d", streams[0].Consumers)
	}
	if streams[1].Name != "homeric-tasks" {
		t.Errorf("expected second stream name 'homeric-tasks', got %q", streams[1].Name)
	}
	if len(streams[1].Subjects) != 2 {
		t.Errorf("expected 2 subjects for homeric-tasks, got %d", len(streams[1].Subjects))
	}

	// Verify connections.
	conns := cache.GetNATSConns()
	if len(conns) != 2 {
		t.Fatalf("expected 2 connections, got %d", len(conns))
	}
	if conns[0].Name != "agamemnon" {
		t.Errorf("expected first conn name 'agamemnon', got %q", conns[0].Name)
	}
	if conns[0].IP != "10.0.0.1" {
		t.Errorf("expected IP '10.0.0.1', got %q", conns[0].IP)
	}
	if conns[0].Subscriptions != 5 {
		t.Errorf("expected 5 subscriptions, got %d", conns[0].Subscriptions)
	}
	if conns[0].InMsgs != 100 {
		t.Errorf("expected 100 in_msgs, got %d", conns[0].InMsgs)
	}
	if conns[0].Uptime != "1h0m0s" {
		t.Errorf("expected uptime '1h0m0s', got %q", conns[0].Uptime)
	}
	if conns[1].Name != "nestor" {
		t.Errorf("expected second conn name 'nestor', got %q", conns[1].Name)
	}
}

// TestNATSPoller_FetchDetail_JszDetailError_LeavesStreamsCacheIntact verifies that
// when /jsz?detail=1 returns an error, the streams cache is left unchanged while
// the connections cache is still updated from /connz.
func TestNATSPoller_FetchDetail_JszDetailError_LeavesStreamsCacheIntact(t *testing.T) {
	varzJSON := `{"connections":1,"in_msgs":5,"out_msgs":3}`
	jszJSON := `{"num_streams":1}`
	connzJSON := `{"connections":[{"name":"hermes","ip":"10.0.0.3","subscriptions":1,"in_msgs":5,"out_msgs":3,"uptime":"5m0s"}]}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/varz":
			_, _ = w.Write([]byte(varzJSON))
		case "/jsz":
			if r.URL.RawQuery == "detail=1" {
				http.Error(w, "error", http.StatusInternalServerError)
			} else {
				_, _ = w.Write([]byte(jszJSON))
			}
		case "/connz":
			_, _ = w.Write([]byte(connzJSON))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	cache := store.NewCache()
	// Pre-populate streams with known data.
	cache.SetNATSStreams([]store.NATSStreamInfo{{Name: "prior-stream", Messages: 99}})

	cfg := makeConfig("", srv.URL)
	p := NewNATSPoller(cfg, cache)
	p.fetch(context.Background())

	// Streams cache must be unchanged after /jsz?detail=1 error.
	streams := cache.GetNATSStreams()
	if len(streams) != 1 || streams[0].Name != "prior-stream" {
		t.Errorf("expected streams cache unchanged, got %+v", streams)
	}

	// Connections cache must be updated from /connz.
	conns := cache.GetNATSConns()
	if len(conns) != 1 || conns[0].Name != "hermes" {
		t.Errorf("expected connections cache updated with hermes, got %+v", conns)
	}
}
