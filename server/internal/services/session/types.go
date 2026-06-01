package session

import "github.com/datmedevil/kinetic/server/internal/models"

// TilePoint is a string key in the format "x, y" (e.g. "3, 5").
// Using a named type instead of plain string makes function signatures clearer
// and prevents accidentally passing the wrong string.
type TilePoint = string

// Player represents a live connected player inside a realm session.
// This is NOT the database model — it's runtime state kept in memory.
type Player struct {
	UID         string  `json:"uid"`
	Username    string  `json:"username"`
	X           int     `json:"x"`
	Y           int     `json:"y"`
	Room        int     `json:"room"`
	SocketID    string  `json:"socketId"` // identifies which WS client this is
	Skin        string  `json:"skin"`
	ProximityID *string `json:"proximityId"` // nil = not in any proximity group
}

// RealmData is an alias for the models type — brought into this package
// so session logic doesn't import models repeatedly.
type RealmData = models.MapData
