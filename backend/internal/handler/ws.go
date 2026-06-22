package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/vidarnilsson/vinlaro-chat/internal/auth"
	appws "github.com/vidarnilsson/vinlaro-chat/internal/ws"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Allow all origins for local development; tighten in production.
	CheckOrigin: func(r *http.Request) bool { return true },
}

type WSHandler struct {
	hub       *appws.Hub
	jwtSecret string
}

func NewWSHandler(hub *appws.Hub, jwtSecret string) *WSHandler {
	return &WSHandler{hub: hub, jwtSecret: jwtSecret}
}

// ServeWS upgrades the connection and registers the client with the hub.
// GET /ws/channels/:id?token=<jwt>
func (h *WSHandler) ServeWS(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	// JWT is passed as a query param because the browser WebSocket API
	// does not support custom headers during the upgrade handshake.
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}
	if _, err := auth.ValidateToken(token, h.jwtSecret); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
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
