package ws

import (
	"encoding/json"
	"sync"

	"github.com/datmedevil/kinetic/server/internal/services/session"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/rs/zerolog/log"
)

// InboundMessage carries a raw WebSocket message + its sender.
type InboundMessage struct {
	Client *Client
	Data   []byte
}

// Hub is the central event bus for all WebSocket connections.
//
// It runs one goroutine (Run) that serially processes:
//   - New connections (Register channel)
//   - Disconnections (Unregister channel)
//   - Incoming messages (Inbound channel)
//
// Serial processing means we don't need a mutex on the clients map —
// only the Run goroutine touches it.
type Hub struct {
	// clients maps socketID → *Client
	clients   map[string]*Client
	clientsMu sync.RWMutex // only needed for SendToSocket from outside Run

	Register   chan *Client
	Unregister chan *Client
	Inbound    chan InboundMessage

	// joiningInProgress prevents a player from double-joining concurrently
	joining   map[string]struct{}
	joiningMu sync.Mutex

	manager *session.Manager
	maxPlayers int
}

// GlobalHub is the singleton hub used across the app.
var GlobalHub *Hub

// NewHub creates a Hub and wires it to the session manager.
func NewHub(manager *session.Manager, maxPlayers int) *Hub {
	h := &Hub{
		clients:    make(map[string]*Client),
		Register:   make(chan *Client, 64),
		Unregister: make(chan *Client, 64),
		Inbound:    make(chan InboundMessage, 512),
		joining:    make(map[string]struct{}),
		manager:    manager,
		maxPlayers: maxPlayers,
	}
	GlobalHub = h
	return h
}

// Run starts the hub's event loop. Call as `go hub.Run()` in main.
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.clientsMu.Lock()
			h.clients[client.SocketID] = client
			h.clientsMu.Unlock()
			log.Info().Str("uid", client.UID).Str("socket", client.SocketID).Msg("Client registered")

		case client := <-h.Unregister:
			h.clientsMu.Lock()
			delete(h.clients, client.SocketID)
			h.clientsMu.Unlock()
			h.handleDisconnect(client)

		case msg := <-h.Inbound:
			h.dispatch(msg.Client, msg.Data)
		}
	}
}

// dispatch decodes the event envelope and routes to the right handler.
//
// Design rule:
//   Events that are PURE IN-MEMORY (move, chat, skin, teleport) run directly
//   on the hub's serial goroutine — no mutex contention, maximum throughput.
//
//   Events that do I/O (joinRealm → 2 DB queries) are spawned in their own
//   goroutine so they don't block the hub loop while waiting for Supabase.
//   All shared state those goroutines touch (session.Manager, hub.clients) is
//   already mutex-protected, so spawning them is safe.
func (h *Hub) dispatch(c *Client, data []byte) {
	var msg Message
	if err := json.Unmarshal(data, &msg); err != nil {
		log.Warn().Str("uid", c.UID).Msg("Invalid message format")
		return
	}

	switch msg.Event {
	case EventJoinRealm:
		// ← spawn goroutine: this does 2 network DB calls (~100ms each)
		//   Without the goroutine, ALL player moves/chats freeze during that time.
		go h.handleJoinRealm(c, msg.Payload)

	case EventMovePlayer:
		// ← stays on hub goroutine: pure in-memory (session map + broadcast)
		h.handleMovePlayer(c, msg.Payload)

	case EventTeleport:
		// ← stays on hub goroutine: pure in-memory
		h.handleTeleport(c, msg.Payload)

	case EventChangedSkin:
		// ← stays on hub goroutine: pure in-memory
		h.handleChangedSkin(c, msg.Payload)

	case EventSendMessage:
		// ← stays on hub goroutine: pure in-memory
		h.handleSendMessage(c, msg.Payload)

	case EventBoardUpdate:
		// ← stays on hub goroutine: pure in-memory
		h.handleBoardUpdate(c, msg.Payload)

	default:
		log.Debug().Str("event", msg.Event).Msg("Unknown event")
	}
}

// ──────────────────────────────────────────
//  Event Handlers
// ──────────────────────────────────────────

func (h *Hub) handleDisconnect(c *Client) {
	uid, realmID, ok := h.manager.LogOutBySocketID(c.SocketID)
	if !ok {
		return
	}

	sess := h.manager.GetSession(realmID)
	if sess == nil {
		return
	}

	// Notify everyone still in the same room
	room, _ := sess.GetPlayerRoom(uid)
	h.broadcastToRoom(sess, room, mustBuildMessage(EventPlayerLeftRoom, uid), c.SocketID)
	log.Info().Str("uid", uid).Str("realm", realmID).Msg("Player disconnected")
}

