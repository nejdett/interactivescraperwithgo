package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// ErrorResponse represents a standardized error response
type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

// ErrorHandler is a middleware that handles errors and panics
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic with stack trace
				log.WithFields(log.Fields{
					"error":      err,
					"stack":      string(debug.Stack()),
					"path":       c.Request.URL.Path,
					"method":     c.Request.Method,
					"ip":         c.ClientIP(),
					"user_agent": c.Request.UserAgent(),
				}).Error("Panic recovered")

				// Return 500 error
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "An unexpected error occurred",
				})
				c.Abort()
			}
		}()

		c.Next()

		// Check if there were any errors during request processing
		if len(c.Errors) > 0 {
			err := c.Errors.Last()
			
			// Log the error
			log.WithFields(log.Fields{
				"error":  err.Error(),
				"path":   c.Request.URL.Path,
				"method": c.Request.Method,
			}).Error("Request error")

			// If response hasn't been written yet, send error response
			if !c.Writer.Written() {
				c.JSON(http.StatusInternalServerError, ErrorResponse{
					Code:    "INTERNAL_ERROR",
					Message: "An error occurred while processing the request",
				})
			}
		}
	}
}

// MapErrorToStatus maps common error messages to HTTP status codes
func MapErrorToStatus(errMsg string) int {
	switch errMsg {
	case "user not found", "content not found", "category not found":
		return http.StatusNotFound
	case "invalid credentials":
		return http.StatusUnauthorized
	case "unauthorized":
		return http.StatusUnauthorized
	case "forbidden":
		return http.StatusForbidden
	case "invalid input":
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
