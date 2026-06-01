package session

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Session represents one live realm in memory.
// It tracks every connected player, their positions, and which room they're in.
//
// Thread safety:
//   All public methods acquire s.mu before touching shared maps.
//   Go's sync.RWMutex allows multiple concurrent Reads (RLock)
//   but only one Write at a time (Lock). This is perfect because
//   reads (getPlayersInRoom) are far more frequent than writes (movePlayer).
type Session struct {
	ID       string
	MapData  RealmData

	mu          sync.RWMutex
	players     map[string]*Player              // uid → Player
	playerRooms map[int]map[string]struct{}     // roomIndex → Set<uid>
	// positions[roomIndex]["x, y"] = Set<uid> — for fast proximity lookup
	positions   map[int]map[TilePoint]map[string]struct{}
}

// NewSession creates an empty Session for a realm.
func NewSession(id string, mapData RealmData) *Session {
	s := &Session{
		ID:          id,
		MapData:     mapData,
		players:     make(map[string]*Player),
		playerRooms: make(map[int]map[string]struct{}),
		positions:   make(map[int]map[TilePoint]map[string]struct{}),
	}

	// Pre-allocate room buckets so we never nil-check them later
	for i := range mapData.Rooms {
		s.playerRooms[i] = make(map[string]struct{})
		s.positions[i] = make(map[TilePoint]map[string]struct{})
	}

	return s
}

// AddPlayer spawns a player at the realm's spawnpoint.
func (s *Session) AddPlayer(socketID, uid, username, skin string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// If the player is already here (e.g. double-join), remove first
	s.removePlayerLocked(uid)

	spawn := s.MapData.Spawnpoint
	p := &Player{
		UID:      uid,
		Username: username,
		X:        spawn.X,
		Y:        spawn.Y,
		Room:     spawn.RoomIndex,
		SocketID: socketID,
		Skin:     skin,
	}

	s.players[uid] = p
	s.playerRooms[spawn.RoomIndex][uid] = struct{}{}
	s.addPosition(spawn.RoomIndex, spawn.X, spawn.Y, uid)
}

// RemovePlayer removes a player from the session entirely.
func (s *Session) RemovePlayer(uid string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.removePlayerLocked(uid)
}

// removePlayerLocked is the internal remove — assumes mu is already held.
func (s *Session) removePlayerLocked(uid string) {
	p, ok := s.players[uid]
	if !ok {
		return
	}

	delete(s.playerRooms[p.Room], uid)
	s.removePosition(p.Room, p.X, p.Y, uid)
	delete(s.players, uid)
}

// MovePlayer moves a player to (x, y) and returns UIDs whose ProximityID changed.
func (s *Session) MovePlayer(uid string, x, y int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.players[uid]
	if !ok {
		return nil
	}

	// Update position index
	s.removePosition(p.Room, p.X, p.Y, uid)
	p.X = x
	p.Y = y
	s.addPosition(p.Room, x, y, uid)

	return s.setProximityLocked(uid)
}

// ChangeRoom moves a player to a different room and returns changed proximity UIDs.
func (s *Session) ChangeRoom(uid string, roomIndex, x, y int) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, ok := s.players[uid]
	if !ok {
		return nil
	}

	// Leave old room
	delete(s.playerRooms[p.Room], uid)
	s.removePosition(p.Room, p.X, p.Y, uid)

	// Enter new room
	p.Room = roomIndex
	p.X = x
	p.Y = y
	s.playerRooms[roomIndex][uid] = struct{}{}
	s.addPosition(roomIndex, x, y, uid)

	return s.setProximityLocked(uid)
}

// GetPlayer returns a copy of the player struct.
func (s *Session) GetPlayer(uid string) (*Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.players[uid]
	if !ok {
		return nil, false
	}
	// Return a copy so callers can't accidentally mutate shared state
	cp := *p
	return &cp, true
}

