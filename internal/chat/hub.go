package chat

import (
	"sync"

	"github.com/gorilla/websocket"
)

// Hub relays E2E-encrypted blobs between participants.
// The server never decrypts message content.
type Hub struct {
	mu    sync.RWMutex
	rooms map[string][]*websocket.Conn // numberID -> connections
}

func NewHub() *Hub {
	return &Hub{rooms: make(map[string][]*websocket.Conn)}
}

func (h *Hub) Join(numberID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.rooms[numberID] = append(h.rooms[numberID], conn)
}

func (h *Hub) Leave(numberID string, conn *websocket.Conn) {
	h.mu.Lock()
	defer h.mu.Unlock()
	conns := h.rooms[numberID]
	for i, c := range conns {
		if c == conn {
			h.rooms[numberID] = append(conns[:i], conns[i+1:]...)
			break
		}
	}
}

func (h *Hub) Broadcast(numberID string, sender *websocket.Conn, msg []byte) {
	h.mu.RLock()
	conns := h.rooms[numberID]
	h.mu.RUnlock()
	for _, c := range conns {
		if c == sender {
			continue
		}
		c.WriteMessage(websocket.BinaryMessage, msg) //nolint:errcheck
	}
}
