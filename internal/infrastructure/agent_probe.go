package infrastructure

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

const (
	AgentHealthPath   = "/health"
	AgentProbeTimeout = 3 * time.Second
)

// ProbeAgentHealth checks whether an agent HTTP health endpoint is reachable (phase 1).
func ProbeAgentHealth(ip string, port int) bool {
	if ip == "" || port <= 0 {
		return false
	}

	url := fmt.Sprintf("http://%s:%d%s", ip, port, AgentHealthPath)
	ctx, cancel := context.WithTimeout(context.Background(), AgentProbeTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false
	}

	client := &http.Client{Timeout: AgentProbeTimeout}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}