func (h *Hub) handleJoinRealm(c *Client, raw json.RawMessage) {
	var payload JoinRealmPayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		c.SendEvent(EventFailedToJoinRoom, map[string]string{"reason": "Invalid request"})
		return
	}

	// Prevent concurrent double-join
	h.joiningMu.Lock()
	if _, inProgress := h.joining[c.UID]; inProgress {
		h.joiningMu.Unlock()
		c.SendEvent(EventFailedToJoinRoom, map[string]string{"reason": "Already joining a realm"})
		return
	}
	h.joining[c.UID] = struct{}{}
	h.joiningMu.Unlock()

	defer func() {
		h.joiningMu.Lock()
		delete(h.joining, c.UID)
		h.joiningMu.Unlock()
	}()

	// Delegate heavy DB work to the realm service
	RealmServiceInstance.JoinRealm(h, c, payload)
}

func (h *Hub) handleMovePlayer(c *Client, raw json.RawMessage) {
	var p MovePlayerPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	sess := h.manager.GetPlayerSession(c.UID)
	if sess == nil {
		return
	}

	// Move in session (also recalculates proximity)
	changedUIDs := sess.MovePlayer(c.UID, p.X, p.Y)

	// Notify players in the same room about the move
	player, ok := sess.GetPlayer(c.UID)
	if !ok {
		return
	}

	h.broadcastToRoom(sess, player.Room,
		mustBuildMessage(EventPlayerMoved, map[string]any{
			"uid": c.UID, "x": p.X, "y": p.Y,
		}),
		c.SocketID, // exclude the sender
	)

	// Notify players whose proximity group changed
	h.sendProximityUpdates(sess, changedUIDs)
}

func (h *Hub) handleTeleport(c *Client, raw json.RawMessage) {
	var p TeleportPayload
	if err := json.Unmarshal(raw, &p); err != nil {
		return
	}

	sess := h.manager.GetPlayerSession(c.UID)
	if sess == nil {
		return
	}

	player, ok := sess.GetPlayer(c.UID)
	if !ok {
		return
	}

	if player.Room != p.RoomIndex {
		// Leaving the room — notify old room
		h.broadcastToRoom(sess, player.Room,
			mustBuildMessage(EventPlayerLeftRoom, c.UID),
			c.SocketID,
		)

		// Change room in session
		changedUIDs := sess.ChangeRoom(c.UID, p.RoomIndex, p.X, p.Y)

		// Notify new room
		newPlayer, _ := sess.GetPlayer(c.UID)
		h.broadcastToRoom(sess, p.RoomIndex,
			mustBuildMessage(EventPlayerJoinedRoom, newPlayer),
			c.SocketID,
		)

		h.sendProximityUpdates(sess, changedUIDs)
	} else {
		// Same room teleport
		changedUIDs := sess.MovePlayer(c.UID, p.X, p.Y)
		h.broadcastToRoom(sess, player.Room,
			mustBuildMessage(EventPlayerTeleported, map[string]any{
				"uid": c.UID, "x": p.X, "y": p.Y,
			}),
			c.SocketID,
		)
		h.sendProximityUpdates(sess, changedUIDs)
	}
}

func (h *Hub) handleChangedSkin(c *Client, raw json.RawMessage) {
	var skin string
	if err := json.Unmarshal(raw, &skin); err != nil {
		return
	}

	sess := h.manager.GetPlayerSession(c.UID)
	if sess == nil {
		return
	}

	player, ok := sess.GetPlayer(c.UID)
	if !ok {
		return
	}

	sess.SetSkin(c.UID, skin)

	h.broadcastToRoom(sess, player.Room,
		mustBuildMessage(EventPlayerChangedSkin, map[string]string{
			"uid": c.UID, "skin": skin,
		}),
		c.SocketID,
	)
}

func (h *Hub) handleSendMessage(c *Client, raw json.RawMessage) {
	var payload SendMessagePayload
	if err := json.Unmarshal(raw, &payload); err != nil {
		return
	}

	// Validate: non-empty, max 300 chars
	message := utils.RemoveExtraSpaces(payload.Message)
	if message == "" || len(message) > 300 {
		return
	}

	sess := h.manager.GetPlayerSession(c.UID)
	if sess == nil {
		return
	}

	player, ok := sess.GetPlayer(c.UID)
	if !ok {
		return
	}

	outMsg := mustBuildMessage(EventReceiveMessage, map[string]string{
		"uid":     c.UID,
		"type":    payload.Type,
		"message": message,
	})

	switch payload.Type {
	case "global":
		socketIDs := sess.GetAllSocketIDs()
		h.clientsMu.RLock()
		defer h.clientsMu.RUnlock()
		for _, sid := range socketIDs {
			if sid == c.SocketID {
				continue
			}
			if client, ok := h.clients[sid]; ok {
				client.Send(outMsg)
			}
		}

	case "local", "proximity":
		if player.ProximityID != nil {
			socketIDs := sess.GetSocketIDsInProximity(*player.ProximityID)
			h.clientsMu.RLock()
			defer h.clientsMu.RUnlock()
			for _, sid := range socketIDs {
				if sid == c.SocketID {
					continue
				}
				if client, ok := h.clients[sid]; ok {
					client.Send(outMsg)
				}
			}
		}

	case "room":
		h.broadcastToRoom(sess, player.Room, outMsg, c.SocketID)

	case "direct":
		targetPlayer, ok := sess.GetPlayer(payload.TargetID)
		if ok {
			h.SendToSocket(targetPlayer.SocketID, outMsg)
		}
	}
}

