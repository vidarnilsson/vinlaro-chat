package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
)

// checkChannelAccess returns true and sets no response if the user may access
// the channel. For public channels this is always true. For private/dm channels
// the user must be a channel_members row. Returns false and writes 403/404/500
// when access is denied — callers should return immediately in that case.
func checkChannelAccess(c *gin.Context, queries *db.Queries, channelID, userID uuid.UUID) bool {
	channel, err := queries.GetChannelByID(c.Request.Context(), channelID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "channel not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "channel lookup failed"})
		}
		return false
	}

	if channel.Kind == "public" {
		return true
	}

	isMember, err := queries.IsChannelMember(c.Request.Context(), db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    userID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "membership check failed"})
		return false
	}
	if !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this channel"})
		return false
	}
	return true
}
