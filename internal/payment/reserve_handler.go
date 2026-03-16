package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ReserveHandler handles POST /numbers/reserve
// It belongs here because it creates a payment record and returns payment details.
type ReserveHandler struct {
	svc *Service
}

func NewReserveHandler(svc *Service) *ReserveHandler {
	return &ReserveHandler{svc: svc}
}

func (h *ReserveHandler) Reserve(c *gin.Context) {
	userID := c.GetString("user_id")
	var req struct {
		Country string `json:"country" binding:"required,oneof=US GB"`
		Plan    string `json:"plan" binding:"required,oneof=24h 7d 30d"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.svc.Reserve(c.Request.Context(), userID, req.Plan, req.Country)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "reservation failed"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{
		"payment_id":     result.PaymentID,
		"monero_address": result.MoneroAddress,
		"amount_xmr":     result.AmountXMR,
	})
}