// GetPlayersInRoom returns a snapshot of all players in a room.
func (s *Session) GetPlayersInRoom(roomIndex int) []*Player {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uids := s.playerRooms[roomIndex]
	result := make([]*Player, 0, len(uids))
	for uid := range uids {
		if p, ok := s.players[uid]; ok {
			cp := *p
			result = append(result, &cp)
		}
	}
	return result
}

// GetSocketIDsInRoom returns all socketIDs of players in a room.
func (s *Session) GetSocketIDsInRoom(roomIndex int) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	uids := s.playerRooms[roomIndex]
	result := make([]string, 0, len(uids))
	for uid := range uids {
		if p, ok := s.players[uid]; ok {
			result = append(result, p.SocketID)
		}
	}
	return result
}

// GetPlayerRoom returns the room index of a player.
func (s *Session) GetPlayerRoom(uid string) (int, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	p, ok := s.players[uid]
	if !ok {
		return 0, false
	}
	return p.Room, true
}

// SetSkin updates a player's skin.
func (s *Session) SetSkin(uid, skin string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if p, ok := s.players[uid]; ok {
		p.Skin = skin
	}
}

// GetPlayerCount returns the number of connected players.
func (s *Session) GetPlayerCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.players)
}

// GetAllPlayerIDs returns all connected UIDs.
func (s *Session) GetAllPlayerIDs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.players))
	for uid := range s.players {
		ids = append(ids, uid)
	}
	return ids
}

// ─────────────────────────────────────────
//  Position index helpers (called with mu held)
// ─────────────────────────────────────────

func (s *Session) addPosition(room, x, y int, uid string) {
	key := tileKey(x, y)
	if s.positions[room][key] == nil {
		s.positions[room][key] = make(map[string]struct{})
	}
	s.positions[room][key][uid] = struct{}{}
}

func (s *Session) removePosition(room, x, y int, uid string) {
	key := tileKey(x, y)
	if set, ok := s.positions[room][key]; ok {
		delete(set, uid)
		if len(set) == 0 {
			delete(s.positions[room], key)
		}
	}
}

func tileKey(x, y int) TilePoint {
	return fmt.Sprintf("%d, %d", x, y)
}

// ─────────────────────────────────────────
//  Proximity detection (called with mu held)
// ─────────────────────────────────────────

const proximityRange = 3 // tiles — creates a 7×7 scan grid

// setProximityLocked recalculates proximity groups for uid and its neighbours.
// Returns the UIDs whose ProximityID changed (so the hub can notify them).
// Must be called with s.mu already locked.
func (s *Session) setProximityLocked(uid string) []string {
	p, ok := s.players[uid]
	if !ok {
		return nil
	}

	changed := make(map[string]struct{})
	originalID := p.ProximityID
	hasNeighbors := false

	// Scan the 7×7 tile grid around the player
	for dx := -proximityRange; dx <= proximityRange; dx++ {
		for dy := -proximityRange; dy <= proximityRange; dy++ {
			key := tileKey(p.X+dx, p.Y+dy)
			neighbors, ok := s.positions[p.Room][key]
			if !ok {
				continue
			}

			for otherUID := range neighbors {
				if otherUID == uid {
					continue
				}
				hasNeighbors = true
				other := s.players[otherUID]

				switch {
				case other.ProximityID == nil && p.ProximityID == nil:
					// Both alone — create a new shared group
					id := uuid.New().String()
					p.ProximityID = &id
					other.ProximityID = &id
					changed[uid] = struct{}{}
					changed[otherUID] = struct{}{}

				case other.ProximityID == nil:
					// Player has a group, other doesn't — absorb other
					other.ProximityID = p.ProximityID
					changed[otherUID] = struct{}{}

				case p.ProximityID == nil:
					// Other has a group, player doesn't — join it
					p.ProximityID = other.ProximityID
					changed[uid] = struct{}{}
				}
			}
		}
	}

	// No one nearby — leave the group
	if !hasNeighbors && p.ProximityID != nil {
		p.ProximityID = nil
		if originalID != nil {
			changed[uid] = struct{}{}
		}
	}

	result := make([]string, 0, len(changed))
	for u := range changed {
		result = append(result, u)
	}
	return result
}
