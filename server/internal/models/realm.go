package models

import (
	"time"

	"github.com/google/uuid"
)

// Realm maps directly to the "realms" table in Supabase PostgreSQL.
// GORM uses struct tags (e.g. `gorm:"column:owner_id"`) to map
// Go field names to actual database column names.
//
// We use `datatypes.JSON` for map_data because it's stored as a
// JSONB blob in Postgres — GORM can scan it in and out automatically.
type Realm struct {
	// gorm:"primaryKey" tells GORM this is the PK — no auto-increment, it's a UUID
	ID        uuid.UUID  `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	OwnerID   uuid.UUID  `gorm:"column:owner_id;type:uuid;not null" json:"owner_id"`
	Name      string     `gorm:"column:name;not null" json:"name"`
	ShareID   uuid.UUID  `gorm:"column:share_id;type:uuid;not null" json:"share_id"`
	MapData   MapData    `gorm:"column:map_data;type:jsonb;serializer:json" json:"map_data"`
	OnlyOwner bool       `gorm:"column:only_owner;default:false" json:"only_owner"`
	CreatedAt time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// TableName overrides the default GORM table name (which would be "realms" anyway,
// but being explicit prevents surprises if you rename the struct).
func (Realm) TableName() string { return "realms" }

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
