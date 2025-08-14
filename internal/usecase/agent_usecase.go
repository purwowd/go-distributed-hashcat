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
	UpdateAgentLastSeen(ctx context.Context, id uuid.UUID) error
	SetWebSocketHub(wsHub WebSocketHub)
	ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error
	GetByNameAndIP(ctx context.Context, name, ip string, port int) (*domain.Agent, error)
	CreateAgent(ctx context.Context, agent *domain.Agent) error
	UpdateAgent(ctx context.Context, agent *domain.Agent) error
	GenerateAgentKey(ctx context.Context, name string) (*domain.Agent, error)
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
	// ✅ Validasi 1: Cek apakah agent key ada di database
	if req.AgentKey == "" {
		return nil, fmt.Errorf("agent key is required")
	}

	// Cek apakah agent key ada di database
	existingAgentByKey, err := u.agentRepo.GetByAgentKey(ctx, req.AgentKey)
	if err != nil {
		if errors.Is(err, domain.ErrAgentNotFound) {
			return nil, fmt.Errorf("agent key '%s' not found in database. Please generate a valid agent key first", req.AgentKey)
		}
		return nil, fmt.Errorf("failed to validate agent key: %w", err)
	}

	// ✅ Validasi 2: Cek apakah agent name sudah sesuai dengan agent key
	if existingAgentByKey.Name != req.Name {
		return nil, fmt.Errorf("agent name '%s' does not match the name associated with agent key '%s' (expected: '%s')",
			req.Name, req.AgentKey, existingAgentByKey.Name)
	}

	// ✅ Validasi 3: Cek apakah agent sudah ada dengan nama yang sama
	existingAgentByName, err := u.agentRepo.GetByName(ctx, req.Name)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, err
	}

	if existingAgentByName != nil {
		// Agent sudah ada, cek apakah ini update atau duplicate
		if existingAgentByName.AgentKey == req.AgentKey {
			// Update existing agent dengan data baru
			// ✅ Validasi IP address sebelum update
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
			existingAgentByName.Status = "offline" // tetap offline saat update
			existingAgentByName.UpdatedAt = time.Now()

			if err := u.agentRepo.Update(ctx, existingAgentByName); err != nil {
				return nil, err
			}

			if u.wsHub != nil {
				u.wsHub.BroadcastAgentStatus(existingAgentByName.ID.String(), existingAgentByName.Status, "")
			}
			return existingAgentByName, nil
		} else {
			// Nama sama tapi agent key berbeda
			return nil, fmt.Errorf("agent name '%s' already exists with a different agent key", req.Name)
		}
	}

	// ✅ Validasi 4: Cek IP address unik untuk agent baru
	if err := u.ValidateUniqueIPForAgentKey(ctx, req.AgentKey, req.IPAddress, req.Name); err != nil {
		return nil, err
	}

	// Set default port 8080 if port is empty or 0
	port := req.Port
	if port == 0 {
		port = 8080
	}

	// Buat agent baru
	agent := &domain.Agent{
		ID:           existingAgentByKey.ID, // Gunakan ID dari agent key yang sudah ada
		Name:         req.Name,
		IPAddress:    req.IPAddress,
		Port:         port,      // Use processed port
		Status:       "offline", // default offline
		Capabilities: req.Capabilities,
		AgentKey:     req.AgentKey,
		CreatedAt:    existingAgentByKey.CreatedAt, // Gunakan created_at dari agent key
		UpdatedAt:    time.Now(),
		LastSeen:     time.Now(),
	}

	if err := u.agentRepo.Create(ctx, agent); err != nil {
		return nil, err
	}

	if u.wsHub != nil {
		u.wsHub.BroadcastAgentStatus(agent.ID.String(), agent.Status, "")
	}
	return agent, nil
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

func (u *agentUsecase) UpdateAgentLastSeen(ctx context.Context, id uuid.UUID) error {
	return u.agentRepo.UpdateLastSeen(ctx, id)
}

func (u *agentUsecase) ValidateUniqueIPForAgentKey(ctx context.Context, agentKey, ipAddress, agentName string) error {
	if ipAddress == "" {
		return nil // IP address kosong tidak perlu divalidasi
	}

	// Cek apakah IP address sudah digunakan oleh agent lain
	if agentWithIP, err := u.agentRepo.GetByIPAddress(ctx, ipAddress); err == nil && agentWithIP != nil {
		// IP address sudah digunakan
		if agentWithIP.AgentKey != agentKey {
			// IP address digunakan oleh agent dengan agent key yang berbeda
			return fmt.Errorf("IP address %s is already used by agent '%s' with agent key '%s'",
				ipAddress, agentWithIP.Name, agentWithIP.AgentKey)
		}
		if agentWithIP.Name != agentName {
			// IP address digunakan oleh agent dengan nama yang berbeda
			return fmt.Errorf("IP address %s is already used by agent '%s'", ipAddress, agentWithIP.Name)
		}
		// Jika agent key dan nama sama, berarti ini update agent yang sama
		return nil
	} else if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		// Error lain selain "not found"
		return fmt.Errorf("failed to validate IP address uniqueness: %w", err)
	}

	// IP address tersedia
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

// GenerateAgentKey creates a new agent key entry in the database
func (u *agentUsecase) GenerateAgentKey(ctx context.Context, name string) (*domain.Agent, error) {
	// Check if agent name already exists
	existingAgent, err := u.agentRepo.GetByName(ctx, name)
	if err != nil && !errors.Is(err, domain.ErrAgentNotFound) {
		return nil, fmt.Errorf("failed to check existing agent: %w", err)
	}

	if existingAgent != nil {
		return nil, fmt.Errorf("agent name '%s' already exists", name)
	}

	// Generate unique agent key
	agentKey, err := generateAgentKey()
	if err != nil {
		return nil, fmt.Errorf("failed to generate agent key: %w", err)
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
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save to database
	if err := u.agentRepo.Create(ctx, agent); err != nil {
		return nil, fmt.Errorf("failed to save agent key: %w", err)
	}

	return agent, nil
}
