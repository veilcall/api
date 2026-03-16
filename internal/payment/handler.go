package payment

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) GetStatus(c *gin.Context) {
	userID := c.GetString("user_id")
	paymentID := c.Param("id")

	p, err := h.svc.GetStatus(c.Request.Context(), paymentID, userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "payment not found"})
		return
	}

	resp := gin.H{
		"id":      p.ID,
		"status":  p.Status,
		"plan":    p.Plan,
		"country": p.Country,
	}
	if p.NumberID != nil {
		resp["number_id"] = *p.NumberID
	}
	c.JSON(http.StatusOK, resp)
}
