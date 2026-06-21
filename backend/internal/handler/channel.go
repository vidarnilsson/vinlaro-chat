package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
)

type ChannelHandler struct {
	queries *db.Queries
}

func NewChannelHandler(queries *db.Queries) *ChannelHandler {
	return &ChannelHandler{queries: queries}
}

type createChannelRequest struct {
	Name        string `json:"name"        binding:"required,min=2,max=64"`
	Description string `json:"description"`
}

// ListChannels returns all channels.
// GET /api/channels
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	channels, err := h.queries.ListChannels(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch channels"})
		return
	}
	c.JSON(http.StatusOK, channels)
}

// CreateChannel creates a new channel.
// POST /api/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.GetString(middleware.UserIDKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	channel, err := h.queries.CreateChannel(c.Request.Context(), db.CreateChannelParams{
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		CreatedBy:   userID,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "channel name already taken"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}
