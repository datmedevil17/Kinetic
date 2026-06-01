package database

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the package-level GORM database handle.
// It is safe to use concurrently — GORM manages a connection pool internally.
var DB *gorm.DB

// Connect opens a connection to the PostgreSQL database (Supabase).
// dsn = "postgres://user:password@host:port/dbname"
//
// Why GORM for Supabase?
//   Supabase is just PostgreSQL under the hood.
//   GORM gives us typed queries, auto-migrations, and no raw SQL boilerplate.
func Connect(dsn string) {
	var err error

	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		// In production, set logger.Silent to suppress SQL logs
		Logger: logger.Default.LogMode(logger.Info),

		// DisableForeignKeyConstraintWhenMigrating avoids issues with
		// Supabase RLS policies during AutoMigrate
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		log.Fatalf("FATAL: failed to connect to database: %v", err)
	}

	// Get the underlying *sql.DB to configure the connection pool
	sqlDB, err := DB.DB()
	if err != nil {
		log.Fatalf("FATAL: failed to get sql.DB: %v", err)
	}

	// Connection pool tuning:
	// MaxIdleConns  = connections kept open even when idle (fast reuse)
	// MaxOpenConns  = hard cap on total connections to Postgres
	// ConnMaxLifetime = recycle connections after this duration (prevents stale TCP)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)

	log.Println("✅ Database connected successfully")
}
