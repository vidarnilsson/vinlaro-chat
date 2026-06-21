package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vidarnilsson/vinlaro-chat/config"
	"github.com/vidarnilsson/vinlaro-chat/internal/auth"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
)

type AuthHandler struct {
	queries *db.Queries
	cfg     *config.Config
}

func NewAuthHandler(queries *db.Queries, cfg *config.Config) *AuthHandler {
	return &AuthHandler{queries: queries, cfg: cfg}
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

type authResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	Username string `json:"username"`
}

// Register creates a new user account.
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
		// A real app would check for unique constraint violations here
		c.JSON(http.StatusConflict, gin.H{"error": "username or email already taken"})
		return
	}

	token, err := auth.GenerateToken(user.ID.String(), user.Username, h.cfg.JWTSecret, h.cfg.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, authResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
	})
}

// Login authenticates a user and returns a JWT.
// POST /api/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.queries.GetUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		// Don't reveal whether the email exists
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	if !auth.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	token, err := auth.GenerateToken(user.ID.String(), user.Username, h.cfg.JWTSecret, h.cfg.JWTExpiryHours)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, authResponse{
		Token:    token,
		UserID:   user.ID.String(),
		Username: user.Username,
	})
}
