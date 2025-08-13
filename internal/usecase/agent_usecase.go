package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
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
	DeleteAgent(ctx context.Context, id uuid.UUID) error
	GetAvailableAgent(ctx context.Context) (*domain.Agent, error)
	UpdateAgentHeartbeat(ctx context.Context, id uuid.UUID) error
	SetWebSocketHub(wsHub WebSocketHub)
	ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error
	GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error)
    CreateAgent(ctx context.Context, agent *domain.Agent) error
    UpdateAgent(ctx context.Context, agent *domain.Agent) error
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
	existingAgent, err := u.agentRepo.GetByName(ctx, req.Name)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			// Cek IP unik
			if req.IPAddress != "" {
				if agentWithIP, err := u.agentRepo.GetByIPAddress(ctx, req.IPAddress); err == nil && agentWithIP != nil {
					return nil, fmt.Errorf("already exists IP address %s", req.IPAddress)
				} else if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
					return nil, err
				}
			}

			// Generate key jika kosong
			agentKey := req.AgentKey
			if agentKey == "" {
				agentKey, err = generateAgentKey()
				if err != nil {
					return nil, err
				}
			}

			agent := &domain.Agent{
				ID:           uuid.New(),
				Name:         req.Name,
				IPAddress:    req.IPAddress,
				Port:         req.Port,
				Status:       "offline", // default offline
				Capabilities: req.Capabilities,
				AgentKey:     agentKey,
				CreatedAt:    time.Now(),
				UpdatedAt:    time.Now(),
			}

			if err := u.agentRepo.Create(ctx, agent); err != nil {
				return nil, err
			}

			if u.wsHub != nil {
				u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, "")
			}
			return agent, nil
		}
		return nil, err
	}

	// Update existing agent → tetap offline
	if req.AgentKey != "" {
		if existingAgent.AgentKey != req.AgentKey {
			return nil, fmt.Errorf("invalid agent key for agent '%s'", req.Name)
		}
		if req.IPAddress != "" && req.IPAddress != existingAgent.IPAddress {
			if agentWithIP, err := u.agentRepo.GetByIPAddress(ctx, req.IPAddress); err == nil && agentWithIP != nil {
				return nil, fmt.Errorf("already exists IP address %s", req.IPAddress)
			} else if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
				return nil, err
			}
		}

		existingAgent.IPAddress = req.IPAddress
		existingAgent.Port = req.Port
		existingAgent.Capabilities = req.Capabilities
		existingAgent.Status = "offline" // tetap offline saat update
		existingAgent.UpdatedAt = time.Now()

		if err := u.agentRepo.Update(ctx, existingAgent); err != nil {
			return nil, err
		}

		if u.wsHub != nil {
			u.wsHub.BroadcastAgentStatus(existingAgent.ID.String(), existingAgent.Status, "")
		}
		return existingAgent, nil
	}

	return nil, &domain.DuplicateAgentError{
		Name:      req.Name,
		IPAddress: req.IPAddress,
		Port:      req.Port,
	}
}

func (u *agentUsecase) GetAgent(ctx context.Context, id uuid.UUID) (*domain.Agent, error) {
	return u.agentRepo.GetByID(ctx, id)
}

func (u *agentUsecase) GetAllAgents(ctx context.Context) ([]domain.Agent, error) {
	return u.agentRepo.GetAll(ctx)
}

func (u *agentUsecase) UpdateAgentStatus(ctx context.Context, id uuid.UUID, status string) error {
	return u.agentRepo.UpdateStatus(ctx, id, status)
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

    // Kalau status masih offline, jangan ubah jadi online otomatis
    if agent.Status != "offline" {
        if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
            return err
        }
    } else {
        // Update last seen saja, biarkan status tetap offline
        if err := u.agentRepo.UpdateLastSeen(ctx, id); err != nil {
            return err
        }
    }

    if u.wsHub != nil {
        u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, agent.LastSeen.Format(time.RFC3339))
    }
    return nil
}

func (u *agentUsecase) ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error {
	if ipAddress == "" {
		return nil
	}
	if agentWithIP, err := u.agentRepo.GetByIPAddress(ctx, ipAddress); err == nil && agentWithIP != nil {
		if agentWithIP.AgentKey != agentKey {
			return fmt.Errorf("IP address %s is already registered with a different agent key", ipAddress)
		}
		if agentWithIP.Name != agentName {
			return fmt.Errorf("IP address %s is already used by another agent name %s", ipAddress, agentWithIP.Name)
		}
	} else if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return err
	}
	return nil
}


func (u *agentUsecase) CreateAgent(ctx context.Context, agent *domain.Agent) error {
	return u.agentRepo.CreateAgent(ctx, agent)
}

func (u *agentUsecase) UpdateAgent(ctx context.Context, agent *domain.Agent) error {
	return u.agentRepo.UpdateAgent(ctx, agent)
}

func (u *agentUsecase) GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error) {
	return u.agentRepo.GetByNameAndIPForStartup(ctx, name, ip, port)
}