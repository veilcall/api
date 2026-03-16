package chat

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	hub *Hub
}

func NewHandler(hub *Hub) *Handler {
	return &Handler{hub: hub}
}

// ChatWS handles /ws/chat/:number_id
// Clients send binary blobs (E2E encrypted); server relays without inspection.
func (h *Handler) ChatWS(c *gin.Context) {
	numberID := c.Param("number_id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Join(numberID, conn)
	defer h.hub.Leave(numberID, conn)

	for {
		msgType, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}
		if msgType != websocket.BinaryMessage {
			conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseUnsupportedData, "binary only"))
			break
		}
		h.hub.Broadcast(numberID, conn, msg)
	}
}
