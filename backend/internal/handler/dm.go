package handler

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
	"github.com/vidarnilsson/vinlaro-chat/internal/service"
)

type DMHandler struct {
	queries *db.Queries
	rawDB   *sql.DB
}

func NewDMHandler(queries *db.Queries, rawDB *sql.DB) *DMHandler {
	return &DMHandler{queries: queries, rawDB: rawDB}
}

// GetOrCreateDM returns or creates a DM channel between the current user and target.
// POST /api/dm/:userID
func (h *DMHandler) GetOrCreateDM(c *gin.Context) {
	currentUserID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	targetUserID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	if currentUserID == targetUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot DM yourself"})
		return
	}

	channelID, err := service.GetOrCreateDM(c.Request.Context(), h.queries, h.rawDB, currentUserID, targetUserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get or create DM"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"channel_id": channelID})
}

// ListDMs returns all DM conversations for the current user.
// GET /api/dm
func (h *DMHandler) ListDMs(c *gin.Context) {
	userID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	dms, err := h.queries.GetUserDMs(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch DMs"})
		return
	}
	if dms == nil {
		dms = []db.GetUserDMsRow{}
	}

	c.JSON(http.StatusOK, dms)
}
