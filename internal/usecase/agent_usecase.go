package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

func generateAgentKey() (string, error) {
	bytes := make([]byte, 4)
	if _, err := rand.Read(bytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	return hex.EncodeToString(bytes), nil
}

type AgentUsecase interface {
	RegisterAgent(ctx context.Context, req *domain.CreateAgentRequest) (*domain.Agent, error)
	GetAgent(ctx context.Context, id uuid.UUID) (*domain.Agent, error)
	GetAllAgents(ctx context.Context) ([]domain.Agent, error)
	UpdateAgentStatus(ctx context.Context, id uuid.UUID, status string) error
	UpdateAgentSpeed(ctx context.Context, id uuid.UUID, speed int64) error
	UpdateAgentSpeedWithStatus(ctx context.Context, id uuid.UUID, speed int64, status string) error
	UpdateAgentStatusOffline(ctx context.Context, id uuid.UUID) error
	DeleteAgent(ctx context.Context, id uuid.UUID) error
	GetAvailableAgent(ctx context.Context) (*domain.Agent, error)
	UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error
	UpdateAgentLastSeen(ctx context.Context, id uuid.UUID) error
	GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error)
	SetWebSocketHub(wsHub WebSocketHub)
	ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error
	GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error)
	CreateAgent(ctx context.Context, agent *domain.Agent) error
	UpdateAgent(ctx context.Context, agent *domain.Agent) error
	UpdateAgentData(ctx context.Context, agentKey string, ipAddress string, port int, capabilities string) error
	GenerateAgentKey(ctx context.Context, name, agentKey string) (*domain.Agent, error)
}

type agentUsecase struct {
	agentRepo domain.AgentRepository
	wsHub     WebSocketHub
}

func NewAgentUsecase(agentRepo domain.AgentRepository) AgentUsecase {
	return &agentUsecase{
		agentRepo: agentRepo,
	}
}

func (u *agentUsecase) SetWebSocketHub(wsHub WebSocketHub) {
	u.wsHub = wsHub
}

func (u *agentUsecase) RegisterAgent(ctx context.Context, req *domain.CreateAgentRequest) (*domain.Agent, error) {
	// ✅ Validation 1: Check if agent key exists in database
	if req.AgentKey == "" {
		return nil, fmt.Errorf("agent key is required")
	}

	// Check if agent key exists in database
	existingAgentByKey, err := u.agentRepo.GetByAgentKey(ctx, req.AgentKey)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return nil, fmt.Errorf("agent key '%s' not found in database. Please generate a valid agent key first", req.AgentKey)
		}
		return nil, fmt.Errorf("failed to validate agent key: %w", err)
	}

	// ✅ Validation 2: Check if agent name matches the agent key
	if existingAgentByKey.Name != req.Name {
		return nil, fmt.Errorf("agent name '%s' does not match the name associated with agent key '%s' (expected: '%s')",
			req.Name, req.AgentKey, existingAgentByKey.Name)
	}

	// ✅ Validation 3: Check if agent already exists with the same name
	existingAgentByName, err := u.agentRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, err
	}

	if existingAgentByName != nil {
		// Agent already exists, check if this is an update or duplicate
		if existingAgentByName.AgentKey == req.AgentKey {
			// Update existing agent with new data
			// ✅ Validate IP address before update
			if err := u.ValidateUniqueIPForAgentKey(ctx, req.AgentKey, req.IPAddress, req.Name); err != nil {
				return nil, err
			}

			// Set default port 8080 if port is empty or 0
			port := req.Port
			if port == 0 {
				port = 8080
			}

			existingAgentByName.IPAddress = req.IPAddress
			existingAgentByName.Port = port // Use processed port
			existingAgentByName.Capabilities = req.Capabilities
			existingAgentByName.Status = "offline"    // stay offline during update
			existingAgentByName.LastSeen = time.Now() // Update LastSeen to current time
			existingAgentByName.UpdatedAt = time.Now()

			if err := u.agentRepo.Update(ctx, existingAgentByName); err != nil {
				return nil, err
			}

			// Update last seen to ensure consistency
			if err := u.agentRepo.UpdateLastSeen(ctx, existingAgentByName.ID); err != nil {
				// Log error but don't fail the request
				log.Printf("Failed to update agent last seen: %v", err)
			}

			// Don't broadcast status update when creating/updating agent
			// Status should remain as set (offline) until agent actually connects
			return existingAgentByName, nil
		} else {
			// Same name but different agent key
			return nil, fmt.Errorf("agent name '%s' already exists with a different agent key", req.Name)
		}
	}

	// ✅ Validation 4: Check unique IP address for new agent
	if err := u.ValidateUniqueIPForAgentKey(ctx, req.AgentKey, req.IPAddress, req.Name); err != nil {
		return nil, err
	}

	// Set default port 8080 if port is empty or 0
	port := req.Port
	if port == 0 {
		port = 8080
	}

	// Create new agent
	agent := &domain.Agent{
		ID:           existingAgentByKey.ID, // Use ID from existing agent key
		Name:         req.Name,
		IPAddress:    req.IPAddress,
		Port:         port,      // Use processed port
		Status:       "offline", // default offline
		Capabilities: req.Capabilities,
		AgentKey:     req.AgentKey,
		CreatedAt:    existingAgentByKey.CreatedAt, // Use created_at from agent key
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	if err := u.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}

	// Don't broadcast status update when creating agent
	// Status should remain as set (offline) until agent actually connects
	return agent, nil
}

