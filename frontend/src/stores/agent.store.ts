// Agent Store for State Management
import { apiService, type Agent } from '@/services/api.service'

interface AgentState {
    agents: Agent[]
    loading: boolean
    error: string | null
    lastUpdated: Date | null
}

class AgentStore {
    private state: AgentState = {
        agents: [],
        loading: false,
        error: null,
        lastUpdated: null
    }

    private listeners: Set<() => void> = new Set()

    // Get current state
    public getState(): AgentState {
        return { ...this.state }
    }

    // Subscribe to state changes
    public subscribe(listener: () => void): () => void {
        this.listeners.add(listener)
        return () => this.listeners.delete(listener)
    }

    // Notify all listeners of state changes
    private notify(): void {
        this.listeners.forEach(listener => listener())
    }

    // Update state and notify listeners
    private setState(updates: Partial<AgentState>): void {
        this.state = { ...this.state, ...updates }
        this.notify()
    }

    // Actions
    public actions = {
        // Fetch all agents
        fetchAgents: async (): Promise<void> => {
            this.setState({ loading: true, error: null })
            
            try {
                const agents = await apiService.getAgents()
                this.setState({ 
                    agents, 
                    loading: false, 
                    lastUpdated: new Date(),
                    error: null 
                })
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to fetch agents' 
                })
            }
        },

        // Fetch single agent
        fetchAgent: async (id: string): Promise<Agent | null> => {
            try {
                const agent = await apiService.getAgent(id)
                if (agent) {
                    // Update agent in the state
                    const updatedAgents = this.state.agents.map(a => 
                        a.id === agent.id ? agent : a
                    )
                    if (!updatedAgents.find(a => a.id === agent.id)) {
                        updatedAgents.push(agent)
                    }
                    this.setState({ agents: updatedAgents })
                }
                return agent
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to fetch agent' 
                })
                return null
            }
        },

        // Create new agent
        createAgent: async (agentData: Partial<Agent>): Promise<Agent | null> => {
            this.setState({ loading: true, error: null })
            
            try {
                const newAgent = await apiService.createAgent(agentData)
                if (newAgent) {
                    this.setState({ 
                        agents: [...this.state.agents, newAgent],
                        loading: false 
                    })
                }
                return newAgent
            } catch (error) {
                this.setState({ 
                    loading: false, 
                    error: error instanceof Error ? error.message : 'Failed to create agent' 
                })
                return null
            }
        },

        // Update agent
        updateAgent: async (id: string, agentData: Partial<Agent>): Promise<Agent | null> => {
            try {
                const updatedAgent = await apiService.updateAgent(id, agentData)
                if (updatedAgent) {
                    const updatedAgents = this.state.agents.map(agent => 
                        agent.id === id ? updatedAgent : agent
                    )
                    this.setState({ agents: updatedAgents })
                }
                return updatedAgent
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to update agent' 
                })
                return null
            }
        },

        // Delete agent
        deleteAgent: async (id: string): Promise<boolean> => {
            try {
                const success = await apiService.deleteAgent(id)
                if (success) {
                    const updatedAgents = this.state.agents.filter(agent => agent.id !== id)
                    this.setState({ agents: updatedAgents })
                }
                return success
            } catch (error) {
                this.setState({ 
                    error: error instanceof Error ? error.message : 'Failed to delete agent' 
                })
                return false
            }
        },

        // Clear error
        clearError: (): void => {
            this.setState({ error: null })
        },

        // ✅ Real-time agent status update (without API call)
        updateAgentStatus: (agentId: string, status: string, lastSeen?: string): void => {
            const updatedAgents = this.state.agents.map(agent => {
                if (agent.id === agentId) {
                    return {
                        ...agent,
                        status: status as Agent['status'],
                        last_seen: lastSeen || agent.last_seen,
                        updated_at: new Date().toISOString()
                    }
                }
                return agent
            })
            
            this.setState({ 
                agents: updatedAgents,
                lastUpdated: new Date()
            })
        },

        // ✅ Real-time agent update (general properties)
        updateAgentRealtime: (agentId: string, updates: Partial<Agent>): void => {
            const updatedAgents = this.state.agents.map(agent => {
                if (agent.id === agentId) {
                    return {
                        ...agent,
                        ...updates,
                        updated_at: new Date().toISOString()
                    }
                }
                return agent
            })
            
            this.setState({ 
                agents: updatedAgents,
                lastUpdated: new Date()
            })
        },

        // Reset store
        reset: (): void => {
            this.setState({
                agents: [],
                loading: false,
                error: null,
                lastUpdated: null
            })
        }
    }

    // Computed getters
    public getters = {
        // Get online agents
        getOnlineAgents: (): Agent[] => {
            return this.state.agents.filter(agent => agent.status === 'online')
        },

        // Get busy agents
        getBusyAgents: (): Agent[] => {
            return this.state.agents.filter(agent => agent.status === 'busy')
        },

        // Get offline agents
        getOfflineAgents: (): Agent[] => {
            return this.state.agents.filter(agent => agent.status === 'offline')
        },

        // Get agent by id
        getAgentById: (id: string): Agent | undefined => {
            return this.state.agents.find(agent => agent.id === id)
        },

        // Get agents by job id (removed current_job_id reference)
        getAgentsByJobId: (jobId: string): Agent[] => {
            // Note: This functionality would require backend API support for job assignments
            return []
        },

        // Get total count
        getTotalCount: (): number => {
            return this.state.agents.length
        },

        // Get status counts
        getStatusCounts: (): Record<string, number> => {
            return this.state.agents.reduce((counts, agent) => {
                counts[agent.status] = (counts[agent.status] || 0) + 1
                return counts
            }, {} as Record<string, number>)
        }
    }
}

// Export singleton instance
export const agentStore = new AgentStore()

// Export types
export type { AgentState } 
