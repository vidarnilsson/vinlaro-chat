package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	appws "github.com/vidarnilsson/vinlaro-chat/internal/ws"
	"github.com/vidarnilsson/vinlaro-chat/internal/session"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local development; tighten in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub     *appws.Hub
	queries *db.Queries
}

func NewWSHandler(hub *appws.Hub, queries *db.Queries) *WSHandler {
	return &WSHandler{hub: hub, queries: queries}
}

// ServeWS upgrades the connection and registers the client with the hub.
// The session cookie is sent automatically by the browser during the WS
// handshake, so no token query param is needed.
// GET /ws/channels/:id
func (h *WSHandler) ServeWS(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	sessionID := session.GetFromCookie(c)
	if sessionID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}

	if _, err := h.queries.GetSession(c.Request.Context(), sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "session expired or invalid"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "session lookup failed"})
		}
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := appws.NewClient(h.hub, conn, channelID)
	h.hub.Register(client)

	go client.WritePump()
	go client.ReadPump()
}
