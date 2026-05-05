//go:build e2e

package e2e

import (
	"net/http"
	"os"
	"testing"
	"time"
)

func atlasURL() string {
	if u := os.Getenv("ATLAS_E2E_URL"); u != "" {
		return u
	}
	return "http://localhost:3002"
}

func TestHealthz(t *testing.T) {
	resp, err := http.Get(atlasURL() + "/healthz")
	if err != nil {
		t.Fatalf("GET /healthz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestReadyz(t *testing.T) {
	resp, err := http.Get(atlasURL() + "/readyz")
	if err != nil {
		t.Fatalf("GET /readyz: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestAPIVersion(t *testing.T) {
	resp, err := http.Get(atlasURL() + "/api/version")
	if err != nil {
		t.Fatalf("GET /api/version: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}

func TestMetrics(t *testing.T) {
	resp, err := http.Get(atlasURL() + "/metrics")
	if err != nil {
		t.Fatalf("GET /metrics: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if ct == "" {
		t.Fatal("Content-Type header missing")
	}
}

func TestSSEHeartbeat(t *testing.T) {
	client := &http.Client{Timeout: 25 * time.Second}
	req, _ := http.NewRequest("GET", atlasURL()+"/events?topics=agent", nil)
	req.Header.Set("Accept", "text/event-stream")
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("GET /events: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
}
