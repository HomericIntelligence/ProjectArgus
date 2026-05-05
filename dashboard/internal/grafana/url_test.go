package grafana

import (
	"strings"
	"testing"
)

func TestPanelURL(t *testing.T) {
	p := Panel{
		DashUID: "agent-health",
		Slug:    "agent-health",
		PanelID: 1,
		Title:   "Agent Status",
	}
	got := PanelURL("http://grafana:3000", p, "now-1h", "now")

	wantPrefix := "http://grafana:3000/d-solo/agent-health/agent-health"
	if !strings.HasPrefix(got, wantPrefix) {
		t.Errorf("PanelURL prefix = %q; want prefix %q", got, wantPrefix)
	}

	checks := []string{
		"panelId=1",
		"kiosk=tv",
		"theme=dark",
		"from=now-1h",
		"to=now",
	}
	for _, c := range checks {
		if !strings.Contains(got, c) {
			t.Errorf("PanelURL(%q) missing %q", got, c)
		}
	}
}

func TestPanelURL_AllQueryParams(t *testing.T) {
	p := Panel{
		DashUID: "task-throughput",
		Slug:    "task-throughput",
		PanelID: 2,
		Title:   "Task Rate",
	}
	got := PanelURL("http://localhost:3000", p, "now-7d", "now")
	want := "http://localhost:3000/d-solo/task-throughput/task-throughput?panelId=2&kiosk=tv&theme=dark&from=now-7d&to=now"
	if got != want {
		t.Errorf("PanelURL =\n  %q\nwant\n  %q", got, want)
	}
}
