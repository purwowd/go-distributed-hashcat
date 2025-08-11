package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

// generateAgentKey creates a random 8-character hex string for agent authentication
func generateAgentKey() (string, error) {
	bytes := make([]byte, 4) // 4 bytes = 8 hex characters
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
	DeleteAgent(ctx context.Context, id uuid.UUID) error
	GetAvailableAgent(ctx context.Context) (*domain.Agent, error)
	UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error
	SetWebSocketHub(wsHub WebSocketHub) // ✅ Add method to interface
}

type agentUsecase struct {
	agentRepo domain.AgentRepository
	wsHub     WebSocketHub // ✅ Add WebSocket hub (interface defined in health monitor)
}

func NewAgentUsecase(agentRepo domain.AgentRepository) AgentUsecase {
	return &agentUsecase{
		agentRepo: agentRepo,
		wsHub:     nil, // Will be set later when available
	}
}

// ✅ NEW: Set WebSocket hub for real-time broadcasts
func (u *agentUsecase) SetWebSocketHub(wsHub WebSocketHub) {
	u.wsHub = wsHub
}

func (u *agentUsecase) RegisterAgent(ctx context.Context, req *domain.CreateAgentRequest) (*domain.Agent, error) {
	// First, check if agent with same name already exists
	existingAgent, err := u.agentRepo.GetByName(ctx, req.Name)
	if err != nil {
		// If agent not found, check if agent key was provided
		if err.Error() == "agent not found" {
			// If client provided a pre-generated key, check if it's already used by another agent
			if req.AgentKey != "" {
				// Check if agent key already exists in the database
				agents, err := u.agentRepo.GetAll(ctx)
				if err != nil {
					return nil, fmt.Errorf("failed to check existing agents: %w", err)
				}

				var foundAgent *domain.Agent
				for _, agent := range agents {
					if agent.AgentKey == req.AgentKey {
						foundAgent = &agent
						break
					}
				}

				// If agent key is not found in database, return error
				if foundAgent == nil {
					return nil, fmt.Errorf("agent key '%s' is not registered in database", req.AgentKey)
				}

				// If agent key is found but used by different agent name, return error
				if foundAgent.Name != req.Name {
					return nil, fmt.Errorf("agent key '%s' is already registered with agent name '%s'", req.AgentKey, foundAgent.Name)
				}
			}

			// If client provided a pre-generated key, use it; otherwise generate server-side
			agentKey := req.AgentKey
			if agentKey == "" {
				var genErr error
				agentKey, genErr = generateAgentKey()
				if genErr != nil {
					return nil, fmt.Errorf("failed to generate agent key: %w", genErr)
				}
			}

			agent := &domain.Agent{
				ID:           uuid.New(),
				Name:         req.Name,
				IPAddress:    "",        // Empty for key generation
				Port:         0,         // Empty for key generation
				Status:       "offline", // Set default status to offline
				Capabilities: "",
				AgentKey:     agentKey,
			}

			if err := u.agentRepo.Create(ctx, agent); err != nil {
				return nil, fmt.Errorf("failed to create agent key: %w", err)
			}

			// ✅ Broadcast new agent key generation via WebSocket
			if u.wsHub != nil {
				log.Printf("📡 Broadcasting new agent key %s generation via WebSocket", agent.Name)
				u.wsHub.BroadcastAgentStatus(
					agent.ID.String(),
					agent.Status,
					agent.LastSeen.Format("2006-01-02T15:04:05Z07:00"),
				)
			}

			return agent, nil
		}
		// Other database errors
		return nil, fmt.Errorf("failed to check existing agent: %w", err)
	}

	// Agent already exists, check if this is registration with key
	if req.AgentKey != "" {
		// Validate the agent key
		if existingAgent.AgentKey != req.AgentKey {
			return nil, fmt.Errorf("invalid agent key for agent '%s'", req.Name)
		}

		// Check if agent key is being used by a different agent name
		if existingAgent.Name != req.Name {
			return nil, fmt.Errorf("agent key '%s' is already registered with a different agent name '%s'", req.AgentKey, existingAgent.Name)
		}

		// Check if agent is already registered (has IP, port, and capabilities)
		// If ALL fields are filled with actual values, then the agent is already registered
		// Note: Port 0 means not set, so we check if Port is not 0
		if existingAgent.IPAddress != "" && existingAgent.Port != 0 && existingAgent.Capabilities != "" {
			return nil, &domain.AlreadyRegisteredAgentError{
				Name:         req.Name,
				IPAddress:    existingAgent.IPAddress,
				Port:         existingAgent.Port,
				Capabilities: existingAgent.Capabilities,
			}
		}

		// If we reach here, the agent exists but is not fully registered
		// We need to update it with the provided details

		// Apply default port if needed
		port := req.Port
		if req.IPAddress != "" && port == 0 {
			port = 8080 // Default port
		}

		// Update the agent with the provided details
		existingAgent.IPAddress = req.IPAddress
		existingAgent.Port = port
		existingAgent.Capabilities = req.Capabilities
		existingAgent.Status = "offline" // Reset status to offline when registering

		if err := u.agentRepo.Update(ctx, existingAgent); err != nil {
			return nil, fmt.Errorf("failed to update agent: %w", err)
		}

		// ✅ Broadcast agent registration via WebSocket
		if u.wsHub != nil {
			log.Printf("📡 Broadcasting agent %s registration via WebSocket", existingAgent.Name)
			u.wsHub.BroadcastAgentStatus(
				existingAgent.ID.String(),
				existingAgent.Status,
				existingAgent.LastSeen.Format("2006-01-02T15:04:05Z07:00"),
			)
		}

		return existingAgent, nil
	}

	// Agent already exists and no key provided - return duplicate error
	return nil, &domain.DuplicateAgentError{
		Name:      req.Name,
		IPAddress: req.IPAddress,
		Port:      req.Port,
	}
}

func (u *agentUsecase) GetAgent(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	agent, err := u.agentRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent: %w", err)
	}
	return agent, nil
}

func (u *agentUsecase) GetAllAgents(ctx context.Context) ([]domain.Agent, error) {
	agents, err := u.agentRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}
	return agents, nil
}

func (u *agentUsecase) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status string) error {
	if err := u.agentRepo.UpdateStatus(ctx, id, status); err != nil {
		return fmt.Errorf("failed to update agent status: %w", err)
	}
	return nil
}

func (u *agentUsecase) DeleteAgent(ctx context.Context, id uuid.UUID) error {
	if err := u.agentRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete agent: %w", err)
	}
	return nil
}

func (u *agentUsecase) GetAvailableAgent(ctx context.Context) (*domain.Agent, error) {
	agents, err := u.agentRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agents: %w", err)
	}

	for _, agent := range agents {
		if agent.Status == "online" {
			return &agent, nil
		}
	}

	return nil, fmt.Errorf("no available agents found")
}

func (u *agentUsecase) UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error {
	// Update last seen timestamp
	if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}

	// ✅ Get agent info and broadcast status update via WebSocket
	if u.wsHub != nil {
		agent, err := u.agentRepo.GetByID(ctx, id)
		if err == nil {
			log.Printf("📡 Broadcasting heartbeat for agent %s via WebSocket", agent.Name)
			u.wsHub.BroadcastAgentStatus(
				agent.ID.String(),
				agent.Status,
				agent.LastSeen.Format("2006-01-02T15:04:05Z07:00"), // RFC3339 format
			)
		}
	}

	return nil
}
