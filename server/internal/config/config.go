package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port                string
	GinMode             string
	FrontendURL         string
	AppwriteEndpoint    string
	AppwriteProjectID   string
	AppwriteAPIKey      string
	AppwriteDatabaseID  string
	AppwriteRealmsID    string
	AppwriteProfilesID  string
	MaxPlayersPerRealm  int
}

func Load() *Config {
	_ = godotenv.Load()

	cfg := &Config{
		Port:                getEnv("PORT", "8080"),
		GinMode:             getEnv("GIN_MODE", "debug"),
		FrontendURL:         getEnv("FRONTEND_URL", "http://localhost:3000"),
		AppwriteEndpoint:    mustGetEnv("APPWRITE_ENDPOINT"),
		AppwriteProjectID:   mustGetEnv("APPWRITE_PROJECT_ID"),
		AppwriteAPIKey:      mustGetEnv("APPWRITE_API_KEY"),
		AppwriteDatabaseID:  mustGetEnv("APPWRITE_DATABASE_ID"),
		AppwriteRealmsID:    mustGetEnv("APPWRITE_REALMS_COLLECTION_ID"),
		AppwriteProfilesID:  mustGetEnv("APPWRITE_PROFILES_COLLECTION_ID"),
		MaxPlayersPerRealm:  getEnvInt("MAX_PLAYERS_PER_REALM", 30),
	}

	return cfg
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func mustGetEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("FATAL: required environment variable %q is not set", key)
	}
	return val
}

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