func (h *Hub) handleBoardUpdate(c *Client, raw json.RawMessage) {
	sess := h.manager.GetPlayerSession(c.UID)
	if sess == nil {
		return
	}

	// Broadcast board update to everyone in the realm except the sender
	outMsg := mustBuildMessage(EventBoardUpdate, raw)
	socketIDs := sess.GetAllSocketIDs()

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()
	for _, sid := range socketIDs {
		if sid == c.SocketID {
			continue
		}
		if client, ok := h.clients[sid]; ok {
			client.Send(outMsg)
		}
	}
}

// ──────────────────────────────────────────
//  Broadcast helpers
// ──────────────────────────────────────────

// broadcastToRoom sends msg to all clients in the room, excluding excludeSocketID.
func (h *Hub) broadcastToRoom(sess *session.Session, roomIndex int, msg []byte, excludeSocketID string) {
	socketIDs := sess.GetSocketIDsInRoom(roomIndex)

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for _, sid := range socketIDs {
		if sid == excludeSocketID {
			continue
		}
		if client, ok := h.clients[sid]; ok {
			client.Send(msg)
		}
	}
}

// SendToSocket sends a message directly to one specific socket.
func (h *Hub) SendToSocket(socketID string, msg []byte) {
	h.clientsMu.RLock()
	client, ok := h.clients[socketID]
	h.clientsMu.RUnlock()

	if ok {
		client.Send(msg)
	}
}

// KickPlayer sends a "kicked" event to one player and closes their connection.
func (h *Hub) KickPlayer(uid, reason string) {
	sess := h.manager.GetPlayerSession(uid)
	if sess == nil {
		return
	}

	player, ok := sess.GetPlayer(uid)
	if !ok {
		return
	}

	// Notify the kicked player
	h.SendToSocket(player.SocketID, mustBuildMessage(EventKicked, map[string]string{"reason": reason}))

	// Notify their room that they left
	h.broadcastToRoom(sess, player.Room,
		mustBuildMessage(EventPlayerLeftRoom, uid),
		player.SocketID,
	)

	h.manager.LogOutPlayer(uid)
}

// TerminateSession kicks everyone in a realm (used when realm is deleted or changed).
func (h *Hub) TerminateSession(realmID, reason string) {
	socketIDs := h.manager.TerminateSession(realmID, reason)
	msg := mustBuildMessage(EventKicked, map[string]string{"reason": reason})

	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for _, sid := range socketIDs {
		if client, ok := h.clients[sid]; ok {
			client.Send(msg)
		}
	}
}

// sendProximityUpdates notifies each player in changedUIDs of their new proximityId.
func (h *Hub) sendProximityUpdates(sess *session.Session, changedUIDs []string) {
	h.clientsMu.RLock()
	defer h.clientsMu.RUnlock()

	for _, uid := range changedUIDs {
		player, ok := sess.GetPlayer(uid)
		if !ok {
			continue
		}
		client, cOk := h.clients[player.SocketID]
		if !cOk {
			continue
		}
		client.Send(mustBuildMessage(EventProximityUpdate, map[string]any{
			"proximityId": player.ProximityID,
		}))
	}
}

// BroadcastPlayerJoined sends the new player's data to everyone already in the room,
// and sends all existing room players' data back to the new player.
// Called by the realm service after a successful join.
func (h *Hub) BroadcastPlayerJoined(sess *session.Session, player *session.Player, newSocketID string) {
	// Tell new player about everyone already in the room (excluding themselves)
	existingPlayers := sess.GetPlayersInRoom(player.Room)
	h.clientsMu.RLock()
	newClient, ok := h.clients[newSocketID]
	h.clientsMu.RUnlock()

	if ok {
		for _, p := range existingPlayers {
			if p.UID == player.UID {
				continue
			}
			newClient.Send(mustBuildMessage(EventPlayerJoinedRoom, p))
		}
	}

	// Tell everyone already in the room about the new player
	h.broadcastToRoom(sess, player.Room,
		mustBuildMessage(EventPlayerJoinedRoom, player),
		newSocketID,
	)
}

// RealmServiceInstance is injected at startup (set in main.go).
// This breaks the import cycle: ws → realm service, not realm service → ws.
var RealmServiceInstance RealmJoiner

// RealmJoiner is an interface so realm service can be injected without a circular import.
type RealmJoiner interface {
	JoinRealm(hub *Hub, client *Client, payload JoinRealmPayload)
}
