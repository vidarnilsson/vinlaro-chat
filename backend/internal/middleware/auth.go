package middleware

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/session"
)

const UserIDKey = "userID"
const UsernameKey = "username"

// Auth validates the session cookie and populates userID and username in the
// Gin context for downstream handlers.
func Auth(queries *db.Queries) gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := session.GetFromCookie(c)
		if sessionID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			return
		}

		row, err := queries.GetSession(c.Request.Context(), sessionID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalid"})
			} else {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
			}
			return
		}

		c.Set(UserIDKey, row.UserID.String())
		c.Set(UsernameKey, row.Username)
		c.Next()
	}
}
