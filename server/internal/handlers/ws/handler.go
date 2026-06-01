package wshandler

import (
	"net/http"
	"strings"

	"github.com/datmedevil/kinetic/server/internal/services/ws"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
)

// upgrader configures the WebSocket upgrade.
// CheckOrigin returns true always — we handle origin security at the CORS layer.
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // CORS middleware handles origin validation
	},
}

// Handler holds dependencies for the WS handler.
type Handler struct {
	hub *ws.Hub
}

// NewHandler creates a new WS handler.
func NewHandler(hub *ws.Hub) *Handler {
	return &Handler{hub: hub}
}

// ServeWS upgrades an HTTP connection to WebSocket.
//
// Auth flow:
//   The client must provide the Supabase access token in one of two ways:
//   1. Authorization header: "Bearer <token>"  (preferred)
//   2. Query param: ?token=<token>             (fallback for browser WS)
//
// Why query param fallback?
//   The browser's native WebSocket API does NOT allow setting custom headers.
//   So the frontend passes the token as a query param on the WS URL.
//   Example: ws://localhost:8080/ws?token=eyJ...
//
// GET /ws
func (h *Handler) ServeWS(c *gin.Context) {
	// ── Extract token ─────────────────────────────────────────────────────
	token := extractToken(c.Request)
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing auth token"})
		return
	}

	// ── Validate with Supabase ────────────────────────────────────────────
	user, err := utils.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
		return
	}

	// ── Upgrade HTTP → WebSocket ──────────────────────────────────────────
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Error().Err(err).Str("uid", user.ID).Msg("WebSocket upgrade failed")
		return
	}

	// ── Create client and register with hub ───────────────────────────────
	socketID := uuid.New().String()
	client := ws.NewClient(h.hub, conn, user.ID, socketID, user.Email)

	h.hub.Register <- client

	log.Info().
		Str("uid", user.ID).
		Str("socketId", socketID).
		Msg("WebSocket connection established")

	// ── Start pumps in goroutines ─────────────────────────────────────────
	// WritePump runs in its own goroutine — it MUST be the only writer to conn.
	go client.WritePump()

	// ReadPump blocks until the connection closes.
	// When it returns, it sends client to hub.Unregister.
	client.ReadPump()
}

// extractToken reads the Bearer token from Authorization header,
// falling back to the "token" query parameter.
func extractToken(r *http.Request) string {
	// Try Authorization header first
	auth := r.Header.Get("Authorization")
	if auth != "" {
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
			return parts[1]
		}
	}

	// Fallback: ?token= query param (browser WebSocket limitation)
	return r.URL.Query().Get("token")
}
