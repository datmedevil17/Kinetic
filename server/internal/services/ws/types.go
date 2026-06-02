package ws

import "encoding/json"

// ─────────────────────────────────────────────
//  Incoming event names (client → server)
// ─────────────────────────────────────────────

const (
	EventJoinRealm   = "joinRealm"
	EventMovePlayer  = "movePlayer"
	EventTeleport    = "teleport"
	EventChangedSkin = "changedSkin"
	EventSendMessage = "sendMessage"
	EventBoardUpdate = "boardUpdate"
	EventDisconnect  = "disconnect"
)

// ─────────────────────────────────────────────
//  Outgoing event names (server → client)
// ─────────────────────────────────────────────

const (
	EventJoinedRealm        = "joinedRealm"
	EventFailedToJoinRoom   = "failedToJoinRoom"
	EventPlayerJoinedRoom   = "playerJoinedRoom"
	EventPlayerLeftRoom     = "playerLeftRoom"
	EventPlayerMoved        = "playerMoved"
	EventPlayerTeleported   = "playerTeleported"
	EventPlayerChangedSkin  = "playerChangedSkin"
	EventReceiveMessage     = "receiveMessage"
	EventProximityUpdate    = "proximityUpdate"
	EventKicked             = "kicked"
)

// ─────────────────────────────────────────────
//  Wire format for ALL WebSocket messages
// ─────────────────────────────────────────────

// Message is the envelope for every WebSocket message in both directions.
// Client sends:  { "event": "movePlayer", "payload": { "x": 3, "y": 5 } }
// Server sends:  { "event": "playerMoved", "payload": { "uid": "...", "x": 3, "y": 5 } }
type Message struct {
	Event   string          `json:"event"`
	Payload json.RawMessage `json:"payload"` // raw JSON — decoded per-event
}

// ─────────────────────────────────────────────
//  Incoming payload structs (validated on receipt)
// ─────────────────────────────────────────────

type JoinRealmPayload struct {
	RealmID string `json:"realmId" binding:"required"`
	ShareID string `json:"shareId" binding:"required"`
}

type MovePlayerPayload struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type TeleportPayload struct {
	X         int `json:"x"`
	Y         int `json:"y"`
	RoomIndex int `json:"roomIndex"`
}

type SendMessagePayload struct {
	Type     string `json:"type"`               // "global", "local", "direct"
	TargetID string `json:"targetId,omitempty"` // For direct messages
	Message  string `json:"message"`
}

// ─────────────────────────────────────────────
//  Outgoing payload builder helpers
// ─────────────────────────────────────────────

func buildMessage(event string, payload any) ([]byte, error) {
	p, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.Marshal(Message{Event: event, Payload: p})
}

func mustBuildMessage(event string, payload any) []byte {
	b, _ := buildMessage(event, payload)
	return b
}
