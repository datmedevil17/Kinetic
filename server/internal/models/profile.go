package models


// Profile maps to the "profiles" table.
type Profile struct {
	ID   string `json:"$id"`
	Skin string `json:"skin"`
}
