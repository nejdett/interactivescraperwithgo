package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

const (
	sessionName = "cti_session"
	userIDKey   = "user_id"
	usernameKey = "username"
	roleKey     = "role"
)

// InitSessionStore initializes the session store with the given secret
func InitSessionStore(secret string) gin.HandlerFunc {
	store := cookie.NewStore([]byte(secret))
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400, // 24 hours
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: 3,     // Strict
	})
	return sessions.Sessions(sessionName, store)
}

// SetSession stores user information in the session
func SetSession(c *gin.Context, userID, username, role string) error {
	session := sessions.Default(c)
	session.Set(userIDKey, userID)
	session.Set(usernameKey, username)
	session.Set(roleKey, role)
	return session.Save()
}

// GetUserID retrieves the user ID from the session
func GetUserID(c *gin.Context) string {
	session := sessions.Default(c)
	userID := session.Get(userIDKey)
	if userID == nil {
		return ""
	}
	return userID.(string)
}

// GetUsername retrieves the username from the session
func GetUsername(c *gin.Context) string {
	session := sessions.Default(c)
	username := session.Get(usernameKey)
	if username == nil {
		return ""
	}
	return username.(string)
}

// GetRole retrieves the user role from the session
func GetRole(c *gin.Context) string {
	session := sessions.Default(c)
	role := session.Get(roleKey)
	if role == nil {
		return ""
	}
	return role.(string)
}

// ClearSession removes all session data
func ClearSession(c *gin.Context) error {
	session := sessions.Default(c)
	session.Clear()
	return session.Save()
}

// IsAuthenticated checks if the user is authenticated
func IsAuthenticated(c *gin.Context) bool {
	return GetUserID(c) != ""
}
