package database

import (
	"github.com/appwrite/sdk-for-go/v4/permission"
	"github.com/appwrite/sdk-for-go/v4/role"
	"github.com/rs/zerolog/log"
)

// Migrate ensures the database, realms collection, and profiles collection
// exist with the correct attributes. Safe to call on every startup — it skips
// anything that already exists.
func Migrate(databaseID, realmsID, profilesID string) {
	ensureDatabase(databaseID)
	ensureRealmsCollection(databaseID, realmsID)
	ensureProfilesCollection(databaseID, profilesID)
	log.Info().Msg("✅ Appwrite schema migration complete")
}

func ensureDatabase(dbID string) {
	_, err := Databases.Get(dbID)
	if err == nil {
		return // already exists
	}
	if _, err := Databases.Create(dbID, "Kinetic"); err != nil {
		log.Warn().Err(err).Msg("Could not create database (may already exist)")
	} else {
		log.Info().Str("id", dbID).Msg("Created database")
	}
}

var userPerms = []string{
	permission.Read(role.Users("")),
	permission.Create(role.Users("")),
	permission.Update(role.Users("")),
	permission.Delete(role.Users("")),
}

func ensureRealmsCollection(dbID, collID string) {
	_, err := Databases.GetCollection(dbID, collID)
	if err == nil {
		// Ensure document security is off so collection-level permissions apply
		Databases.UpdateCollection(dbID, collID,
			Databases.WithUpdateCollectionDocumentSecurity(false),
			Databases.WithUpdateCollectionPermissions(userPerms),
		)
		return
	}

	if _, err := Databases.CreateCollection(dbID, collID, "Realms",
		Databases.WithCreateCollectionPermissions(userPerms),
		Databases.WithCreateCollectionDocumentSecurity(false),
	); err != nil {
		log.Error().Err(err).Msg("Failed to create realms collection")
		return
	}
	log.Info().Str("id", collID).Msg("Created realms collection")

	attrs := []func(){
		func() { Databases.CreateStringAttribute(dbID, collID, "owner_id", 36, true) },
		func() { Databases.CreateStringAttribute(dbID, collID, "name", 64, true) },
		func() { Databases.CreateStringAttribute(dbID, collID, "share_id", 36, true) },
		func() { Databases.CreateBooleanAttribute(dbID, collID, "only_owner", true) },
		func() {
			Databases.CreateStringAttribute(dbID, collID, "map_data", 5_000_000, false,
				Databases.WithCreateStringAttributeDefault(""),
			)
		},
	}
	for _, fn := range attrs {
		fn()
	}

	// Indexes required for listDocuments queries
	Databases.CreateIndex(dbID, collID, "owner_id_idx", "key", []string{"owner_id"})
	Databases.CreateIndex(dbID, collID, "share_id_idx", "key", []string{"share_id"})
}

func ensureProfilesCollection(dbID, collID string) {
	_, err := Databases.GetCollection(dbID, collID)
	if err == nil {
		Databases.UpdateCollection(dbID, collID,
			Databases.WithUpdateCollectionDocumentSecurity(false),
			Databases.WithUpdateCollectionPermissions(userPerms),
		)
		return
	}

	if _, err := Databases.CreateCollection(dbID, collID, "Profiles",
		Databases.WithCreateCollectionPermissions(userPerms),
		Databases.WithCreateCollectionDocumentSecurity(false),
	); err != nil {
		log.Error().Err(err).Msg("Failed to create profiles collection")
		return
	}
	log.Info().Str("id", collID).Msg("Created profiles collection")

	Databases.CreateStringAttribute(dbID, collID, "skin", 64, false,
		Databases.WithCreateStringAttributeDefault(""),
	)
	Databases.CreateStringAttribute(dbID, collID, "visited_realms", 36, false,
		Databases.WithCreateStringAttributeArray(true),
	)
}
