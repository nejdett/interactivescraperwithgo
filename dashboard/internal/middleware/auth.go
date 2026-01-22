package middleware

import (
	"net/http"

	"github.com/cti-dashboard/dashboard/internal/auth"
	"github.com/gin-gonic/gin"
)

// RequireAuth is a middleware that requires authentication
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !auth.IsAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin is a middleware that requires admin role
func RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !auth.IsAuthenticated(c) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		role := auth.GetRole(c)
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{
				"code":    "FORBIDDEN",
				"message": "Admin access required",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
