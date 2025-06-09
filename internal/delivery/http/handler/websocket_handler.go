package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow connections from localhost in development
		origin := r.Header.Get("Origin")
		return origin == "http://localhost:3000" || origin == "http://localhost:5173" || origin == ""
	},
}

type WebSocketMessage struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp string      `json:"timestamp"`
}

type WebSocketClient struct {
	conn *websocket.Conn
	send chan WebSocketMessage
	hub  *WebSocketHub
	id   string
}

type WebSocketHub struct {
	clients    map[*WebSocketClient]bool
	broadcast  chan WebSocketMessage
	register   chan *WebSocketClient
	unregister chan *WebSocketClient
	mutex      sync.RWMutex
}

var Hub = &WebSocketHub{
	clients:    make(map[*WebSocketClient]bool),
	broadcast:  make(chan WebSocketMessage),
	register:   make(chan *WebSocketClient),
	unregister: make(chan *WebSocketClient),
}

func (h *WebSocketHub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected: %s", client.id)

			// Send welcome message
			welcome := WebSocketMessage{
				Type:      "connection",
				Data:      map[string]interface{}{"connected": true, "message": "Connected to Hashcat WebSocket"},
				Timestamp: time.Now().UTC().Format(time.RFC3339),
			}
			select {
			case client.send <- welcome:
			default:
				close(client.send)
				delete(h.clients, client)
			}

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)
				log.Printf("WebSocket client disconnected: %s", client.id)
			}
			h.mutex.Unlock()

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

func (h *WebSocketHub) BroadcastJobProgress(jobID string, progress float64, speed int64, eta string, status string) {
	message := WebSocketMessage{
		Type: "job_progress",
		Data: map[string]interface{}{
			"job_id":   jobID,
			"progress": progress,
			"speed":    speed,
			"eta":      eta,
			"status":   status,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Failed to broadcast job progress - channel full")
	}
}

func (h *WebSocketHub) BroadcastJobStatus(jobID string, status string, result string) {
	message := WebSocketMessage{
		Type: "job_status",
		Data: map[string]interface{}{
			"job_id": jobID,
			"status": status,
			"result": result,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Failed to broadcast job status - channel full")
	}
}

func (h *WebSocketHub) BroadcastAgentStatus(agentID string, status string, lastSeen string) {
	message := WebSocketMessage{
		Type: "agent_status",
		Data: map[string]interface{}{
			"agent_id":  agentID,
			"status":    status,
			"last_seen": lastSeen,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	select {
	case h.broadcast <- message:
	default:
		log.Println("Failed to broadcast agent status - channel full")
	}
}

func (c *WebSocketClient) readPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle incoming messages (subscription requests, etc.)
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msgType, ok := msg["type"].(string); ok {
				log.Printf("Received WebSocket message type: %s", msgType)
			}
		}
	}
}

func (c *WebSocketClient) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			if err := c.conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error: %v", err)
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

type WebSocketHandler struct{}

func NewWebSocketHandler() *WebSocketHandler {
	return &WebSocketHandler{}
}

func (h *WebSocketHandler) HandleWebSocket(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to upgrade to WebSocket"})
		return
	}

	clientID := c.ClientIP() + "-" + time.Now().Format("20060102150405")
	client := &WebSocketClient{
		conn: conn,
		send: make(chan WebSocketMessage, 256),
		hub:  Hub,
		id:   clientID,
	}

	client.hub.register <- client

	go client.writePump()
	go client.readPump()
}

// Initialize the WebSocket hub
func init() {
	go Hub.Run()
}
