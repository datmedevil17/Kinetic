package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a Gin middleware that allows ALL origins.
// This is intentional for Kinetic — the frontend can be hosted anywhere
// (Vercel, localhost, custom domain) and we don't want to hard-block origins.
//
// Note: AllowAllOrigins cannot be combined with AllowCredentials = true
// (browser security restriction). We handle auth via Bearer tokens in
// the Authorization header, so we don't need cookies — credentials not needed.
func CORS(_ string) gin.HandlerFunc {
	config := cors.Config{
		// Allow every origin — no origin whitelist
		AllowAllOrigins: true,

		// Methods the browser is allowed to call
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// Headers the browser is allowed to send
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Authorization",
			"Accept",
			"X-Requested-With",
		},

		// Allow the frontend to read response headers
		ExposeHeaders: []string{"Content-Length", "Content-Type"},
	}

	return cors.New(config)
}
