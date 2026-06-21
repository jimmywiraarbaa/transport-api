package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/jwt"
	"github.com/jimmywiraarbaa/transport-api/internal/utils/response"
)

// extractBearer pulls the token from the Authorization header.
func extractBearer(c *gin.Context) string {
	h := c.GetHeader("Authorization")
	if h == "" {
		return ""
	}
	parts := strings.SplitN(h, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

// RequireAuth enforces a valid access token; aborts with 401 otherwise.
func RequireAuth(mgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			response.Unauthorized(c, "Missing bearer token")
			c.Abort()
			return
		}
		claims, err := mgr.Verify(jwt.AccessToken, token)
		if err != nil {
			response.Unauthorized(c, "Invalid or expired token")
			c.Abort()
			return
		}
		SetUser(c, claims.UserID, claims.Email)
		c.Next()
	}
}

// OptionalAuth populates the user context when a valid token is present,
// but never blocks the request.
func OptionalAuth(mgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		token := extractBearer(c)
		if token == "" {
			c.Next()
			return
		}
		if claims, err := mgr.Verify(jwt.AccessToken, token); err == nil {
			SetUser(c, claims.UserID, claims.Email)
		}
		c.Next()
	}
}
