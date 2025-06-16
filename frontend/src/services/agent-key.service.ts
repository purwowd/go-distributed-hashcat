import { apiService } from './api.service';

export interface GenerateAgentKeyRequest {
    name: string;
    description?: string;
    expires_at?: string;
}

export interface GenerateAgentKeyResponse {
    agent_key: string;
    name: string;
    description?: string;
    created_at: string;
    expires_at?: string;
}

export interface AgentKeyInfo {
    agent_key: string;
    name: string;
    description?: string;
    status: 'active' | 'expired' | 'revoked';
    created_at: string;
    expires_at?: string;
    last_used?: string;
    agent_id?: string;
}

export interface ListAgentKeysResponse {
    agent_keys: AgentKeyInfo[];
    count: number;
}

export class AgentKeyService {
    constructor() {
        // Use the singleton apiService instance
    }

    /**
     * Generate a new agent key
     */
    async generateAgentKey(request: GenerateAgentKeyRequest): Promise<GenerateAgentKeyResponse> {
        try {
            const response = await fetch(`${apiService.getBaseUrl()}/api/v1/agent-keys/generate`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(request)
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('Failed to generate agent key:', error);
            throw error;
        }
    }

    /**
     * List all agent keys
     */
    async listAgentKeys(): Promise<AgentKeyInfo[]> {
        try {
            const response = await fetch(`${apiService.getBaseUrl()}/api/v1/agent-keys/`, {
                method: 'GET',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data: ListAgentKeysResponse = await response.json();
            return data.agent_keys || [];
        } catch (error) {
            console.error('Failed to list agent keys:', error);
            throw error;
        }
    }

    /**
     * Revoke an agent key
     */
    async revokeAgentKey(agentKey: string): Promise<void> {
        try {
            const response = await fetch(`${apiService.getBaseUrl()}/api/v1/agent-keys/${agentKey}/revoke`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Failed to revoke agent key:', error);
            throw error;
        }
    }

    /**
     * Delete an agent key permanently
     */
    async deleteAgentKey(agentKey: string): Promise<void> {
        try {
            const response = await fetch(`${apiService.getBaseUrl()}/api/v1/agent-keys/${agentKey}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                }
            });

            if (!response.ok) {
                const errorData = await response.json().catch(() => ({}));
                throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
            }
        } catch (error) {
            console.error('Failed to delete agent key:', error);
            throw error;
        }
    }
} 
