package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"go-distributed-hashcat/internal/domain"

	"github.com/google/uuid"
)

type AgentKeyUsecase interface {
	GenerateAgentKey(ctx context.Context, req *domain.GenerateAgentKeyRequest) (*domain.GenerateAgentKeyResponse, error)
	ValidateAgentKey(ctx context.Context, agentKey string) (*domain.Agent, error)
	ListAgentKeys(ctx context.Context) ([]*domain.AgentKeyInfo, error)
	RevokeAgentKey(ctx context.Context, agentKey string) error
	DeleteAgentKey(ctx context.Context, agentKey string) error
}

type agentKeyUsecase struct {
	agentKeyRepo domain.AgentKeyRepository
	agentRepo    domain.AgentRepository
}

func NewAgentKeyUsecase(agentKeyRepo domain.AgentKeyRepository, agentRepo domain.AgentRepository) AgentKeyUsecase {
	return &agentKeyUsecase{
		agentKeyRepo: agentKeyRepo,
		agentRepo:    agentRepo,
	}
}

// GenerateAgentKey generates a new agent key
func (u *agentKeyUsecase) GenerateAgentKey(ctx context.Context, req *domain.GenerateAgentKeyRequest) (*domain.GenerateAgentKeyResponse, error) {
	// Generate random key
	agentKeyStr, err := generateRandomKey(32) // 32 bytes = 64 hex chars
	if err != nil {
		return nil, fmt.Errorf("failed to generate agent key: %w", err)
	}

	// Create agent key record in separate table
	now := time.Now()
	agentKey := &domain.AgentKey{
		ID:          uuid.New(),
		AgentKey:    agentKeyStr,
		Name:        req.Name,
		Description: req.Description,
		Status:      "active",
		CreatedAt:   now,
		ExpiresAt:   req.ExpiresAt,
		LastUsedAt:  nil,
		AgentID:     nil, // Will be set when agent registers
	}

	if err := u.agentKeyRepo.Create(ctx, agentKey); err != nil {
		return nil, fmt.Errorf("failed to store agent key: %w", err)
	}

	return &domain.GenerateAgentKeyResponse{
		AgentKey:    agentKeyStr,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   now,
		ExpiresAt:   req.ExpiresAt,
	}, nil
}

// ValidateAgentKey validates an agent key and returns the associated agent (if any)
func (u *agentKeyUsecase) ValidateAgentKey(ctx context.Context, agentKeyStr string) (*domain.Agent, error) {
	if agentKeyStr == "" {
		return nil, fmt.Errorf("agent key is required")
	}

	// Get agent key from agent_keys table
	agentKey, err := u.agentKeyRepo.GetByKey(ctx, agentKeyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid agent key")
	}

	// Check if key is expired
	if agentKey.ExpiresAt != nil && time.Now().After(*agentKey.ExpiresAt) {
		return nil, fmt.Errorf("agent key expired")
	}

	// Check if key is revoked
	if agentKey.Status == "revoked" {
		return nil, fmt.Errorf("agent key revoked")
	}

	// Update last used timestamp
	u.agentKeyRepo.UpdateLastUsed(ctx, agentKeyStr)

	// If key is linked to an agent, return the agent
	if agentKey.AgentID != nil {
		agent, err := u.agentRepo.GetByID(ctx, *agentKey.AgentID)
		if err != nil {
			return nil, fmt.Errorf("linked agent not found")
		}
		return agent, nil
	}

	// Key is valid but not yet linked to an agent
	// Return a temporary agent object with the key info
	return &domain.Agent{
		ID:        uuid.New(), // Temporary ID
		Name:      agentKey.Name,
		AgentKey:  agentKeyStr,
		Status:    "key_only", // Special status indicating key exists but agent not registered
		CreatedAt: agentKey.CreatedAt,
		UpdatedAt: agentKey.CreatedAt,
	}, nil
}

// ListAgentKeys returns all agent keys with their status
func (u *agentKeyUsecase) ListAgentKeys(ctx context.Context) ([]*domain.AgentKeyInfo, error) {
	agentKeys, err := u.agentKeyRepo.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get agent keys: %w", err)
	}

	var keyInfos []*domain.AgentKeyInfo
	now := time.Now()

	for _, agentKey := range agentKeys {
		status := agentKey.Status
		if agentKey.ExpiresAt != nil && now.After(*agentKey.ExpiresAt) {
			status = "expired"
		}

		keyInfo := &domain.AgentKeyInfo{
			AgentKey:    agentKey.AgentKey,
			Name:        agentKey.Name,
			Description: agentKey.Description,
			Status:      status,
			CreatedAt:   agentKey.CreatedAt,
			ExpiresAt:   agentKey.ExpiresAt,
			LastUsed:    agentKey.LastUsedAt,
			AgentID:     agentKey.AgentID,
		}

		keyInfos = append(keyInfos, keyInfo)
	}

	return keyInfos, nil
}

// RevokeAgentKey revokes an agent key
func (u *agentKeyUsecase) RevokeAgentKey(ctx context.Context, agentKeyStr string) error {
	// Check if key exists
	_, err := u.agentKeyRepo.GetByKey(ctx, agentKeyStr)
	if err != nil {
		return fmt.Errorf("agent key not found")
	}

	// Update key status to revoked
	if err := u.agentKeyRepo.UpdateStatus(ctx, agentKeyStr, "revoked"); err != nil {
		return fmt.Errorf("failed to revoke agent key: %w", err)
	}

	return nil
}

// DeleteAgentKey permanently deletes an agent key
func (u *agentKeyUsecase) DeleteAgentKey(ctx context.Context, agentKeyStr string) error {
	// Get the agent key to validate it exists and get its ID
	agentKey, err := u.agentKeyRepo.GetByKey(ctx, agentKeyStr)
	if err != nil {
		return fmt.Errorf("agent key not found")
	}

	// Check if key is currently linked to an active agent
	if agentKey.AgentID != nil {
		// Get the linked agent to check its status
		agent, err := u.agentRepo.GetByID(ctx, *agentKey.AgentID)
		if err == nil && agent.Status == "online" {
			return fmt.Errorf("cannot delete agent key: linked agent is currently online")
		}
	}

	// Delete the agent key permanently
	if err := u.agentKeyRepo.Delete(ctx, agentKey.ID); err != nil {
		return fmt.Errorf("failed to delete agent key: %w", err)
	}

	return nil
}

// generateRandomKey generates a random hex key of specified byte length
func generateRandomKey(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
