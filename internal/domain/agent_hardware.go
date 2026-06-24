package domain

import "strings"

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
