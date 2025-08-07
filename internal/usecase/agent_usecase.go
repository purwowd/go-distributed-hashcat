package usecase

import (
	"context"
	"fmt"
	"log"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

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
	_, err := u.agentRepo.GetByName(ctx, req.Name)
	if err != nil {
		// If agent not found, create new one
		if err.Error() == "agent not found" {
			agent := &domain.Agent{
				ID:           uuid.New(),
				Name:         req.Name,
				IPAddress:    req.IPAddress,
				Port:         req.Port,
				Status:       "offline", // Set default status to offline saat pendaftaran manual
				Capabilities: req.Capabilities,
			}

			if err := u.agentRepo.Create(ctx, agent); err != nil {
				return nil, fmt.Errorf("failed to create agent: %w", err)
			}

			// ✅ Broadcast new agent registration via WebSocket
			if u.wsHub != nil {
				log.Printf("📡 Broadcasting new agent %s registration via WebSocket", agent.Name)
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

	// Agent already exists, return custom error
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