func (u *agentUsecase) GetAgent(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	return u.agentRepo.GetByID(ctx, id)
}

func (u *agentUsecase) GetAllAgents(ctx context.Context) ([]domain.Agent, error) {
	return u.agentRepo.GetAll(ctx)
}

func (u *agentUsecase) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status string) error {
	// Update status in database
	if err := u.agentRepo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	// Get updated agent info for WebSocket broadcast
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get agent info for WebSocket broadcast: %v", err)
		return nil // Don't fail the status update if broadcast fails
	}

	// Broadcast real-time status update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("Real-time agent status broadcast: %s -> %s", agent.Name, status)
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}

	return nil
}

func (u *agentUsecase) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	return u.agentRepo.Delete(ctx, id)
}

func (u *agentUsecase) GetAvailableAgent(ctx context.Context) (*domain.Agent, error) {
	agents, err := u.agentRepo.GetAll(ctx)
	if err != nil {
		return nil, err
	}
	for _, agent := range agents {
		if agent.Status == "online" {
			return &agent, nil
		}
	}
	return nil, fmt.Errorf("no available agents found")
}

func (u *agentUsecase) UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error {
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// If status is still offline, don't change to online automatically
	if agent.Status != "offline" {
		if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
			return err
		}
	} else {
		// Update last seen only, keep status offline
		if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
			return err
		}
	}

	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("Real-time agent heartbeat broadcast: %s -> %s (LastSeen: %s)",
			agent.Name, agent.Status, agent.LastSeen.Format(time.RFC3339))
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}
	return nil
}

func (u *agentUsecase) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID) error {
	// Update last seen in database
	if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
		return err
	}

	// Get updated agent info for WebSocket broadcast
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get agent info for WebSocket broadcast: %v", err)
		return nil // Don't fail the last seen update if broadcast fails
	}

	// Broadcast real-time last seen update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("Real-time agent last seen broadcast: %s -> %s", agent.Name, agent.LastSeen.Format(time.RFC3339))
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}

	return nil
}

func (u *agentUsecase) UpdateAgentSpeed(ctx context.Context, id uuid.UUID, speed int64) error {
	// Update speed in database
	if err := u.agentRepo.UpdateSpeed(ctx, id, speed); err != nil {
		return err
	}

	// Get updated agent info for WebSocket broadcast
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get agent info for WebSocket broadcast: %v", err)
		return nil // Don't fail the speed update if broadcast fails
	}

	// Broadcast real-time speed update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentSpeed(agent.ID.String(), agent.Speed)
		log.Printf("Real-time agent speed broadcast: %s -> %d H/s", agent.Name, speed)
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}

	return nil
}

