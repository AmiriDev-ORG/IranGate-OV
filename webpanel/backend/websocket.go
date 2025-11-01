package main
import (
	"log"
	"net/http"
	"sync"
	"time"
	"github.com/gorilla/websocket"
)
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
type WSEventType string
const (
	EventClientConnected       WSEventType = "client_connected"
	EventClientDisconnected    WSEventType = "client_disconnected"
	EventClientCreated         WSEventType = "client_created"
	EventClientRemoved         WSEventType = "client_removed"
	EventStatsUpdate           WSEventType = "stats_update"
	EventSystemResourcesUpdate WSEventType = "system_resources_update"
	EventServerRestart         WSEventType = "server_restart"
	EventOrphanedFound         WSEventType = "orphaned_found"
)
type WSMessage struct {
	Type      WSEventType `json:"type"`
	Timestamp time.Time   `json:"timestamp"`
	Data      interface{} `json:"data"`
}
type Hub struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan WSMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mutex      sync.RWMutex
}
var hub = &Hub{
	clients:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan WSMessage, 256),
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}
func (h *Hub) run() {
	for {
		select {
		case conn := <-h.register:
			h.mutex.Lock()
			h.clients[conn] = true
			h.mutex.Unlock()
			log.Printf("WebSocket client connected. Total: %d", len(h.clients))
		case conn := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[conn]; ok {
				delete(h.clients, conn)
				conn.Close()
			}
			h.mutex.Unlock()
			log.Printf("WebSocket client disconnected. Total: %d", len(h.clients))
		case message := <-h.broadcast:
			h.mutex.RLock()
			for conn := range h.clients {
				err := conn.WriteJSON(message)
				if err != nil {
					log.Printf("Error sending to client: %v", err)
					conn.Close()
					delete(h.clients, conn)
				}
			}
			h.mutex.RUnlock()
		}
	}
}
func BroadcastEvent(eventType WSEventType, data interface{}) {
	message := WSMessage{
		Type:      eventType,
		Timestamp: time.Now(),
		Data:      data,
	}
	select {
	case hub.broadcast <- message:
	default:
		log.Println("Broadcast channel full, message dropped")
	}
}
func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}
	hub.register <- conn
	welcome := WSMessage{
		Type:      "connected",
		Timestamp: time.Now(),
		Data: map[string]string{
			"message": "Connected to IranGate WebSocket",
		},
	}
	conn.WriteJSON(welcome)
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				hub.unregister <- conn
				return
			}
		}
	}()
	for {
		var msg map[string]interface{}
		err := conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			hub.unregister <- conn
			break
		}
		handleClientMessage(conn, msg)
	}
}
func handleClientMessage(conn *websocket.Conn, msg map[string]interface{}) {
	msgType, ok := msg["type"].(string)
	if !ok {
		return
	}
	switch msgType {
	case "ping":
		conn.WriteJSON(map[string]string{"type": "pong"})
	case "subscribe":
	}
}
func startStatsBroadcaster() {
	ticker := time.NewTicker(5 * time.Second)
	go func() {
		for range ticker.C {
			stats := map[string]interface{}{
				"timestamp":         time.Now(),
				"connected_clients": 0,
				"cpu_usage":         0.0,
				"memory_usage_mb":   0.0,
			}
			BroadcastEvent(EventStatsUpdate, stats)
		}
	}()
}