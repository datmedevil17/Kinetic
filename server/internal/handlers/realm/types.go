package realmhandler

// PlayersInRoomResponse is returned by GET /api/v1/players-in-room
type PlayersInRoomResponse struct {
	Players any `json:"players"`
}

// PlayerCountsRequest mirrors the query param shape for GET /api/v1/player-counts
// ?realmIds=uuid1,uuid2,uuid3
type PlayerCountsResponse struct {
	PlayerCounts []int `json:"playerCounts"`
}
