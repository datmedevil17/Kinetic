package models

import "github.com/google/uuid"

// Profile maps to the "profiles" table in Supabase.
// This stores the user's chosen avatar skin.
// The ID is the same UUID as auth.users.id — a 1:1 relationship.
type Profile struct {
	ID   uuid.UUID `gorm:"column:id;primaryKey;type:uuid" json:"id"`
	Skin string    `gorm:"column:skin;default:'009'" json:"skin"`
}

func (Profile) TableName() string { return "profiles" }
