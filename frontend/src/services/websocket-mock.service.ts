// Mock WebSocket Service for Demo Purposes
// This simulates real-time updates until backend WebSocket is implemented

import { webSocketService } from './websocket.service'

class MockWebSocketService {
    private intervalId: number | null = null
    private isActive = false
    
    public start(): void {
        if (this.isActive) return
        
        this.isActive = true
        // console.log('🎭 Starting mock WebSocket service for demo...')
        
        // Simulate connection after short delay
        setTimeout(() => {
            webSocketService['emit']('connection', { connected: true })
        }, 1000)
        
        // Start mock job progress updates
        this.startMockJobUpdates()
        
        // Start mock agent status updates
        this.startMockAgentUpdates()
        
        // Send mock notifications
        this.startMockNotifications()
    }
    
    public stop(): void {
        if (this.intervalId) {
            clearInterval(this.intervalId)
            this.intervalId = null
        }
        this.isActive = false
        // console.log('🎭 Stopped mock WebSocket service')
    }
    
    private startMockJobUpdates(): void {
        let progress = 0
        
        setInterval(() => {
            if (!this.isActive) return
            
            // Simulate job progress
            progress += Math.random() * 10
            if (progress > 100) progress = 100
            
            const mockJobUpdate = {
                job_id: 'mock-job-id',
                progress: Math.floor(progress),
                speed: Math.floor(Math.random() * 50000) + 10000,
                eta: new Date(Date.now() + Math.random() * 3600000).toISOString(),
                status: progress >= 100 ? 'completed' : 'running'
            }
            
            webSocketService['emit']('job_progress', mockJobUpdate)
            
            // Simulate job completion
            if (progress >= 100) {
                setTimeout(() => {
                    webSocketService['emit']('job_status', {
                        job_id: 'mock-job-id',
                        status: 'completed',
                        result: 'Password found: admin123'
                    })
                }, 500)
                progress = 0 // Reset for next cycle
            }
        }, 2000)
    }
    
    private startMockAgentUpdates(): void {
        const agentStatuses = ['online', 'busy', 'offline']
        
        setInterval(() => {
            if (!this.isActive) return
            
            const mockAgentUpdate = {
                agent_id: 'mock-agent-id',
                status: agentStatuses[Math.floor(Math.random() * agentStatuses.length)],
                last_seen: new Date().toISOString()
            }
            
            webSocketService['emit']('agent_status', mockAgentUpdate)
        }, 5000)
    }
    
    private startMockNotifications(): void {
        const notifications = [
            {
                type: 'success',
                title: 'Job Completed',
                message: 'Password cracking job completed successfully!'
            },
            {
                type: 'info',
                title: 'Agent Connected',
                message: 'New agent GPU-Server-02 has joined the network'
            },
            {
                type: 'warning',
                title: 'High CPU Usage',
                message: 'Agent GPU-Server-01 is running at 95% capacity'
            }
        ]
        
        setInterval(() => {
            if (!this.isActive) return
            
            const notification = notifications[Math.floor(Math.random() * notifications.length)]
            webSocketService['emit']('notification', notification)
        }, 15000) // Every 15 seconds
    }
}

// Export singleton
export const mockWebSocketService = new MockWebSocketService()

// Auto-start in development mode (DISABLED - using real backend now)
if (import.meta.env.MODE === 'development' && import.meta.env.VITE_USE_MOCK === 'true') {
    // console.log('🎭 Mock WebSocket service will auto-start in development mode')
    // Start after a short delay to allow proper initialization
    setTimeout(() => {
        mockWebSocketService.start()
    }, 3000)
} else {
    // console.log('🎭 Mock WebSocket service disabled - using real backend')
} 
