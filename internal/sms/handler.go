package sms

import (
	"crypto/ed25519"
	"encoding/base64"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type Handler struct {
	svc           *Service
	hub           *Hub
	webhookPubKey ed25519.PublicKey
}

func NewHandler(svc *Service, hub *Hub, webhookPublicKeyBase64 string) (*Handler, error) {
	keyBytes, err := base64.StdEncoding.DecodeString(webhookPublicKeyBase64)
	if err != nil {
		return nil, err
	}
	return &Handler{svc: svc, hub: hub, webhookPubKey: ed25519.PublicKey(keyBytes)}, nil
}

func (h *Handler) SendSMS(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		FromNumberID string `json:"from_number_id" binding:"required"`
		ToE164       string `json:"to_e164" binding:"required"`
		Text         string `json:"text" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.svc.SendSMS(c.Request.Context(), userID, req.FromNumberID, req.ToE164, req.Text); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// TelnyxWebhook receives inbound SMS from Telnyx.
// Ed25519 signature is verified; body is never logged.
func (h *Handler) TelnyxWebhook(c *gin.Context) {
	sigHeader := c.GetHeader("telnyx-signature-ed25519")
	tsHeader := c.GetHeader("telnyx-timestamp")

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	// Verify Ed25519 signature: sign(timestamp + "|" + body)
	if !verifyTelnyxSignature(h.webhookPubKey, sigHeader, tsHeader, body) {
		c.Status(http.StatusForbidden)
		return
	}

	// Parse minimally — only what we need to route the SMS
	var payload struct {
		Data struct {
			EventType string `json:"event_type"`
			Payload   struct {
				From struct {
					PhoneNumber string `json:"phone_number"`
				} `json:"from"`
				To []struct {
					PhoneNumber string `json:"phone_number"`
				} `json:"to"`
				Text string `json:"text"`
			} `json:"payload"`
		} `json:"data"`
	}

	if err := parseJSON(body, &payload); err != nil {
		c.Status(http.StatusOK) // Accept but ignore malformed
		return
	}

	if payload.Data.EventType == "message.received" && len(payload.Data.Payload.To) > 0 {
		h.svc.HandleInbound(
			c.Request.Context(),
			payload.Data.Payload.From.PhoneNumber,
			payload.Data.Payload.To[0].PhoneNumber,
			payload.Data.Payload.Text,
		)
	}

	c.Status(http.StatusOK)
}

// NotifyWS handles /ws/notify — pushes SMS and expiry events to the client.
func (h *Handler) NotifyWS(c *gin.Context) {
	userID := c.GetString("user_id")
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	h.hub.Register(userID, conn)
	defer h.hub.Unregister(userID, conn)

	// Keep connection alive until client disconnects
	for {
		if _, _, err := conn.ReadMessage(); err != nil {
			break
		}
	}
}
