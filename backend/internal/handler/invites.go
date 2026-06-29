package handler

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/vidarnilsson/vinlaro-chat/internal/db"
	"github.com/vidarnilsson/vinlaro-chat/internal/middleware"
)

type InviteHandler struct {
	queries *db.Queries
}

func NewInviteHandler(queries *db.Queries) *InviteHandler {
	return &InviteHandler{queries: queries}
}

// SendInvite invites a user to a private channel.
// POST /api/channels/:id/invite/:userID
func (h *InviteHandler) SendInvite(c *gin.Context) {
	currentID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	channelID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid channel id"})
		return
	}

	inviteeID, err := uuid.Parse(c.Param("userID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	// Sender must be a member.
	isMember, err := h.queries.IsChannelMember(c.Request.Context(), db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    currentID,
	})
	if err != nil || !isMember {
		c.JSON(http.StatusForbidden, gin.H{"error": "not a member of this channel"})
		return
	}

	// Target must not already be a member.
	alreadyMember, err := h.queries.IsChannelMember(c.Request.Context(), db.IsChannelMemberParams{
		ChannelID: channelID,
		UserID:    inviteeID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "membership check failed"})
		return
	}
	if alreadyMember {
		c.JSON(http.StatusConflict, gin.H{"error": "user is already a member"})
		return
	}

	// Check for existing pending invite.
	_, err = h.queries.GetExistingInvite(c.Request.Context(), db.GetExistingInviteParams{
		ChannelID: channelID,
		InviteeID: inviteeID,
	})
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "invite already pending"})
		return
	}
	if !errors.Is(err, sql.ErrNoRows) {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invite check failed"})
		return
	}

	invite, err := h.queries.CreateChannelInvite(c.Request.Context(), db.CreateChannelInviteParams{
		ChannelID: channelID,
		InviterID: currentID,
		InviteeID: inviteeID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invite"})
		return
	}

	c.JSON(http.StatusCreated, invite)
}

// AcceptInvite accepts a pending channel invite.
// POST /api/invites/:inviteID/accept
func (h *InviteHandler) AcceptInvite(c *gin.Context) {
	currentID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	inviteID, err := uuid.Parse(c.Param("inviteID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	invite, err := h.queries.GetChannelInvite(c.Request.Context(), inviteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		}
		return
	}

	if invite.InviteeID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not the invitee"})
		return
	}

	// Add to channel.
	if err := h.queries.AddChannelMember(c.Request.Context(), db.AddChannelMemberParams{
		ChannelID: invite.ChannelID,
		UserID:    currentID,
	}); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to join channel"})
		return
	}

	updated, err := h.queries.UpdateInviteStatus(c.Request.Context(), db.UpdateInviteStatusParams{
		Status: "accepted",
		ID:     inviteID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update invite"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeclineInvite declines a pending channel invite.
// POST /api/invites/:inviteID/decline
func (h *InviteHandler) DeclineInvite(c *gin.Context) {
	currentID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	inviteID, err := uuid.Parse(c.Param("inviteID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite id"})
		return
	}

	invite, err := h.queries.GetChannelInvite(c.Request.Context(), inviteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "invite not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "lookup failed"})
		}
		return
	}

	if invite.InviteeID != currentID {
		c.JSON(http.StatusForbidden, gin.H{"error": "not the invitee"})
		return
	}

	updated, err := h.queries.UpdateInviteStatus(c.Request.Context(), db.UpdateInviteStatusParams{
		Status: "declined",
		ID:     inviteID,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update invite"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

// ListPendingInvites lists pending channel invites for the current user.
// GET /api/invites
func (h *InviteHandler) ListPendingInvites(c *gin.Context) {
	currentID, err := uuid.Parse(c.GetString(middleware.UserIDKey))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user"})
		return
	}

	invites, err := h.queries.GetPendingInvites(c.Request.Context(), currentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch invites"})
		return
	}
	if invites == nil {
		invites = []db.GetPendingInvitesRow{}
	}

	c.JSON(http.StatusOK, invites)
}
