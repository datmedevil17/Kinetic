package main

import (
	"os"

	"github.com/datmedevil/kinetic/server/internal/config"
	"github.com/datmedevil/kinetic/server/internal/database"
	realmhandler "github.com/datmedevil/kinetic/server/internal/handlers/realm"
	playerhandler "github.com/datmedevil/kinetic/server/internal/handlers/player"
	wshandler "github.com/datmedevil/kinetic/server/internal/handlers/ws"
	"github.com/datmedevil/kinetic/server/internal/middleware"
	"github.com/datmedevil/kinetic/server/internal/services/realm"
	"github.com/datmedevil/kinetic/server/internal/services/session"
	"github.com/datmedevil/kinetic/server/internal/services/ws"
	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// ── 1. Structured logging ─────────────────────────────────────────────
	// zerolog.ConsoleWriter gives human-readable output in dev.
	// In production, switch to json: log.Logger = zerolog.New(os.Stdout)
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout}).
		With().
		Timestamp().
		Caller().
		Logger()

	// ── 2. Load config ────────────────────────────────────────────────────
	// Reads .env file + environment variables. Panics if required vars missing.
	cfg := config.Load()

	// ── 3. Gin mode ───────────────────────────────────────────────────────
	// "release" suppresses debug output, "debug" shows route table on startup.
	gin.SetMode(cfg.GinMode)

	// ── 4. Supabase client ────────────────────────────────────────────────
	// Used by JWT validation and (optionally) PostgREST queries.
	utils.InitSupabase(cfg.SupabaseURL, cfg.SupabaseServiceRole)
	log.Info().Msg("✅ Supabase client initialised")

	// ── 5. Database (GORM → Supabase PostgreSQL) ──────────────────────────
	database.Connect(cfg.DatabaseURL)
	database.Migrate()

	// ── 6. Session manager ────────────────────────────────────────────────
	// In-memory state for all live realms. Thread-safe with RWMutex.
	manager := session.NewManager()

	// ── 7. WebSocket Hub ──────────────────────────────────────────────────
	// The hub runs a single goroutine that serialises all WS event processing.
	hub := ws.NewHub(manager, cfg.MaxPlayersPerRealm)
	go hub.Run() // start the event loop

	// ── 8. Realm service ──────────────────────────────────────────────────
	// Handles join logic: DB queries, access checks, session creation.
	// Injected into the hub via the RealmJoiner interface to break the import cycle:
	//   hub (ws package) does not import realm package directly.
	realmSvc := realm.NewService(manager, cfg.MaxPlayersPerRealm)
	ws.RealmServiceInstance = realmSvc

	// ── 9. HTTP handlers ──────────────────────────────────────────────────
	wsHandler     := wshandler.NewHandler(hub)
	realmHandler  := realmhandler.NewHandler(manager)
	playerHandler := playerhandler.NewHandler(hub)

	// ── 10. Gin router ────────────────────────────────────────────────────
	r := gin.New()

	// Global middleware:
	//   - Recovery: catches panics and returns 500 instead of crashing
	//   - Logger: prints request method, path, status, latency
	//   - CORS: allow all origins (AllowAllOrigins = true)
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS(cfg.FrontendURL))

	// ── Routes ────────────────────────────────────────────────────────────

	// Public routes (no auth)
	r.GET("/health", playerHandler.Health)

	// WebSocket endpoint — auth is validated inside the handler via query param
	// because the browser WebSocket API cannot set custom headers.
	// URL: ws://host/ws?token=<supabase-access-token>
	r.GET("/ws", wsHandler.ServeWS)

	// Protected REST routes (require Authorization: Bearer <token>)
	api := r.Group("/api/v1", middleware.Auth())
	{
		// GET /api/v1/players-in-room?roomIndex=N
		api.GET("/players-in-room", realmHandler.GetPlayersInRoom)

		// GET /api/v1/player-counts?realmIds=uuid1,uuid2
		// (Auth required so we can't be easily spammed)
		api.GET("/player-counts", realmHandler.GetPlayerCounts)

		// POST /api/v1/admin/kick   body: { uid, reason }
		api.POST("/admin/kick", playerHandler.KickPlayer)
	}

	// ── 11. Start server ──────────────────────────────────────────────────
	addr := ":" + cfg.Port
	log.Info().Str("addr", addr).Msg("🚀 Kinetic server starting")

	if err := r.Run(addr); err != nil {
		log.Fatal().Err(err).Msg("Server failed to start")
	}
}
