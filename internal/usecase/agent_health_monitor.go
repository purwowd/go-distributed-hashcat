package usecase

import (
	"context"
	"sync"
	"time"

	"go-distributed-hashcat/internal/domain"
	"go-distributed-hashcat/internal/infrastructure"

	"github.com/google/uuid"
)

// AgentHealthMonitor handles agent health checking and status management
type AgentHealthMonitor interface {
	Start(ctx context.Context)
	Stop()
	RegisterAgent(agentID uuid.UUID)
	UnregisterAgent(agentID uuid.UUID)
	GetHealthStatus() HealthStatus
}

type HealthStatus struct {
	OnlineAgents      int       `json:"online_agents"`
	OfflineAgents     int       `json:"offline_agents"`
	RecentlyOffline   int       `json:"recently_offline"`
	LastHealthCheck   time.Time `json:"last_health_check"`
	HealthCheckErrors int       `json:"health_check_errors"`
}

type agentHealthMonitor struct {
	agentUsecase     AgentUsecase
	wsHub            WebSocketHub
	config           HealthConfig
	ticker           *time.Ticker
	done             chan struct{}
	registeredAgents sync.Map // agentID -> registration time
	mu               sync.RWMutex
	healthStatus     HealthStatus
}

type HealthConfig struct {
	CheckInterval       time.Duration `json:"check_interval"`  // How often to check (default: 1 minute)
	AgentTimeout        time.Duration `json:"agent_timeout"`   // When to mark offline (default: 3 minutes)
	HeartbeatGrace      time.Duration `json:"heartbeat_grace"` // Grace period for heartbeat (default: 30s)
	MaxConcurrentChecks int           `json:"max_concurrent"`  // Max concurrent health checks
}

type WebSocketHub interface {
	BroadcastAgentStatus(agentID string, status string, lastSeen string)
	BroadcastAgentSpeed(agentID string, speed int64)
}

func NewAgentHealthMonitor(
	agentUsecase AgentUsecase,
	wsHub WebSocketHub,
	config HealthConfig,
) AgentHealthMonitor {
	// Set ultra-fast real-time defaults for instant responsiveness
	if config.CheckInterval == 0 {
		config.CheckInterval = 1 * time.Second // ✅ Ultra-fast health checks every 1 second
	}
	if config.AgentTimeout == 0 {
		config.AgentTimeout = 5 * time.Second // ✅ Ultra-fast offline detection in 5 seconds
	}
	if config.HeartbeatGrace == 0 {
		config.HeartbeatGrace = 2 * time.Second // ✅ Very short grace period
	}
	if config.MaxConcurrentChecks == 0 {
		config.MaxConcurrentChecks = 20 // ✅ More concurrent checks
	}

	return &agentHealthMonitor{
		agentUsecase: agentUsecase,
		wsHub:        wsHub,
		config:       config,
		done:         make(chan struct{}),
	}
}

func (h *agentHealthMonitor) Start(ctx context.Context) {
	infrastructure.ServerLogger.Info("Starting Agent Health Monitor (check interval: %v, timeout: %v)",
		h.config.CheckInterval, h.config.AgentTimeout)

	h.ticker = time.NewTicker(h.config.CheckInterval)

	go h.healthCheckLoop(ctx)

	// Initial health check
	go h.performHealthCheck(ctx)
}

func (h *agentHealthMonitor) Stop() {
	if h.ticker != nil {
		h.ticker.Stop()
	}
	select {
	case <-h.done:
		// Channel already closed
	default:
		close(h.done)
	}
	infrastructure.ServerLogger.Info("Agent Health Monitor stopped")
}

func (h *agentHealthMonitor) RegisterAgent(agentID uuid.UUID) {
	h.registeredAgents.Store(agentID, time.Now())
	infrastructure.ServerLogger.Info("Registered agent for health monitoring: %s", agentID.String()[:8])
}

func (h *agentHealthMonitor) UnregisterAgent(agentID uuid.UUID) {
	h.registeredAgents.Delete(agentID)
	infrastructure.ServerLogger.Info("Unregistered agent from health monitoring: %s", agentID.String()[:8])
}

func (h *agentHealthMonitor) GetHealthStatus() HealthStatus {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.healthStatus
}

func (h *agentHealthMonitor) healthCheckLoop(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-h.done:
			return
		case <-h.ticker.C:
			go h.performHealthCheck(ctx)
		}
	}
}

