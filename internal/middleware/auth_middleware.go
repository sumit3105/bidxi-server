package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"bidxi/internal/config"
	"bidxi/internal/utils"
)

// AuthMiddleware validates the JWT token in the Authorization header
// and injects user claims into the Gin context for downstream handlers.
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := ""
		
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" {
				tokenString = parts[1]
			}
		}

		// Fallback for WebSockets which can't send headers
		if tokenString == "" {
			tokenString = c.Query("token")
		}

		if tokenString == "" {
			utils.RespondError(c, http.StatusUnauthorized, "Authorization token is required")
			c.Abort()
			return
		}

		claims, err := utils.ValidateToken(tokenString, config.AppConfig.JWTSecret)
		if err != nil {
			utils.RespondError(c, http.StatusUnauthorized, "Invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context for handlers
		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)
		c.Set("username", claims.Username)

		c.Next()
	}
}
