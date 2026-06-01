package session

import "sync"

// Manager is the top-level coordinator of all live realm sessions.
// It maintains three lookup maps for O(1) access in every direction.
//
// Think of it as the "lobby server" — it knows where every player is.
type Manager struct {
	mu sync.RWMutex

	// realmID → Session
	sessions map[string]*Session

	// uid → realmID  (which realm is this player in?)
	playerRealm map[string]string

	// socketID → uid  (which player owns this WebSocket connection?)
	socketPlayer map[string]string
}

// Global singleton — created once in main.go and passed around.
var GlobalManager = NewManager()

// NewManager creates an empty Manager.
func NewManager() *Manager {
	return &Manager{
		sessions:     make(map[string]*Session),
		playerRealm:  make(map[string]string),
		socketPlayer: make(map[string]string),
	}
}

// CreateSession starts a new in-memory session for a realm.
func (m *Manager) CreateSession(realmID string, mapData RealmData) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.sessions[realmID]; !exists {
		m.sessions[realmID] = NewSession(realmID, mapData)
	}
}

// GetSession returns the session for a realm (nil if none).
func (m *Manager) GetSession(realmID string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.sessions[realmID]
}

// GetPlayerSession returns the session the given player is currently in.
func (m *Manager) GetPlayerSession(uid string) *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()

	realmID, ok := m.playerRealm[uid]
	if !ok {
		return nil
	}
	return m.sessions[realmID]
}

// AddPlayerToSession adds a player to a realm's session and registers lookups.
func (m *Manager) AddPlayerToSession(socketID, realmID, uid, username, skin string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	session := m.sessions[realmID]
	if session == nil {
		return
	}

	session.AddPlayer(socketID, uid, username, skin)
	m.playerRealm[uid] = realmID
	m.socketPlayer[socketID] = uid
}

// LogOutBySocketID removes the player associated with a socket.
// Returns the uid and realmID so the hub can notify the room.
func (m *Manager) LogOutBySocketID(socketID string) (uid, realmID string, ok bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	uid, ok = m.socketPlayer[socketID]
	if !ok {
		return "", "", false
	}

	realmID = m.playerRealm[uid]

	if s, exists := m.sessions[realmID]; exists {
		s.RemovePlayer(uid)
	}

	delete(m.socketPlayer, socketID)
	delete(m.playerRealm, uid)

	return uid, realmID, true
}

// LogOutPlayer removes a player by UID.
func (m *Manager) LogOutPlayer(uid string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	realmID, ok := m.playerRealm[uid]
	if !ok {
		return
	}

	if s, exists := m.sessions[realmID]; exists {
		p, pOk := s.GetPlayer(uid)
		if pOk {
			delete(m.socketPlayer, p.SocketID)
		}
		s.RemovePlayer(uid)
	}

	delete(m.playerRealm, uid)
}

// TerminateSession kicks all players from a realm and destroys the session.
// Returns socket IDs that need to receive the "kicked" event.
func (m *Manager) TerminateSession(realmID, reason string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	session, ok := m.sessions[realmID]
	if !ok {
		return nil
	}

	// Collect all socket IDs before wiping state
	allUIDs := session.GetAllPlayerIDs()
	socketIDs := make([]string, 0, len(allUIDs))

	for _, uid := range allUIDs {
		if p, pOk := session.GetPlayer(uid); pOk {
			socketIDs = append(socketIDs, p.SocketID)
			delete(m.socketPlayer, p.SocketID)
		}
		delete(m.playerRealm, uid)
	}

	delete(m.sessions, realmID)
	return socketIDs
}

// GetPlayerCounts returns the player count for each of the given realm IDs.
func (m *Manager) GetPlayerCounts(realmIDs []string) []int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	counts := make([]int, len(realmIDs))
	for i, id := range realmIDs {
		if s, ok := m.sessions[id]; ok {
			counts[i] = s.GetPlayerCount()
		}
	}
	return counts
}
