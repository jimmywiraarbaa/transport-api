package middleware

import (
	"github.com/gin-gonic/gin"
)

// contextKey is unexported to avoid key collisions.
type contextKey string

const (
	keyUserID contextKey = "user_id"
	keyEmail  contextKey = "email"
)

// SetUser stores the authenticated user's identity in the request context.
func SetUser(c *gin.Context, userID, email string) {
	c.Set(string(keyUserID), userID)
	c.Set(string(keyEmail), email)
}

// UserID extracts the authenticated user id from the context, empty if absent.
func UserID(c *gin.Context) string {
	v, _ := c.Get(string(keyUserID))
	s, _ := v.(string)
	return s
}

// Email extracts the authenticated user email from the context.
func Email(c *gin.Context) string {
	v, _ := c.Get(string(keyEmail))
	s, _ := v.(string)
	return s
}
