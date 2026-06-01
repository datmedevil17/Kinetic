package realm

import (
	"encoding/json"
	"os"
	"strings"

	"github.com/datmedevil/kinetic/server/internal/database"
	"github.com/datmedevil/kinetic/server/internal/models"
	"github.com/datmedevil/kinetic/server/internal/services/session"
	"github.com/datmedevil/kinetic/server/internal/services/ws"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/rs/zerolog/log"
)

// Service handles all realm-related business logic:
//   - Validating share links
//   - Fetching realm + profile data from Supabase (via GORM)
//   - Adding players to the in-memory session
//
// It implements ws.RealmJoiner so it can be injected into the hub
// without a circular import.
type Service struct {
	manager    *session.Manager
	maxPlayers int
}

// NewService creates a realm Service.
func NewService(manager *session.Manager, maxPlayers int) *Service {
	return &Service{
		manager:    manager,
		maxPlayers: maxPlayers,
	}
}

// JoinRealm is called by the hub when a client emits "joinRealm".
//
// Flow:
//  1. Check capacity — reject if realm is full
//  2. Fetch realm from DB — get map_data, share_id, owner_id, only_owner
//  3. Validate access — owner always allowed; others need matching shareId
//  4. Fetch player's skin from profiles table
//  5. Create session if it doesn't exist yet
//  6. Kick the player from any previous session (logged in elsewhere)
//  7. Add player to session
//  8. Emit "joinedRealm" back to the client
//  9. Notify the room that a new player joined
func (s *Service) JoinRealm(hub *ws.Hub, client *ws.Client, payload ws.JoinRealmPayload) {
	rejectJoin := func(reason string) {
		client.SendEvent(ws.EventFailedToJoinRoom, map[string]string{"reason": reason})
		log.Warn().Str("uid", client.UID).Str("realm", payload.RealmID).Str("reason", reason).Msg("Join rejected")
	}

	// ── Step 1: Capacity check ──────────────────────────────────────────────
	if existingSession := s.manager.GetSession(payload.RealmID); existingSession != nil {
		if existingSession.GetPlayerCount() >= s.maxPlayers {
			rejectJoin("Space is full. Max 30 players.")
			return
		}
	}

	// ── Step 2: Fetch realm from DB ─────────────────────────────────────────
	var realm models.Realm
	doc, err := database.Databases.GetDocument(
		os.Getenv("APPWRITE_DATABASE_ID"),
		os.Getenv("APPWRITE_REALMS_COLLECTION_ID"),
		payload.RealmID,
	)

	if err != nil {
		rejectJoin("Realm not found.")
		return
	}

	if err := doc.Decode(&realm); err != nil {
		rejectJoin("Failed to decode realm data.")
		return
	}

	// ── Step 3: Access validation ───────────────────────────────────────────
	isOwner := strings.EqualFold(realm.OwnerID, client.UID)

	if !isOwner {
		if realm.OnlyOwner {
			rejectJoin("This realm is private right now. Come back later!")
			return
		}
		if realm.ShareID != payload.ShareID {
			rejectJoin("The share link has been changed.")
			return
		}
	}

	// ── Step 4: Fetch player profile (skin) ─────────────────────────────────
	var profile models.Profile
	skinDoc, err := database.Databases.GetDocument(
		os.Getenv("APPWRITE_DATABASE_ID"),
		os.Getenv("APPWRITE_PROFILES_COLLECTION_ID"),
		client.UID,
	)

	skin := "009" // default skin
	if err == nil {
		if err := skinDoc.Decode(&profile); err == nil && profile.Skin != "" {
			skin = profile.Skin
		}
	}

	// ── Step 5: Parse Map Data ───────────────────────────────────────────────
	var mapData models.MapData
	if err := json.Unmarshal([]byte(realm.MapData), &mapData); err != nil {
		rejectJoin("Failed to parse map data.")
		return
	}

	// ── Step 6: Create session if not yet running ────────────────────────────
	if s.manager.GetSession(payload.RealmID) == nil {
		s.manager.CreateSession(payload.RealmID, mapData)
	}

	// ── Step 6: Kick player from any existing session (multi-tab / re-login) ─
	if oldSession := s.manager.GetPlayerSession(client.UID); oldSession != nil {
		hub.KickPlayer(client.UID, "You have logged in from another location.")
	}

	// ── Step 7: Add player to session ───────────────────────────────────────
	username := utils.FormatEmailToName(client.Email)
	s.manager.AddPlayerToSession(client.SocketID, payload.RealmID, client.UID, username, skin)

	// ── Step 8: Confirm join to the client ──────────────────────────────────
	client.SendEvent(ws.EventJoinedRealm, map[string]any{})

	// ── Step 9: Notify everyone else in the spawn room ──────────────────────
	newSession := s.manager.GetSession(payload.RealmID)
	if newSession != nil {
		player, ok := newSession.GetPlayer(client.UID)
		if ok {
			hub.BroadcastPlayerJoined(newSession, player, client.SocketID)
		}
	}

	log.Info().
		Str("uid", client.UID).
		Str("realm", payload.RealmID).
		Str("username", username).
		Msg("Player joined realm")
}
