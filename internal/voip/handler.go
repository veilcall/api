package voip

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	vertoURL string
	secret   string
}

func NewHandler(vertoURL, secret string) *Handler {
	return &Handler{vertoURL: vertoURL, secret: secret}
}

func (h *Handler) IssueToken(c *gin.Context) {
	userID := c.GetString("user_id")
	creds, err := IssueVertoCredentials(userID, h.secret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue verto credentials"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"username":   creds.Username,
		"password":   creds.Password,
		"expires_at": creds.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		"verto_url":  "/ws/verto",
	})
}

// VertoProxy proxies WebSocket frames between browser and FreeSWITCH Verto.
func (h *Handler) VertoProxy(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		c.Status(http.StatusUnauthorized)
		return
	}

	clientConn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer clientConn.Close()

	fsURL, _ := url.Parse(h.vertoURL)
	fsConn, _, err := websocket.DefaultDialer.Dial(fsURL.String(), nil)
	if err != nil {
		clientConn.WriteMessage(websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseInternalServerErr, "freeswitch unavailable"))
		return
	}
	defer fsConn.Close()

	done := make(chan struct{})
	go proxy(clientConn, fsConn, done)
	go proxy(fsConn, clientConn, done)
	<-done
}

func proxy(src, dst *websocket.Conn, done chan struct{}) {
	defer func() {
		select {
		case done <- struct{}{}:
		default:
		}
	}()
	for {
		msgType, msg, err := src.ReadMessage()
		if err != nil {
			return
		}
		if err := dst.WriteMessage(msgType, msg); err != nil {
			return
		}
	}
}
