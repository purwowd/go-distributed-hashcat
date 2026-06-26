package domain

import (
	"sort"
	"strings"
)

// NormalizeAgentHardware fills type/processor from legacy capabilities when needed.
func NormalizeAgentHardware(agent *Agent) {
	if agent == nil {
		return
	}
	if agent.ResourceType != "" || agent.Processor != "" {
		agent.ResourceType = strings.ToUpper(strings.TrimSpace(agent.ResourceType))
		return
	}

	caps := strings.TrimSpace(agent.Capabilities)
	if caps == "" {
		return
	}

	lower := strings.ToLower(caps)
	if lower == "cpu" || lower == "gpu" {
		agent.ResourceType = strings.ToUpper(lower)
		return
	}

	if idx := strings.Index(caps, ":"); idx > 0 {
		agent.ResourceType = strings.ToUpper(strings.TrimSpace(caps[:idx]))
		agent.Processor = strings.TrimSpace(caps[idx+1:])
	}
}

// PrepareAgentForResponse ensures API consumers always get type and processor fields.
func PrepareAgentForResponse(agent *Agent) {
	NormalizeAgentHardware(agent)
}

// PrepareAgentsForResponse normalizes hardware fields for a list of agents.
func PrepareAgentsForResponse(agents []Agent) {
	for i := range agents {
		NormalizeAgentHardware(&agents[i])
	}
}

// AgentSortRank returns 0 for real GPU workers, 1 for CPU/PoCL-class workers.
func AgentSortRank(agent *Agent) int {
	if agent == nil {
		return 1
	}
	NormalizeAgentHardware(agent)
	processor := strings.ToLower(agent.Processor)
	capabilities := strings.ToLower(agent.Capabilities)
	if strings.Contains(processor, "portable computing language") ||
		strings.Contains(processor, "pocl") ||
		strings.Contains(capabilities, "portable computing language") ||
		strings.Contains(capabilities, "pocl") {
		return 1
	}
	if AgentUsesGPU(agent) {
		return 0
	}
	return 1
}

// SortAgentsByPriority orders agents with real GPU workers first, then by speed DESC, created_at DESC, ID.
func SortAgentsByPriority(agents []Agent) {
	PrepareAgentsForResponse(agents)
	sort.SliceStable(agents, func(i, j int) bool {
		aRank := AgentSortRank(&agents[i])
		bRank := AgentSortRank(&agents[j])
		if aRank != bRank {
			return aRank < bRank
		}
		if agents[i].Speed != agents[j].Speed {
			return agents[i].Speed > agents[j].Speed
		}
		if !agents[i].CreatedAt.Equal(agents[j].CreatedAt) {
			return agents[i].CreatedAt.After(agents[j].CreatedAt)
		}
		return agents[i].ID.String() < agents[j].ID.String()
	})
}

// AgentUsesGPU reports whether the agent should be treated as a GPU worker.
func AgentUsesGPU(agent *Agent) bool {
	if agent == nil {
		return false
	}
	NormalizeAgentHardware(agent)
	if t := strings.ToUpper(agent.ResourceType); t == "GPU" {
		return true
	}
	if t := strings.ToUpper(agent.ResourceType); t == "CPU" {
		return false
	}

	caps := strings.ToLower(agent.Capabilities)
	return strings.Contains(caps, "gpu") ||
		strings.Contains(caps, "cuda") ||
		strings.Contains(caps, "opencl") ||
		strings.Contains(caps, "rtx") ||
		strings.Contains(caps, "gtx") ||
		strings.Contains(caps, "radeon")
}

// AgentHardwareText returns combined text for legacy consumers.
func AgentHardwareText(agent *Agent) string {
	if agent == nil {
		return ""
	}
	NormalizeAgentHardware(agent)
	if agent.Processor != "" {
		if agent.ResourceType != "" {
			return agent.ResourceType + ": " + agent.Processor
		}
		return agent.Processor
	}
	return agent.Capabilities
}
