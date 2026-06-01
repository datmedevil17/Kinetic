package realmhandler

import (
	"strings"

	"github.com/datmedevil/kinetic/server/internal/middleware"
	"github.com/datmedevil/kinetic/server/internal/services/session"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/gin-gonic/gin"
)

// Handler holds the session manager (all realm queries go through it).
type Handler struct {
	manager *session.Manager
}

// NewHandler creates a realm handler.
func NewHandler(manager *session.Manager) *Handler {
	return &Handler{manager: manager}
}

// GetPlayersInRoom returns all players currently in the given room.
//
// GET /api/v1/players-in-room?roomIndex=N
//
// The player must be authenticated (JWT) and must already be in a realm.
// We look up which realm they're in via the session manager — no query param needed
// because a player can only be in one realm at a time.
func (h *Handler) GetPlayersInRoom(c *gin.Context) {
	uid := middleware.GetUID(c)

	// Parse roomIndex from query string
	var roomIndex int
	if err := bindIntQuery(c, "roomIndex", &roomIndex); err != nil {
		utils.BadRequest(c, "roomIndex must be a valid integer")
		return
	}

	// Find the session this player is in
	sess := h.manager.GetPlayerSession(uid)
	if sess == nil {
		utils.BadRequest(c, "You are not currently in any realm")
		return
	}

	players := sess.GetPlayersInRoom(roomIndex)
	utils.OK(c, gin.H{"players": players})
}

// GetPlayerCounts returns the live player count for each realm ID in the query.
//
// GET /api/v1/player-counts?realmIds=uuid1,uuid2,uuid3
//
// Used by the frontend to show "X people in this realm" on the lobby page.
// No auth required on this route (it's called before joining).
func (h *Handler) GetPlayerCounts(c *gin.Context) {
	raw := c.Query("realmIds")
	if raw == "" {
		utils.BadRequest(c, "realmIds query parameter is required")
		return
	}

	// Split comma-separated UUIDs and trim whitespace
	realmIDs := splitTrim(raw, ",")
	if len(realmIDs) == 0 {
		utils.BadRequest(c, "No valid realm IDs provided")
		return
	}

	counts := h.manager.GetPlayerCounts(realmIDs)
	utils.OK(c, gin.H{"playerCounts": counts})
}

// ─────────────────────────────────────────
//  Small query-binding helpers
// ─────────────────────────────────────────

func bindIntQuery(c *gin.Context, key string, out *int) error {
	val := c.Query(key)
	if val == "" {
		return nil
	}
	_, err := parse(val, out)
	return err
}

func parse(val string, out *int) (int, error) {
	n := 0
	for _, ch := range val {
		if ch < '0' || ch > '9' {
			return 0, gin.Error{Err: nil, Type: gin.ErrorTypePublic}
		}
		n = n*10 + int(ch-'0')
	}
	*out = n
	return n, nil
}

func splitTrim(s, sep string) []string {
	parts := strings.Split(s, sep)
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			result = append(result, t)
		}
	}
	return result
}
