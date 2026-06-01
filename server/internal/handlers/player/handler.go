package playerhandler

import (
	"net/http"

	"github.com/datmedevil/kinetic/server/internal/services/ws"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handler holds hub reference for player admin actions.
type Handler struct {
	hub *ws.Hub
}

// NewHandler creates a player handler.
func NewHandler(hub *ws.Hub) *Handler {
	return &Handler{hub: hub}
}

// Health is a simple liveness probe.
// GET /health
// Used by load balancers and Docker to check if the server is up.
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"service": "kinetic-server",
	})
}

// KickPlayer allows kicking a player from their realm.
// POST /api/v1/admin/kick
// Body: { "uid": "...", "reason": "..." }
//
// Note: in production you'd protect this with an admin-only middleware.
func (h *Handler) KickPlayer(c *gin.Context) {
	var body struct {
		UID    string `json:"uid" binding:"required"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		utils.BadRequest(c, "uid is required")
		return
	}

	reason := body.Reason
	if reason == "" {
		reason = "You have been removed by an administrator."
	}

	h.hub.KickPlayer(body.UID, reason)
	utils.OK(c, gin.H{"kicked": body.UID})
}
