package database

import (
	"log"

	"github.com/datmedevil/kinetic/server/internal/models"
)

// Migrate runs GORM's AutoMigrate on all models.
//
// What AutoMigrate does:
//   - Creates tables that don't exist yet
//   - Adds missing columns to existing tables
//   - Does NOT drop columns or change existing column types (safe for production)
//
// We do NOT auto-migrate the 'realms' or 'profiles' tables here because
// they already exist in Supabase with RLS policies.
// AutoMigrate would try to re-create them and fail.
//
// Instead, we only migrate tables that are NEW to the Go backend.
// For now, that's none — we read realms & profiles directly.
// This file is here as a pattern for when you add new Go-owned tables.
func Migrate() {
	err := DB.AutoMigrate(
		// Add new Go-owned models here as you create them
		// e.g. &models.AuditLog{},
		&models.Profile{}, // read-only in practice; GORM just verifies column mapping
	)

	if err != nil {
		log.Fatalf("FATAL: auto migration failed: %v", err)
	}

	log.Println("✅ Database migration complete")
}
