package poller

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/HomericIntelligence/atlas/internal/config"
	"github.com/HomericIntelligence/atlas/internal/store"
)

// agentAPIRecord is the camelCase JSON shape returned by Agamemnon's /v1/agents.
type agentAPIRecord struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Host      string    `json:"host"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// taskAPIRecord is the camelCase JSON shape returned by Agamemnon's /v1/tasks.
type taskAPIRecord struct {
	ID         string    `json:"id"`
	TeamID     string    `json:"teamId"`
	Subject    string    `json:"subject"`
	Status     string    `json:"status"`
	AssigneeID string    `json:"assigneeAgentId"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
}

// tasksAPIResponse is the envelope returned by Agamemnon's /v1/tasks.
type tasksAPIResponse struct {
	Tasks []taskAPIRecord `json:"tasks"`
}

// AgamemnonPoller polls the Agamemnon service for agent and task data.
type AgamemnonPoller struct {
	base
	cache *store.Cache
	url   string
}

// NewAgamemnonPoller constructs an AgamemnonPoller with a 3-second HTTP timeout.
func NewAgamemnonPoller(cfg *config.Config, cache *store.Cache) *AgamemnonPoller {
	return &AgamemnonPoller{
		base: base{
			name:   "agamemnon",
			client: &http.Client{Timeout: 3 * time.Second},
		},
		cache: cache,
		url:   cfg.AgamemnonURL,
	}
}

// Start runs the poller in a ticker loop until ctx is cancelled.
func (p *AgamemnonPoller) Start(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Fetch immediately on start.
	p.fetch(ctx)

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			p.fetch(ctx)
		}
	}
}

// fetch retrieves agents and tasks from Agamemnon and updates the cache.
// On any error it logs a warning and leaves the cache unchanged.
func (p *AgamemnonPoller) fetch(ctx context.Context) {
	// Fetch agents.
	var rawAgents []agentAPIRecord
	if err := p.getJSON(ctx, p.url+"/v1/agents", &rawAgents); err != nil {
		slog.Warn("agamemnon poller: failed to fetch agents", "err", err)
		return
	}

	agents := make([]store.AgentRecord, len(rawAgents))
	for i, a := range rawAgents {
		agents[i] = store.AgentRecord{
			ID:        a.ID,
			Name:      a.Name,
			Host:      a.Host,
			Status:    a.Status,
			UpdatedAt: a.UpdatedAt,
		}
	}
	p.cache.SetAgents(agents)

	// Fetch tasks.
	var envelope tasksAPIResponse
	if err := p.getJSON(ctx, p.url+"/v1/tasks", &envelope); err != nil {
		slog.Warn("agamemnon poller: failed to fetch tasks", "err", err)
		return
	}

	tasks := make([]store.TaskRecord, len(envelope.Tasks))
	for i, t := range envelope.Tasks {
		tasks[i] = store.TaskRecord{
			ID:         t.ID,
			TeamID:     t.TeamID,
			Subject:    t.Subject,
			Status:     t.Status,
			AssigneeID: t.AssigneeID,
			CreatedAt:  t.CreatedAt,
			UpdatedAt:  t.UpdatedAt,
		}
	}
	p.cache.SetTasks(tasks)
}
