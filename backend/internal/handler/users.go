package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
)

type UserHandler struct {
	queries *db.Queries
}

func NewUserHandler(queries *db.Queries) *UserHandler {
	return &UserHandler{queries: queries}
}

// SearchUsers returns users matching a username query (excluding current user).
// GET /api/users?search=<query>
func (h *UserHandler) SearchUsers(c *gin.Context) {
	query := c.Query("search")
	if query == "" {
		c.JSON(http.StatusOK, []db.SearchUsersRow{})
		return
	}

	currentUserID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	users, err := h.queries.SearchUsers(c.Request.Context(), db.SearchUsersParams{
		Username: "%" + query + "%",
		ID:       currentUserID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "search failed"})
		return
	}
	if users == nil {
		users = []db.SearchUsersRow{}
	}

	c.JSON(http.StatusOK, users)
}