// UpdateAgentSpeedWithStatus updates agent speed and status simultaneously with comprehensive logging
// This method is used for real-time monitoring and comprehensive agent state updates
func (u *agentUsecase) UpdateAgentSpeedWithStatus(ctx context.Context, id uuid.UUID, speed int64, status string) error {
	// Update speed and status in database using new repository method
	if err := u.agentRepo.UpdateSpeedWithStatus(ctx, id, speed, status); err != nil {
		log.Printf("❌ [REAL-TIME UPDATE FAILED] Agent %s: speed=%d H/s, status=%s, error=%v",
			id.String(), speed, status, err)
		return err
	}

	// Get updated agent info for WebSocket broadcast
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get agent info for WebSocket broadcast: %v", err)
		return nil // Don't fail the update if broadcast fails
	}

	// Broadcast real-time speed and status update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentSpeed(agent.ID.String(), agent.Speed)
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("[REAL-TIME BROADCAST] Agent %s: speed=%d H/s, status=%s",
			agent.Name, speed, status)
	} else {
		log.Printf("⚠️ Warning: WebSocket hub not available for real-time broadcast")
	}

	return nil
}

func (u *agentUsecase) UpdateAgentStatusOffline(ctx context.Context, id uuid.UUID) error {
	// Update status in database
	if err := u.agentRepo.UpdateStatus(ctx, id, "offline"); err != nil {
		return err
	}

	// Get updated agent info for WebSocket broadcast
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("⚠️ Warning: Failed to get agent info for WebSocket broadcast: %v", err)
		return nil // Don't fail the status update if broadcast fails
	}

	// Broadcast real-time status update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("Real-time agent status broadcast: %s -> %s (offline)", agent.Name, agent.Status)
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}

	return nil
}

func (u *agentUsecase) GetByAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error) {
	return u.agentRepo.GetByAgentKey(ctx, agentKey)
}

func (u *agentUsecase) ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error {
	if ipAddress == "" {
		return nil // Empty IP address doesn't need validation
	}

	// Check if IP address is already used by another agent
	if agentWithIP, err := u.agentRepo.GetByIPAddress(ctx, ipAddress); err == nil && agentWithIP != nil {
		// IP address already in use
		if agentWithIP.AgentKey != agentKey {
			// IP address used by agent with different agent key
			return fmt.Errorf("IP address %s is already used by agent '%s' with agent key '%s'",
				ipAddress, agentWithIP.Name, agentWithIP.AgentKey)
		}
		if agentWithIP.Name != agentName {
			// IP address used by agent with different name
			return fmt.Errorf("IP address %s is already used by agent '%s'", ipAddress, agentWithIP.Name)
		}
		// If agent key and name are the same, this is an update of the same agent
		return nil
	} else if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		// Other error besides "not found"
		return fmt.Errorf("failed to validate IP address uniqueness: %w", err)
	}

	// IP address is available
	return nil
}

func (u *agentUsecase) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	return u.agentRepo.CreateAgent(ctx, agent)
}

func (u *agentUsecase) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	return u.agentRepo.UpdateAgent(ctx, agent)
}

