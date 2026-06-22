package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/messaging"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
	"github.com/vidarnilsson/vinlaro-chat/internal/model"
)

type MessageHandler struct {
	queries  *db.Queries
	producer *messaging.Producer
}

func NewMessageHandler(queries *db.Queries, producer *messaging.Producer) *MessageHandler {
	return &MessageHandler{queries: queries, producer: producer}
}

type sendMessageRequest struct {
	Content string `json:"content" binding:"required,min=1,max=4000"`
}

// SendMessage publishes a message event to Kafka.
// POST /api/channels/:id/messages
func (h *MessageHandler) SendMessage(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDStr := c.GetString(middleware.UserIDKey)
	usernameStr := c.GetString(middleware.UsernameKey)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}
	_ = userID

	event := model.MessageEvent{
		ID:        uuid.New().String(),
		ChannelID: channelID.String(),
		UserID:    userIDStr,
		Username:  usernameStr,
		Content:   req.Content,
		CreatedAt: time.Now().UTC(),
	}

	if err := h.producer.Publish(c.Request.Context(), event); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to queue message"})
		return
	}

	c.JSON(http.StatusAccepted, event)
}

// GetMessages returns paginated message history for a channel.
// GET /api/channels/:id/messages?limit=50
func (h *MessageHandler) GetMessages(c *gin.Context) {
	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	limit := int32(50)
	messages, err := h.queries.GetMessagesByChannel(c.Request.Context(), db.GetMessagesByChannelParams{
		ChannelID: channelID,
		Limit:     limit,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}
	if messages == nil {
		messages = []db.GetMessagesByChannelRow{}
	}

	c.JSON(http.StatusOK, messages)
}
