package models



// Realm maps directly to the "realms" table in Supabase PostgreSQL.
// GORM uses struct tags (e.g. `gorm:"column:owner_id"`) to map
// Go field names to actual database column names.
//
// We use `datatypes.JSON` for map_data because it's stored as a
// JSONB blob in Postgres — GORM can scan it in and out automatically.
type Realm struct {
	ID        string    `json:"$id"`
	OwnerID   string    `json:"owner_id"`
	Name      string    `json:"name"`
	ShareID   string    `json:"share_id"`
	MapData   string    `json:"map_data"`
	OnlyOwner bool      `json:"only_owner"`
	CreatedAt string    `json:"$createdAt"`
	UpdatedAt string    `json:"$updatedAt"`
}



// ──────────────────────────────────────────────
//  Nested types that live inside MapData (JSONB)
// ──────────────────────────────────────────────

// MapData is the full realm map — serialised as JSONB in Postgres.
type MapData struct {
	Spawnpoint Spawnpoint `json:"spawnpoint"`
	Rooms      []Room     `json:"rooms"`
}

type Spawnpoint struct {
	RoomIndex int `json:"roomIndex"`
	X         int `json:"x"`
	Y         int `json:"y"`
}

type Room struct {
	Name      string              `json:"name"`
	ChannelID string              `json:"channelId,omitempty"`
	Tilemap   map[string]*Tile    `json:"tilemap"` // key = "x, y"
}

type Tile struct {
	Floor         string      `json:"floor,omitempty"`
	AboveFloor    string      `json:"above_floor,omitempty"`
	Object        string      `json:"object,omitempty"`
	Impassable    bool        `json:"impassable,omitempty"`
	PrivateAreaID string      `json:"privateAreaId,omitempty"`
	Teleporter    *Teleporter `json:"teleporter,omitempty"`
}

type Teleporter struct {
	RoomIndex int `json:"roomIndex"`
	X         int `json:"x"`
	Y         int `json:"y"`
}