// UpdateAgentData updates only the data fields (ip_address, port, capabilities) without changing status
func (u *agentUsecase) UpdateAgentData(ctx context.Context, agentKey string, ipAddress string, port int, capabilities string) error {
	// Get agent by agent key
	agent, err := u.agentRepo.GetByAgentKey(ctx, agentKey)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return fmt.Errorf("AGENT_KEY_NOT_FOUND: agent key %s not found in database", agentKey)
		}
		return fmt.Errorf("failed to get agent by key %s: %w", agentKey, err)
	}

	log.Printf("Debug: Found agent: %+v", agent)

	// Validate IP address uniqueness if provided
	if ipAddress != "" {
		if err := u.ValidateUniqueIPForAgentKey(ctx, agentKey, ipAddress, agent.Name); err != nil {
			return err
		}
	}

	// Set default port 8080 if port is empty or 0
	if port == 0 {
		port = 8080
	}

	// Update only data fields, keep status unchanged
	agent.IPAddress = ipAddress
	agent.Port = port
	agent.Capabilities = capabilities
	agent.UpdatedAt = time.Now()
	// Note: Status remains unchanged (stays offline until agent binary runs)

	log.Printf("Debug: Updated agent data: IP=%s, Port=%d, Capabilities=%s, UpdatedAt=%v",
		agent.IPAddress, agent.Port, agent.Capabilities, agent.UpdatedAt)

	// Update in database
	if err := u.agentRepo.Update(ctx, agent); err != nil {
		log.Printf("Debug: Failed to update agent in database: %v", err)
		return fmt.Errorf("failed to update agent data: %w", err)
	}

	log.Printf("Debug: Agent updated successfully in database")

	// Broadcast real-time agent data update via WebSocket
	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
		log.Printf("Real-time agent data broadcast: %s (IP=%s, Port=%d, Capabilities=%s)",
			agent.Name, agent.IPAddress, agent.Port, agent.Capabilities)
	} else {
		log.Printf("Warning: WebSocket hub not available for real-time broadcast")
	}

	// Don't broadcast status update - status should remain offline
	// until agent binary actually starts running
	return nil
}

func (u *agentUsecase) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	return u.agentRepo.GetByNameAndIPForStartup(ctx, name, ip, port)
}

// GenerateAgentKey creates a new agent key entry in the database
func (u *agentUsecase) GenerateAgentKey(ctx context.Context, name, agentKey string) (*domain.Agent, error) {
	// Check if agent name already exists
	existingAgent, err := u.agentRepo.GetByName(ctx, name)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, fmt.Errorf("failed to check existing agent: %w", err)
	}

	if existingAgent != nil {
		return nil, fmt.Errorf("agent name '%s' already exists", name)
	}

	// Check if agent key already exists
	existingAgentByKey, err := u.agentRepo.GetByAgentKey(ctx, agentKey)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, fmt.Errorf("failed to check existing agent key: %w", err)
	}

	if existingAgentByKey != nil {
		return nil, fmt.Errorf("agent key '%s' already exists", agentKey)
	}

	// Create new agent entry with just the name and key
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         name,
		IPAddress:    "", // Will be filled when agent registers
		Port:         0,  // Will be filled when agent registers
		Status:       "offline",
		Capabilities: "", // Will be filled when agent registers
		AgentKey:     agentKey,
		LastSeen:     time.Now(), // Set LastSeen to current time like created_at and updated_at
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Debug logging
	log.Printf("Debug: Creating agent with LastSeen: %v", agent.LastSeen)
	log.Printf("Debug: LastSeen.IsZero(): %v", agent.LastSeen.IsZero())
	log.Printf("Debug: LastSeen.Format(time.RFC3339): %v", agent.LastSeen.Format(time.RFC3339))
	log.Printf("Debug: Agent struct: %+v", agent)

	// Save to database
	if err := u.agentRepo.Create(ctx, agent); err != nil {
		log.Printf("Debug: Failed to save agent: %v", err)
		return nil, fmt.Errorf("failed to save agent key: %w", err)
	}

	log.Printf("Debug: Agent saved successfully with ID: %s", agent.ID.String())

	// Verify the saved agent
	savedAgent, err := u.agentRepo.GetByID(ctx, agent.ID)
	if err != nil {
		log.Printf("Debug: Could not retrieve saved agent: %v", err)
	} else {
		log.Printf("Debug: Retrieved saved agent LastSeen: %v", savedAgent.LastSeen)
		log.Printf("Debug: Retrieved saved agent LastSeen.IsZero(): %v", savedAgent.LastSeen.IsZero())
	}

	return agent, nil
}
