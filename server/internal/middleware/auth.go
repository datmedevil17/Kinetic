package middleware

import (
	"strings"

	"github.com/datmedevil/kinetic/server/internal/utils"
	"github.com/gin-gonic/gin"
)

// Context keys — typed constants prevent string key collisions.
const (
	CtxUID   = "uid"
	CtxEmail = "email"
	CtxToken = "access_token"
)

// Auth is a Gin middleware that:
//  1. Reads the "Authorization: Bearer <token>" header
//  2. Calls Supabase to validate the token
//  3. Injects the user's UID and email into the Gin context
//  4. Calls c.Abort() if auth fails — stops the request dead
//
// How Gin middleware works:
//   Gin middleware is just a HandlerFunc that calls c.Next() to pass
//   control to the next handler, or c.Abort() to stop the chain.
//
//   router.GET("/protected", Auth(), myHandler)
//   → Auth() runs first. If it calls c.Next(), myHandler runs.
//   → If it calls c.Abort(), myHandler never runs.
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract the Bearer token from the Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Unauthorized(c, "Authorization header is required")
			c.Abort()
			return
		}

		// Header format: "Bearer eyJ..."
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			utils.Unauthorized(c, "Authorization header must be: Bearer <token>")
			c.Abort()
			return
		}

		token := parts[1]

		// Validate with Supabase
		user, err := utils.ValidateToken(token)
		if err != nil {
			utils.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}

		// Inject user info into Gin context for downstream handlers
		c.Set(CtxUID, user.ID)
		c.Set(CtxEmail, user.Email)
		c.Set(CtxToken, token)

		// Pass control to the next handler/middleware in the chain
		c.Next()
	}
}

// GetUID is a helper to safely extract the UID from the Gin context.
// Use this in handlers instead of c.GetString(CtxUID) directly.
func GetUID(c *gin.Context) string {
	return c.GetString(CtxUID)
}

// GetToken extracts the access token from the Gin context.
func GetToken(c *gin.Context) string {
	return c.GetString(CtxToken)
}
