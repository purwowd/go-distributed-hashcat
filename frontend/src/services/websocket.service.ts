// WebSocket Service for Real-time Updates
import { getConfig } from '@/config/build.config'

interface WebSocketMessage {
    type: 'job_progress' | 'job_status' | 'agent_status' | 'notification'
    data: any
    timestamp: string
}

interface JobProgressUpdate {
    job_id: string
    progress: number
    speed?: number
    eta?: string
    status?: string
}

interface AgentStatusUpdate {
    agent_id: string
    status: 'online' | 'offline' | 'busy'
    last_seen: string
}

interface NotificationMessage {
    type: 'success' | 'error' | 'info' | 'warning'
    title: string
    message: string
    duration?: number
}

type MessageHandler = (message: WebSocketMessage) => void

class WebSocketService {
    private ws: WebSocket | null = null
    private config = getConfig()
    private reconnectAttempts = 0
    private maxReconnectAttempts = 5
    private reconnectDelay = 1000
    private handlers = new Map<string, Set<MessageHandler>>()
    private isConnected = false
    private shouldReconnect = true
    
    constructor() {
        this.connect()
    }

    private getWebSocketUrl(): string {
        const baseUrl = this.config.apiBaseUrl
        const wsUrl = baseUrl
            .replace('http://', 'ws://')
            .replace('https://', 'wss://')
        return `${wsUrl}/ws`
    }

    public connect(): void {
        try {
            const wsUrl = this.getWebSocketUrl()
            // console.log('🔗 Connecting to WebSocket:', wsUrl)
            
            this.ws = new WebSocket(wsUrl)
            
            this.ws.onopen = () => {
                // console.log('✅ WebSocket connected')
                this.isConnected = true
                this.reconnectAttempts = 0
                this.emit('connection', { connected: true })
            }
            
            this.ws.onmessage = (event) => {
                try {
                    const message: WebSocketMessage = JSON.parse(event.data)
                    this.handleMessage(message)
                } catch (error) {
                    // console.error('❌ Failed to parse WebSocket message:', error)
                }
            }
            
            this.ws.onclose = (event) => {
                // console.log('🔌 WebSocket disconnected:', event.code, event.reason)
                this.isConnected = false
                this.emit('connection', { connected: false })
                
                if (this.shouldReconnect && this.reconnectAttempts < this.maxReconnectAttempts) {
                    this.scheduleReconnect()
                }
            }
            
            this.ws.onerror = (error) => {
                // console.error('❌ WebSocket error:', error)
                this.emit('error', { error })
            }
            
        } catch (error) {
            // console.error('❌ Failed to create WebSocket connection:', error)
        }
    }

    private scheduleReconnect(): void {
        this.reconnectAttempts++
        const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1)
        
        // console.log(`Scheduling reconnect attempt ${this.reconnectAttempts}/${this.maxReconnectAttempts} in ${delay}ms`)
        
        setTimeout(() => {
            if (this.shouldReconnect) {
                this.connect()
            }
        }, delay)
    }

    private handleMessage(message: WebSocketMessage): void {
        // console.log('📨 WebSocket message received:', message.type, message.data)
        
        // Emit to specific type handlers
        this.emit(message.type, message.data)
        
        // Emit to general message handlers
        this.emit('message', message)
    }

    private emit(type: string, data: any): void {
        const typeHandlers = this.handlers.get(type)
        if (typeHandlers) {
            typeHandlers.forEach(handler => {
                try {
                    handler({ type, data, timestamp: new Date().toISOString() } as WebSocketMessage)
                } catch (error) {
                    console.error('❌ Error in WebSocket handler:', error)
                }
            })
        }
    }

    public on(type: string, handler: MessageHandler): () => void {
        if (!this.handlers.has(type)) {
            this.handlers.set(type, new Set())
        }
        
        const typeHandlers = this.handlers.get(type)!
        typeHandlers.add(handler)
        
        // Return unsubscribe function
        return () => {
            typeHandlers.delete(handler)
            if (typeHandlers.size === 0) {
                this.handlers.delete(type)
            }
        }
    }

    public send(message: any): boolean {
        if (this.ws && this.isConnected) {
            try {
                this.ws.send(JSON.stringify(message))
                return true
            } catch (error) {
                console.error('❌ Failed to send WebSocket message:', error)
                return false
            }
        }
        console.warn('⚠️ WebSocket not connected, message not sent')
        return false
    }

    public disconnect(): void {
        this.shouldReconnect = false
        if (this.ws) {
            this.ws.close()
            this.ws = null
        }
        this.isConnected = false
    }

    public getConnectionStatus(): boolean {
        return this.isConnected
    }

    // Convenience methods for specific message types
    public onJobProgress(handler: (update: JobProgressUpdate) => void): () => void {
        return this.on('job_progress', (message) => handler(message.data))
    }

    public onJobStatus(handler: (update: any) => void): () => void {
        return this.on('job_status', (message) => handler(message.data))
    }

    public onAgentStatus(handler: (update: AgentStatusUpdate) => void): () => void {
        return this.on('agent_status', (message) => handler(message.data))
    }

    public onNotification(handler: (notification: NotificationMessage) => void): () => void {
        return this.on('notification', (message) => handler(message.data))
    }

    public onConnection(handler: (status: { connected: boolean }) => void): () => void {
        return this.on('connection', (message) => handler(message.data))
    }

    // Subscribe to specific job updates
    public subscribeToJob(jobId: string): boolean {
        return this.send({
            type: 'subscribe',
            resource: 'job',
            id: jobId
        })
    }

    public unsubscribeFromJob(jobId: string): boolean {
        return this.send({
            type: 'unsubscribe',
            resource: 'job',
            id: jobId
        })
    }

    // Subscribe to all job updates
    public subscribeToJobs(): boolean {
        return this.send({
            type: 'subscribe',
            resource: 'jobs'
        })
    }

    // Subscribe to agent updates
    public subscribeToAgents(): boolean {
        return this.send({
            type: 'subscribe',
            resource: 'agents'
        })
    }
}

// Export singleton instance
export const webSocketService = new WebSocketService()

// Export types for external use
export type { 
    WebSocketMessage, 
    JobProgressUpdate, 
    AgentStatusUpdate, 
    NotificationMessage 
} 
