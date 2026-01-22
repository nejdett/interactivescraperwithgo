package handlers

import (
	"net/http"

	"github.com/cti-dashboard/dashboard/internal/auth"
	"github.com/cti-dashboard/dashboard/internal/repository"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userRepo *repository.UserRepository
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(userRepo *repository.UserRepository) *AuthHandler {
	return &AuthHandler{
		userRepo: userRepo,
	}
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code":    "INVALID_INPUT",
			"message": "Invalid request body",
			"details": err.Error(),
		})
		return
	}

	// Get user from database
	user, err := h.userRepo.GetByUsername(req.Username)
	if err != nil {
		log.WithFields(log.Fields{
			"username": req.Username,
			"error":    err.Error(),
		}).Warn("Login attempt with invalid username")

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid username or password",
		})
		return
	}

	log.WithFields(log.Fields{
		"username":     req.Username,
		"password_len": len(req.Password),
		"hash_len":     len(user.PasswordHash),
	}).Info("Password check attempt")

	// Check password
	if !auth.CheckPassword(req.Password, user.PasswordHash) {
		log.WithFields(log.Fields{
			"username": req.Username,
		}).Warn("Login attempt with invalid password")

		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "INVALID_CREDENTIALS",
			"message": "Invalid username or password",
		})
		return
	}

	// Set session
	if err := auth.SetSession(c, user.ID, user.Username, user.Role); err != nil {
		log.WithError(err).Error("Failed to set session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to create session",
		})
		return
	}

	log.WithFields(log.Fields{
		"username": user.Username,
		"role":     user.Role,
	}).Info("User logged in successfully")

	c.JSON(http.StatusOK, gin.H{
		"message": "Login successful",
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"role":     user.Role,
		},
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	username := auth.GetUsername(c)

	if err := auth.ClearSession(c); err != nil {
		log.WithError(err).Error("Failed to clear session")
		c.JSON(http.StatusInternalServerError, gin.H{
			"code":    "INTERNAL_ERROR",
			"message": "Failed to logout",
		})
		return
	}

	log.WithField("username", username).Info("User logged out")

	c.JSON(http.StatusOK, gin.H{
		"message": "Logout successful",
	})
}

// GetSession returns the current session information
func (h *AuthHandler) GetSession(c *gin.Context) {
	if !auth.IsAuthenticated(c) {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "UNAUTHORIZED",
			"message": "Not authenticated",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"id":       auth.GetUserID(c),
			"username": auth.GetUsername(c),
			"role":     auth.GetRole(c),
		},
	})
}
