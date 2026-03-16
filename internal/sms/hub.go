package sms

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub manages per-user WebSocket connections for SMS/notification delivery.
type Hub struct {
	mu      sync.RWMutex
	clients map[string][]*websocket.Conn // userID -> connections
}

func NewHub() *Hub {
	return &Hub{clients: make(map[string][]*websocket.Conn)}
}

func (h *Hub) Register(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.clients[userID] = append(h.clients[userID], conn)
}

func (h *Hub) Unregister(userID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.clients[userID]
	for i, c := range conns {
		if c == conn {
			h.clients[userID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *Hub) SendNotification(userID string, msg interface{}) {
	h.mu.RLock()
	conns := h.clients[userID]
	h.mu.RUnlock()
	for _, conn := range conns {
		conn.WriteJSON(msg) //nolint:errcheck
	}
}
