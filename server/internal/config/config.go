package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds every environment variable the server needs.
// Having one central config struct means you never scatter os.Getenv() calls
// throughout your codebase — everything is read once at startup.
type Config struct {
	Port                string
	GinMode             string
	FrontendURL         string
	SupabaseURL         string
	SupabaseAnonKey     string
	SupabaseServiceRole string
	DatabaseURL         string
	MaxPlayersPerRealm  int
}

// Load reads the .env file (if present) and then reads environment variables.
// Panics if any required variable is missing — fail fast at startup, not mid-request.
func Load() *Config {
	// Load .env file — in production this won't exist (env vars are set by the host)
	// so we ignore the error
	_ = godotenv.Load()

	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		SupabaseURL:         mustGetEnv("SUPABASE_URL"),
		SupabaseAnonKey:     mustGetEnv("SUPABASE_ANON_KEY"),
		SupabaseServiceRole: mustGetEnv("SUPABASE_SERVICE_ROLE_KEY"),
		DatabaseURL:         mustGetEnv("DATABASE_URL"),
		MaxPlayersPerRealm:  getEnvInt("MAX_PLAYERS_PER_REALM", 30),
	}

	return cfg
}

// getEnv returns the value of an env var, or a default if not set.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// mustGetEnv panics if the env var is not set — used for required variables.
func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: required environment variable %q is not set", key)
	}
	return val
}

// getEnvInt reads an env var as an integer with a fallback default.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("WARN: %q is not a valid int, using default %d", key, defaultVal)
		return defaultVal
	}
	return n
}
