package auth

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

func (h *Handler) Register(c *gin.Context) {
	userID, recoveryCode, err := h.svc.Register(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		return
	}
	// Recovery code shown only once — never stored in plaintext
	c.JSON(http.StatusCreated, gin.H{
		"user_id":       userID,
		"recovery_code": recoveryCode,
		"warning":       "save this recovery code — it will not be shown again",
	})
}

func (h *Handler) Login(c *gin.Context) {
	var req struct {
		RecoveryCode string `json:"recovery_code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "recovery_code required"})
		return
	}

	token, err := h.svc.Login(c.Request.Context(), req.RecoveryCode)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"session_token": token})
}

func (h *Handler) Logout(c *gin.Context) {
	token := c.GetHeader("X-Session-Token")
	if token != "" {
		h.svc.Logout(c.Request.Context(), token)
	}
	c.Status(http.StatusNoContent)
}
