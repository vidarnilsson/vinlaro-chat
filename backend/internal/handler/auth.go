package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/auth"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
	"github.com/vidarnilsson/vinlaro-chat/internal/session"
)

type AuthHandler struct {
	queries *db.Queries
}

func NewAuthHandler(queries *db.Queries) *AuthHandler {
	return &AuthHandler{queries: queries}
}

type registerRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type userResponse struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// Register creates a new user account and starts a session.
// POST /api/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user, err := h.queries.CreateUser(c.Request.Context(), db.CreateUserParams{
		Username: req.Username,
		Email:    req.Email,
		Password: hashed,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "username or email already taken"})
		return
	}

	if err := h.startSession(c, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusCreated, userResponse{UserID: user.ID.String(), Username: user.Username})
}

// Login authenticates a user and starts a session.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil || !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if err := h.startSession(c, user.ID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create session"})
		return
	}

	c.JSON(http.StatusOK, userResponse{UserID: user.ID.String(), Username: user.Username})
}

// Logout deletes the session and clears the cookie.
// POST /api/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	sessionID := session.GetFromCookie(c)
	if sessionID != "" {
		_ = h.queries.DeleteSession(c.Request.Context(), sessionID)
	}
	session.ClearCookie(c)
	c.Status(http.StatusOK)
}

// Me returns the authenticated user's identity.
// GET /api/auth/me
func (h *AuthHandler) Me(c *gin.Context) {
	c.JSON(http.StatusOK, userResponse{
		UserID:   c.GetString(middleware.UserIDKey),
		Username: c.GetString(middleware.UsernameKey),
	})
}

func (h *AuthHandler) startSession(c *gin.Context, userID uuid.UUID) error {
	sessionID, err := session.GenerateID()
	if err != nil {
		return err
	}

	_, err = h.queries.CreateSession(c.Request.Context(), db.CreateSessionParams{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return err
	}

	session.SetCookie(c, sessionID)
	return nil
}
