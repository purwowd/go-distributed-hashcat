package usecase

import (
	"context"
	"fmt"

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
}

type agentUsecase struct {
	agentRepo domain.AgentRepository
}

func NewAgentUsecase(agentRepo domain.AgentRepository) AgentUsecase {
	return &agentUsecase{
		agentRepo: agentRepo,
	}
}

func (u *agentUsecase) RegisterAgent(ctx context.Context, req *domain.CreateAgentRequest) (*domain.Agent, error) {
	agent := &domain.Agent{
		ID:           uuid.New(),
		Name:         req.Name,
		IPAddress:    req.IPAddress,
		Port:         req.Port,
		Status:       "online",
		Capabilities: req.Capabilities,
	}

	if err := u.agentRepo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to create agent: %w", err)
	}

	return agent, nil
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
	if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
		return fmt.Errorf("failed to update agent heartbeat: %w", err)
	}
	return nil
}
