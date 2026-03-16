package number

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

func (h *Handler) ListNumbers(c *gin.Context) {
	userID := c.GetString("user_id")
	nums, err := h.svc.ListNumbers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list numbers"})
		return
	}

	type item struct {
		ID        string `json:"id"`
		Number    string `json:"number"`
		Country   string `json:"country"`
		Plan      string `json:"plan"`
		ExpiresAt string `json:"expires_at"`
	}
	out := make([]item, 0, len(nums))
	for _, n := range nums {
		out = append(out, item{
			ID:        n.ID,
			Number:    n.TelnyxNumber,
			Country:   n.Country,
			Plan:      n.Plan,
			ExpiresAt: n.ExpiresAt.Format("2006-01-02T15:04:05Z"),
		})
	}
	c.JSON(http.StatusOK, out)
}

func (h *Handler) ReleaseNumber(c *gin.Context) {
	userID := c.GetString("user_id")
	id := c.Param("id")

	if err := h.svc.ReleaseNumber(c.Request.Context(), id, userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
