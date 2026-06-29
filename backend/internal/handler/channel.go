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
	Kind        string `json:"kind"`
}

// ListChannels returns all channels visible to the current user.
// Public channels + private channels the user is a member of.
// GET /api/channels
func (h *ChannelHandler) ListChannels(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	channels, err := h.queries.GetUserChannels(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch channels"})
		return
	}
	if channels == nil {
		channels = []db.Channel{}
	}
	c.JSON(http.StatusOK, channels)
}

// CreateChannel creates a new public or private channel.
// POST /api/channels
func (h *ChannelHandler) CreateChannel(c *gin.Context) {
	var req createChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	desc := sql.NullString{String: req.Description, Valid: req.Description != ""}

	if req.Kind == "private" {
		channel, err := h.queries.CreatePrivateChannel(c.Request.Context(), db.CreatePrivateChannelParams{
			Name:        req.Name,
			Description: desc,
			CreatedBy:   userID,
		})
		if err != nil {
			c.JSON(http.StatusConflict, gin.H{"error": "channel name already taken"})
			return
		}

		// Creator automatically becomes owner.
		_ = h.queries.AddChannelMemberWithRole(c.Request.Context(), db.AddChannelMemberWithRoleParams{
			ChannelID: channel.ID,
			UserID:    userID,
			Role:      "owner",
		})

		c.JSON(http.StatusCreated, channel)
		return
	}

	channel, err := h.queries.CreateChannel(c.Request.Context(), db.CreateChannelParams{
		Name:        req.Name,
		Description: desc,
		CreatedBy:   userID,
	})
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "channel name already taken"})
		return
	}

	c.JSON(http.StatusCreated, channel)
}