func (h *agentHealthMonitor) performHealthCheck(ctx context.Context) {
	start := time.Now()
	infrastructure.ServerLogger.Debug("Starting health check...")

	agents, err := h.agentUsecase.GetAllAgents(ctx)
	if err != nil {
		infrastructure.ServerLogger.Error("Failed to get agents for health check: %v", err)
		h.updateHealthStatus(0, 0, 0, 1)
		return
	}

	onlineCount := 0
	offlineCount := 0
	recentlyOfflineCount := 0

	// Use semaphore to limit concurrent checks
	sem := make(chan struct{}, h.config.MaxConcurrentChecks)
	var wg sync.WaitGroup

	for _, agent := range agents {
		wg.Add(1)
		go func(agent domain.Agent) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			h.checkSingleAgent(ctx, &agent, &onlineCount, &offlineCount, &recentlyOfflineCount)
		}(agent)
	}

	wg.Wait()

	h.updateHealthStatus(onlineCount, offlineCount, recentlyOfflineCount, 0)

	duration := time.Since(start)
	infrastructure.ServerLogger.Debug("Health check completed in %v - Online: %d, Offline: %d, Recently Offline: %d",
		duration, onlineCount, offlineCount, recentlyOfflineCount)
}

func (h *agentHealthMonitor) checkSingleAgent(ctx context.Context, agent *domain.Agent, onlineCount, offlineCount, recentlyOfflineCount *int) {
	now := time.Now()
	timeSinceLastSeen := now.Sub(agent.LastSeen)

	// Kondisi baru: agent tanpa IPAddress selalu offline
	if agent.IPAddress == "" {
		infrastructure.ServerLogger.Warning("Agent %s (%s) has no IP address, forcing offline status",
			agent.Name, agent.ID.String()[:8])
		*offlineCount++
		// Update status ke offline jika belum
		if agent.Status != "offline" {
			if err := h.agentUsecase.UpdateAgentStatus(ctx, agent.ID, "offline"); err != nil {
				infrastructure.ServerLogger.Error("Failed to update agent %s status to offline: %v", agent.Name, err)
			} else if h.wsHub != nil {
				h.wsHub.BroadcastAgentStatus(agent.ID.String(), "offline", agent.LastSeen.Format(time.RFC3339))
			}
		}
		return
	}

	// Determine if agent should be considered offline
	shouldBeOffline := timeSinceLastSeen > h.config.AgentTimeout
	wasRecentlyOnline := timeSinceLastSeen > h.config.AgentTimeout &&
		timeSinceLastSeen < (h.config.AgentTimeout+5*time.Minute)

	currentlyOnline := agent.Status == "online" || agent.Status == "busy"

	// Update counters (thread-safe with atomic operations would be better)
	// For simplicity, using direct increment here
	if shouldBeOffline {
		*offlineCount++
		if wasRecentlyOnline {
			*recentlyOfflineCount++
		}
	} else {
		*onlineCount++
	}

	// Status change needed?
	if shouldBeOffline && currentlyOnline {
		infrastructure.ServerLogger.Warning("Agent %s (%s) timeout detected - last seen %v ago",
			agent.Name, agent.ID.String()[:8], timeSinceLastSeen)

		// Update status to offline
		if err := h.agentUsecase.UpdateAgentStatus(ctx, agent.ID, "offline"); err != nil {
			infrastructure.ServerLogger.Error("Failed to update agent %s status to offline: %v", agent.Name, err)
			return
		}

		// Broadcast status change via WebSocket
		if h.wsHub != nil {
			infrastructure.ServerLogger.Debug("Broadcasting agent %s status change to offline via WebSocket", agent.Name)
			h.wsHub.BroadcastAgentStatus(
				agent.ID.String(),
				"offline",
				agent.LastSeen.Format(time.RFC3339),
			)
		}

		infrastructure.ServerLogger.Info("Agent %s status updated to offline", agent.Name)
	} else if !shouldBeOffline && !currentlyOnline {
		// Handle online status change (agent came back online)
		infrastructure.ServerLogger.Info("Agent %s (%s) came back online - last seen %v ago",
			agent.Name, agent.ID.String()[:8], timeSinceLastSeen)

		// Update status to online
		if err := h.agentUsecase.UpdateAgentStatus(ctx, agent.ID, "online"); err != nil {
			infrastructure.ServerLogger.Error("Failed to update agent %s status to online: %v", agent.Name, err)
			return
		}

		// Broadcast status change via WebSocket
		if h.wsHub != nil {
			infrastructure.ServerLogger.Debug("Broadcasting agent %s status change to online via WebSocket", agent.Name)
			h.wsHub.BroadcastAgentStatus(
				agent.ID.String(),
				"online",
				agent.LastSeen.Format(time.RFC3339),
			)
		}

		infrastructure.ServerLogger.Info("Agent %s status updated to online", agent.Name)
	}

	// Optional: Auto-cleanup very old offline agents
	if timeSinceLastSeen > 24*time.Hour && agent.Status == "offline" {
		infrastructure.ServerLogger.Warning("Agent %s is offline for >24h, consider cleanup", agent.Name)
		// Could trigger cleanup logic here
	}
}

func (h *agentHealthMonitor) updateHealthStatus(online, offline, recentlyOffline, errors int) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.healthStatus = HealthStatus{
		OnlineAgents:      online,
		OfflineAgents:     offline,
		RecentlyOffline:   recentlyOffline,
		LastHealthCheck:   time.Now(),
		HealthCheckErrors: errors,
	}
}
